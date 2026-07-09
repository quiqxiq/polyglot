package genieacs

import "errors"

// ErrDeviceIDMissing indicates NewDriver was called without
// target.Extra["device_id"] — the GenieACS device ID of the CPE/ONT to
// manage. Without it the NBI has no way to address a device, so construction
// must fail rather than silently produce a Driver that errors on every call.
var ErrDeviceIDMissing = errors.New("genieacs: target.Extra[\"device_id\"] is required")

// ErrUnsupportedTaskType indicates Execute was called with a cmd.Raw that is
// neither a recognized GenieACS task type nor a direct-GET operation (such as
// "getDevice"). The driver refuses to forward unknown names to the NBI
// rather than letting the NBI reject them — fail-safe at the boundary.
var ErrUnsupportedTaskType = errors.New("genieacs: unsupported task type")

// ErrTaskQueued indicates a POSTed task returned 202 (queued, not yet
// executed) and the caller's context expired before the task reached a
// terminal state. The task itself is NOT cancelled — it remains in GenieACS's
// queue and will run on the CPE's next inform. Result.Output carries the
// task ID so the caller can poll GET /tasks/{id} or retry deliberately.
var ErrTaskQueued = errors.New("genieacs: task queued but not completed before context deadline")

// ErrTaskFault indicates a task reached the "fault" state — the CPE rejected
// or failed the operation. The fault detail (as reported by GenieACS) is
// attached to the error message.
var ErrTaskFault = errors.New("genieacs: task fault")
