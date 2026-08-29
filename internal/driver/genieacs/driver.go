package genieacs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	defer func() { _ = resp.Body.Close() }()
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
