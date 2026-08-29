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
)

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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
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
