package genieacs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// DefaultPollInterval is the gap between GET /tasks/{id} polls when a POSTed
// task returned 202 (queued). Tunable per-Driver via
// target.Extra["poll_interval"] (a time.Duration string like "1s").
const DefaultPollInterval = 500 * time.Millisecond

// defaultFaultChannel is the GenieACS fault channel checked after a task
// disappears from the queue. "default" is the standard channel GenieACS
// assigns unless a preset/task explicitly sets another — overridable per
// Driver via target.Extra["fault_channel"].
const defaultFaultChannel = "default"

// Driver implements port.DeviceDriver for TR-069/ACS via the GenieACS NBI
// (net/http standard library — no extra dependency, per
// TECH-STACK-DAN-PERSIAPAN.md §5 Ringkasan Tegas). Unlike every other driver
// in this package, GenieACS does not hold a persistent connection to the
// managed device: it is a stateless HTTP client to the NBI, which in turn
// talks to the CPE/ONT via TR-069. One Driver instance is bound to ONE CPE
// (deviceID) at NewDriver time — consistent with the DeviceDriver contract
// where Execute carries no device parameter.
//
// NBI has no built-in auth (analisis-api-genieacs.md §5.2) — must be
// network-isolated. An optional x-api-key (community PR #374, not in official
// docs) is supported via target.Extra["api_key"] as defense-in-depth only.
type Driver struct {
	baseURL      url.URL
	deviceID     string
	apiKey       string
	httpClient   *http.Client
	pollInterval time.Duration
	faultChannel string
}

// Compile-time proof that *Driver satisfies port.DeviceDriver — required
// in every vendor driver per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver builds a REST client for the GenieACS NBI bound to one CPE.
//
// Target field mapping (see agreed design — Host/Port = NBI endpoint, CPE
// identity in Extra because device.Target has no dedicated deviceID field):
//   - Host, Port         : NBI endpoint (Port 0 → 7557 default)
//   - Extra["use_tls"]   : "true" → https, anything else → http
//   - Extra["device_id"] : GenieACS device ID of the CPE to manage (REQUIRED)
//   - Extra["api_key"]   : optional x-api-key header value (defense-in-depth)
//   - Extra["poll_interval"] : time.Duration string for 202 poll gap
//   - Extra["fault_channel"] : GenieACS fault channel (default "default")
//   - Timeout            : http.Client.Timeout (0 = rely on ctx only)
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	deviceID := target.Extra["device_id"]
	if deviceID == "" {
		return nil, ErrDeviceIDMissing
	}

	port := target.Port
	if port == 0 {
		port = 7557
	}

	scheme := "http"
	if target.Extra["use_tls"] == "true" {
		scheme = "https"
	}

	pollInterval := DefaultPollInterval
	if raw := target.Extra["poll_interval"]; raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			pollInterval = d
		}
	}

	faultChannel := target.Extra["fault_channel"]
	if faultChannel == "" {
		faultChannel = defaultFaultChannel
	}

	d := &Driver{
		baseURL: url.URL{
			Scheme: scheme,
			Host:   fmt.Sprintf("%s:%d", target.Host, port),
		},
		deviceID:     deviceID,
		apiKey:       target.Extra["api_key"],
		httpClient:   &http.Client{Timeout: target.Timeout},
		pollInterval: pollInterval,
		faultChannel: faultChannel,
	}

	// Verify the NBI is reachable — consistent with every other driver's
	// "NewDriver returns a ready Driver" contract, and surfaces a
	// misconfigured host/port immediately rather than on the first Execute.
	if err := d.ping(ctx); err != nil {
		return nil, fmt.Errorf("genieacs: NBI unreachable: %w", err)
	}
	return d, nil
}

// ping issues a throwaway GET against the devices collection to confirm the
// NBI is up and the baseURL is correct. The query matches nothing
// ("__ping__") so it is cheap and side-effect-free.
func (d *Driver) ping(ctx context.Context) error {
	q := url.Values{}
	q.Set("query", `{"_id":"__ping__"}`)
	resp, err := d.do(ctx, http.MethodGet, "/devices/", q, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NBI responded %d", resp.StatusCode)
	}
	return nil
}

