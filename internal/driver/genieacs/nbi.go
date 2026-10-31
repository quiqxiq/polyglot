package genieacs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidNBIURL indicates NewNBI was given a URL that is not an absolute
// http/https URL. The NBI is only ever reached over those two schemes (TLS is
// terminated by a reverse proxy when used), so anything else is a
// configuration error rather than something to silently mangle.
var ErrInvalidNBIURL = errors.New("genieacs: invalid NBI URL")

// ErrDeviceNotFound indicates a GenieACS query for one specific device ID
// returned no documents. The CPE is unknown to the ACS (or was deleted), as
// opposed to a transport failure that should be retried.
var ErrDeviceNotFound = errors.New("genieacs: device not found")

// ErrEmptyArgument indicates a required identifier (device ID, preset name,
// tag, fault ID) was blank. Path arguments are rejected before a request is
// built because an empty segment silently targets the wrong NBI resource.
var ErrEmptyArgument = errors.New("genieacs: required argument is empty")

// NBI is a stateless client for the GenieACS northbound interface (NBI) —
// the admin/management REST API for devices, tasks, presets, faults, and
// tags, as opposed to the per-CPE port.DeviceDriver in driver.go.
//
// Unlike Driver, one NBI value is not bound to a single CPE: callers pass the
// device ID per operation, so a single client serves inventory sync, ACS↔ONT
// matching, telemetry reads, and remote actions. It holds no mutable state and
// is safe for concurrent use.
//
// The NBI has no built-in authentication (community builds accept an optional
// x-api-key header) — it must stay on a network-isolated segment.
type NBI struct {
	baseURL    url.URL
	apiKey     string
	httpClient *http.Client
}

