package genieacs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Execute runs cmd against the bound CPE via the GenieACS NBI.
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