// Execute runs cmd against the bound CPE via the GenieACS NBI. cmd.Raw
// selects the operation:
//   - "getDevice" : direct GET /devices/?query={"_id":...} (cached DB read,
//     no CPE contact — fast path for OpGetStatus)
//   - any task type (getParameterValues, reboot, setParameterValues, ...):
//     POST /devices/{id}/tasks, then handle 200 (done) or 202 (queued →
//     poll GET /tasks/{id} until the task leaves the queue or ctx expires).
//
// For getParameterValues, the CPE's reported values are NOT in the POST
// response (analisis-api-genieacs.md §3) — they land in the device record in
// DB, so Execute follows up with a GET /devices projection to fetch them
// into Result.Rows.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if cmd.Raw == "getDevice" {
		return d.execGetDevice(ctx, cmd.Args["projection"])
	}

	if !isTaskCommand(cmd) {
		return command.Result{}, fmt.Errorf("%w: %q", ErrUnsupportedTaskType, cmd.Raw)
	}

	body, err := buildTaskBody(cmd)
	if err != nil {
		return command.Result{}, err
	}

	_, connectionRequest := cmd.Args["connection_request"]
	taskID, immediate, err := d.postTask(ctx, body, connectionRequest)
	if err != nil {
		return command.Result{}, err
	}

	if !immediate {
		if !d.waitForTask(ctx, taskID) {
			return command.Result{Output: taskID}, fmt.Errorf("%w: task %s", ErrTaskQueued, taskID)
		}
	}

	// Task has left the queue — check whether it faulted.
	if fault, err := d.taskFaultDetail(ctx); err == nil && fault != "" {
		return command.Result{Output: taskID}, fmt.Errorf("%w: %s", ErrTaskFault, fault)
	}

	// Read tasks fetch values from the device record, not the task response.
	if cmd.Raw == "getParameterValues" {
		rows, err := d.fetchParamValues(ctx, cmd.Args["parameterNames"])
		if err != nil {
			return command.Result{Output: taskID}, fmt.Errorf("genieacs: fetch parameter values: %w", err)
		}
		return command.Result{Output: taskID, Rows: rows}, nil
	}

	return command.Result{Output: taskID}, nil
}

// Classify reports the risk class of cmd according to GenieACS task types.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a GenieACS task type Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close is effectively a no-op: the NBI is stateless REST. Idle connections
// are released so a long-lived Driver that is done being used does not hold
// sockets open against the NBI.
func (d *Driver) Close() error {
	d.httpClient.CloseIdleConnections()
	return nil
}

// execGetDevice performs GET /devices/?query={"_id":"<deviceID>"} and returns
// the device record as a single Row. projection (comma-separated param paths)
// is forwarded as the projection query param to limit payload size.
func (d *Driver) execGetDevice(ctx context.Context, projection string) (command.Result, error) {
	q := url.Values{}
	q.Set("query", fmt.Sprintf(`{"_id":"%s"}`, d.deviceID))
	if projection != "" {
		q.Set("projection", projection)
	}

	resp, err := d.do(ctx, http.MethodGet, "/devices/", q, nil)
	if err != nil {
		return command.Result{}, fmt.Errorf("genieacs: get device %s: %w", d.deviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return command.Result{}, fmt.Errorf("genieacs: get device %s: NBI %d", d.deviceID, resp.StatusCode)
	}

	// Device collection returns a JSON array; expect exactly one element.
	var arr []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return command.Result{}, fmt.Errorf("genieacs: decode device response: %w", err)
	}
	if len(arr) == 0 {
		return command.Result{}, fmt.Errorf("genieacs: device %s not found", d.deviceID)
	}

	row := make(map[string]string, len(arr[0]))
	flattenJSON(arr[0], "", row)
	return command.Result{Rows: []map[string]string{row}}, nil
}

// postTask POSTs body to /devices/{deviceID}/tasks and returns the task ID
// plus immediate=true when the NBI reports 200 (CPE executed the task within
// the request window). A 202 means the task is queued — immediate=false,
// caller must poll.
func (d *Driver) postTask(ctx context.Context, body []byte, connectionRequest bool) (string, bool, error) {
	path := "/devices/" + url.PathEscape(d.deviceID) + "/tasks"
	q := url.Values{}
	if connectionRequest {
		q.Set("connection_request", "")
	}
	if deadline, ok := ctx.Deadline(); ok {
		q.Set("timeout", strconv.FormatInt(time.Until(deadline).Milliseconds(), 10))
	}

	resp, err := d.do(ctx, http.MethodPost, path, q, body)
	if err != nil {
		return "", false, fmt.Errorf("genieacs: post task: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
	default:
		return "", false, fmt.Errorf("genieacs: post task: NBI %d", resp.StatusCode)
	}

	var task map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return "", false, fmt.Errorf("genieacs: decode task response: %w", err)
	}

	id, _ := task["_id"].(string)
	if id == "" {
		return "", false, fmt.Errorf("genieacs: task response missing _id")
	}
	return id, resp.StatusCode == http.StatusOK, nil
}