// NewNBI builds a GenieACS NBI client for rawURL (for example
// "http://genieacs:7557" or "https://acs.example/genieacs"). The scheme must
// be http or https and the URL must carry a host; a path prefix is preserved
// for deployments behind a reverse proxy, while query and fragment components
// are rejected as configuration mistakes.
//
// apiKey, when non-empty, is sent as the x-api-key header on every request.
// timeout bounds a single HTTP exchange; zero means the caller's context is
// the only deadline.
//
// NewNBI performs no network I/O — call Ping to verify reachability. This
// keeps construction cheap for use cases that construct the client while the
// NAT/port-forward to the ACS is still being established.
func NewNBI(rawURL, apiKey string, timeout time.Duration) (*NBI, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, fmt.Errorf("%w: empty", ErrInvalidNBIURL)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrInvalidNBIURL, rawURL, err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("%w: %q: scheme must be http or https", ErrInvalidNBIURL, rawURL)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("%w: %q: host is required", ErrInvalidNBIURL, rawURL)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("%w: %q: query and fragment are not allowed", ErrInvalidNBIURL, rawURL)
	}

	return &NBI{
		baseURL:    *parsed,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

// NBIListOptions filters and paginates a devices collection query. The zero
// value lists every device with its full parameter tree — expensive on a
// production ACS, so callers should normally set Projection and Limit.
type NBIListOptions struct {
	// Query is the raw MongoDB query JSON accepted by the NBI query
	// parameter (for example the output of DeviceBySerialQuery). Empty means
	// no filter.
	Query string
	// Projection limits the returned parameter paths (the NBI's
	// comma-separated projection parameter). Empty returns full device
	// documents.
	Projection []string
	// Skip is the number of matching documents to skip; negative values are
	// ignored (treated as zero).
	Skip int
	// Limit is the maximum number of documents to return; zero or negative
	// means no server-side limit.
	Limit int
}

// values renders opts as NBI query parameters, omitting zero values so a bare
// request stays bare.
func (o NBIListOptions) values() url.Values {
	q := url.Values{}
	if o.Query != "" {
		q.Set("query", o.Query)
	}
	if projection := formatProjection(o.Projection); projection != "" {
		q.Set("projection", projection)
	}
	if o.Skip > 0 {
		q.Set("skip", strconv.Itoa(o.Skip))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	return q
}

// Ping verifies the NBI is reachable with a cheap side-effect-free device
// query that can never match a real CPE.
func (n *NBI) Ping(ctx context.Context) error {
	q := url.Values{}
	q.Set("query", `{"_id":"__ping__"}`)

	resp, err := n.do(ctx, http.MethodGet, "/devices/", q, nil)
	if err != nil {
		return fmt.Errorf("genieacs: ping NBI: %w", err)
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("genieacs: ping NBI: %w", statusError(resp))
	}
	return nil
}

// ListDevices reads device records from the NBI devices collection. Records
// are normalized through parseDeviceRecord so callers get OUI/ProductClass/
// serial/software version without walking the raw parameter tree themselves.
func (n *NBI) ListDevices(ctx context.Context, opts NBIListOptions) ([]DeviceRecord, error) {
	resp, err := n.do(ctx, http.MethodGet, "/devices/", opts.values(), nil)
	if err != nil {
		return nil, fmt.Errorf("genieacs: list devices: %w", err)
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genieacs: list devices: %w", statusError(resp))
	}

	var docs []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, fmt.Errorf("genieacs: list devices: decode response: %w", err)
	}

	records := make([]DeviceRecord, 0, len(docs))
	for _, doc := range docs {
		record, err := parseDeviceRecord(doc)
		if err != nil {
			return nil, fmt.Errorf("genieacs: list devices: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// GetDevice reads one device record by its GenieACS device ID. projection
// limits the returned parameter paths; empty fetches the full tree.
func (n *NBI) GetDevice(ctx context.Context, deviceID string, projection []string) (DeviceRecord, error) {
	if err := requireArgument("device id", deviceID); err != nil {
		return DeviceRecord{}, err
	}

	records, err := n.ListDevices(ctx, NBIListOptions{
		Query:      deviceIDQuery(deviceID),
		Projection: projection,
		Limit:      1,
	})
	if err != nil {
		return DeviceRecord{}, fmt.Errorf("genieacs: get device %q: %w", deviceID, err)
	}
	if len(records) == 0 {
		return DeviceRecord{}, fmt.Errorf("%w: %q", ErrDeviceNotFound, deviceID)
	}
	return records[0], nil
}

// PostTask enqueues one GenieACS task (getParameterValues, refreshObject,
// setParameterValues, addObject, deleteObject, reboot, factoryReset, or
// download) for a device. body is the task object and is JSON-encoded by this
// method.
//
// immediate reports whether the CPE executed the task within the request: the
// NBI answers 200 when it did and 202 when the task was queued for the next
// inform. For 202, callers may wait with WaitTask and must poll ListFaults to
// distinguish success from a device-side fault.
//
// When connectionRequest is true the NBI attempts a connection request so the
// task runs immediately; this can extend the request up to the caller's
// context deadline.
func (n *NBI) PostTask(ctx context.Context, deviceID string, body any, connectionRequest bool) (taskID string, immediate bool, err error) {
	if err := requireArgument("device id", deviceID); err != nil {
		return "", false, err
	}
	if body == nil {
		return "", false, fmt.Errorf("%w: task body", ErrEmptyArgument)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", false, fmt.Errorf("genieacs: post task to %q: marshal body: %w", deviceID, err)
	}

	q := url.Values{}
	if connectionRequest {
		q.Set("connection_request", "")
	}
	// GenieACS waits up to `timeout` milliseconds for an immediate execution
	// before falling back to queueing; mirror the caller's deadline so the
	// HTTP client and the NBI agree on when to give up.
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline).Milliseconds(); remaining > 0 {
			q.Set("timeout", strconv.FormatInt(remaining, 10))
		}
	}

	path := "/devices/" + url.PathEscape(deviceID) + "/tasks"
	resp, err := n.do(ctx, http.MethodPost, path, q, payload)
	if err != nil {
		return "", false, fmt.Errorf("genieacs: post task to %q: %w", deviceID, err)
	}
	defer closeResponse(resp)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
	default:
		return "", false, fmt.Errorf("genieacs: post task to %q: %w", deviceID, statusError(resp))
	}

	var task struct {
		ID string `json:"_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return "", false, fmt.Errorf("genieacs: post task to %q: decode response: %w", deviceID, err)
	}
	if task.ID == "" {
		return "", false, fmt.Errorf("genieacs: post task to %q: response missing _id", deviceID)
	}
	return task.ID, resp.StatusCode == http.StatusOK, nil
}

// WaitTask polls the NBI task queue until taskID leaves it or ctx expires.
// A task that has left the queue was dispatched to the CPE — success or fault
// is then distinguished with ListFaults, exactly like the per-CPE driver.
//
// poll is the gap between GET requests; zero or negative falls back to
// DefaultPollInterval. Transient request errors are swallowed and retried
// because only ctx should end the wait. Returns false when ctx expires first.
func (n *NBI) WaitTask(ctx context.Context, taskID string, poll time.Duration) (gone bool) {
	if poll <= 0 {
		poll = DefaultPollInterval
	}

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			gone, err := n.taskGone(ctx, taskID)
			if err != nil || !gone {
				continue
			}
			return true
		}
	}
}

// taskGone reports whether the task has left the NBI queue. A task that no
// longer appears in /tasks has been dispatched to the CPE.
func (n *NBI) taskGone(ctx context.Context, taskID string) (bool, error) {
	if err := requireArgument("task id", taskID); err != nil {
		return false, err
	}

	q := url.Values{}
	q.Set("query", deviceIDQuery(taskID))

	resp, err := n.do(ctx, http.MethodGet, "/tasks/", q, nil)
	if err != nil {
		return false, err
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return false, statusError(resp)
	}

	var tasks []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return false, err
	}
	return len(tasks) == 0, nil
}

// do sends one NBI HTTP request with the client's base URL, optional API key,
// and ctx. body is nil for GET/DELETE without a payload. The caller closes the
// returned response body via closeResponse.
func (n *NBI) do(ctx context.Context, method, path string, query url.Values, body []byte) (*http.Response, error) {
	u := n.baseURL.JoinPath(path)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("genieacs: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if n.apiKey != "" {
		req.Header.Set("x-api-key", n.apiKey)
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("genieacs: http: %w", err)
	}
	return resp, nil
}

// doAcknowledged performs a mutating NBI request whose success GenieACS
// reports as an empty 200 response (204 is accepted for forward
// compatibility). body is nil for requests that carry no payload.
func (n *NBI) doAcknowledged(ctx context.Context, method, path string, body []byte) error {
	resp, err := n.do(ctx, method, path, nil, body)
	if err != nil {
		return err
	}
	defer closeResponse(resp)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return statusError(resp)
	}
}

// formatProjection renders parameter paths as the NBI's comma-separated
// projection value. Blank entries are dropped: the NBI treats an empty path as
// a request for the whole tree, which is the opposite of the caller's intent.
func formatProjection(paths []string) string {
	cleaned := make([]string, 0, len(paths))
	for _, path := range paths {
		if path = strings.TrimSpace(path); path != "" {
			cleaned = append(cleaned, path)
		}
	}
	return strings.Join(cleaned, ",")
}

// requireArgument rejects blank identifiers before they reach the NBI.
func requireArgument(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s", ErrEmptyArgument, name)
	}
	return nil
}

// closeResponse releases an NBI response body. The close error is ignored:
// the body has already been consumed (or deliberately discarded) and there is
// no recovery action for a failed close on an HTTP response.
func closeResponse(resp *http.Response) {
	_ = resp.Body.Close()
}

// statusError converts an unexpected NBI HTTP status into an error carrying a
// bounded excerpt of the response body. GenieACS returns short plain-text
// diagnostics such as "No such device"; the excerpt is capped so a proxy's
// HTML error page cannot flood the error chain.
func statusError(resp *http.Response) error {
	excerpt, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	if message := strings.TrimSpace(string(excerpt)); message != "" {
		return fmt.Errorf("NBI status %d: %s", resp.StatusCode, message)
	}
	return fmt.Errorf("NBI status %d", resp.StatusCode)
}