// waitForTask polls GET /tasks/?query={"_id":"<taskID>"} at d.pollInterval
// until the task leaves the queue (array empty = executed) or ctx expires.
// Returns true if the task completed within ctx, false on ctx cancellation.
// Transient request errors are swallowed and retried — only ctx ends the
// loop, matching the blocking-Execute contract of port.DeviceDriver.
func (d *Driver) waitForTask(ctx context.Context, taskID string) bool {
	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			gone, err := d.taskGone(ctx, taskID)
			if err != nil || !gone {
				continue
			}
			return true
		}
	}
}

// taskGone reports whether the task has left the NBI queue. A task that no
// longer appears in /tasks has been dispatched to the CPE — success or fault
// is then distinguished by checking /faults.
func (d *Driver) taskGone(ctx context.Context, taskID string) (bool, error) {
	q := url.Values{}
	q.Set("query", fmt.Sprintf(`{"_id":"%s"}`, taskID))

	resp, err := d.do(ctx, http.MethodGet, "/tasks/", q, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("NBI %d", resp.StatusCode)
	}

	var arr []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return false, err
	}
	return len(arr) == 0, nil
}

// taskFaultDetail returns the fault message for this Driver's device on its
// fault channel, or "" if no fault exists. Called after a task leaves the
// queue to distinguish success from failure. fault_id format is
// "<device_id>:<channel>" (analisis-api-genieacs.md §2.4).
func (d *Driver) taskFaultDetail(ctx context.Context) (string, error) {
	faultID := d.deviceID + ":" + d.faultChannel
	q := url.Values{}
	q.Set("query", fmt.Sprintf(`{"_id":"%s"}`, faultID))

	resp, err := d.do(ctx, http.MethodGet, "/faults/", q, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NBI %d", resp.StatusCode)
	}

	var arr []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return "", err
	}
	if len(arr) == 0 {
		return "", nil
	}
	if msg, _ := arr[0]["message"].(string); msg != "" {
		return msg, nil
	}
	return "fault (no detail)", nil
}

// fetchParamValues reads the requested parameter paths from the device
// record. parameterNamesJSON is the same JSON array string carried in
// cmd.Args["parameterNames"] — reused here as a projection to avoid pulling
// the entire device tree.
func (d *Driver) fetchParamValues(ctx context.Context, parameterNamesJSON string) ([]map[string]string, error) {
	projection := projectionFromParamNames(parameterNamesJSON)
	result, err := d.execGetDevice(ctx, projection)
	if err != nil {
		return nil, err
	}
	return result.Rows, nil
}

// do builds and sends one NBI request with this Driver's base URL, device
// auth header, and ctx. body is nil for GET. The caller closes the returned
// response body.
func (d *Driver) do(ctx context.Context, method, path string, query url.Values, body []byte) (*http.Response, error) {
	u := d.baseURL.JoinPath(path)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if d.apiKey != "" {
		req.Header.Set("x-api-key", d.apiKey)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	return resp, nil
}

// projectionFromParamNames turns a JSON array of TR-069 parameter paths into
// the comma-separated projection string the NBI expects. On parse failure it
// returns "" (no projection = full device tree) rather than erroring — a
// failed projection should not block a read that can still succeed.
func projectionFromParamNames(parameterNamesJSON string) string {
	if parameterNamesJSON == "" {
		return ""
	}
	var names []string
	if err := json.Unmarshal([]byte(parameterNamesJSON), &names); err != nil {
		return ""
	}
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ","
		}
		out += n
	}
	return out
}

// flattenJSON walks a GenieACS device document (map[string]any — the JSON
// deserialization boundary, exempt from CLAUDE.md §6's any-restriction) and
// writes every scalar leaf to out with dotted-path keys. Nested objects
// recurse; arrays stringify. This is deliberately generic rather than
// aware of GenieACS's "_value"/"_type" leaf convention so it tolerates
// document-shape variation across NBI versions.
func flattenJSON(m map[string]any, prefix string, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flattenJSON(val, key, out)
		default:
			out[key] = fmt.Sprint(val)
		}
	}
}
