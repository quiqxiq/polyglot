package network

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// CollectStream collects bounded results from a streaming device driver into a command.Result.
// It is used when an execution pipeline encounters a streaming command (like /ping,
// /interface/monitor-traffic, or print interval=) that requires stream listening rather
// than a blocking one-shot execute.
func CollectStream(ctx context.Context, sd port.StreamingDeviceDriver, cmd command.Command) (command.Result, error) {
	if sd == nil {
		return command.Result{}, command.ErrDriverUnavailable
	}

	// For ping without count, set safe default count=4 to avoid an unbounded stream.
	if cmd.Raw == "/ping" {
		if cmd.Args == nil {
			cmd.Args = make(map[string]string)
		}
		if _, ok := cmd.Args["count"]; !ok {
			cmd.Args["count"] = "4"
		}
	}

	maxRows := 10
	if cStr, ok := cmd.Args["count"]; ok {
		if c, err := strconv.Atoi(cStr); err == nil && c > 0 {
			maxRows = c
		}
	} else if cmd.Raw == "/interface/monitor-traffic" {
		maxRows = 3
	} else if strings.HasSuffix(cmd.Raw, "/print") {
		maxRows = 10
	}

	// Bound maximum collection duration: (maxRows + 2) seconds, between 5s and 15s.
	timeoutSec := maxRows + 2
	if timeoutSec < 5 {
		timeoutSec = 5
	}
	if timeoutSec > 15 {
		timeoutSec = 15
	}
	timeout := time.Duration(timeoutSec) * time.Second

	if deadline, ok := ctx.Deadline(); ok {
		if rem := time.Until(deadline); rem > 0 && rem < timeout {
			timeout = rem
		}
	}

	streamCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	handle, err := sd.Stream(streamCtx, cmd)
	if err != nil {
		return command.Result{}, fmt.Errorf("stream %q: %w", cmd.Raw, err)
	}
	defer func() { _ = handle.Cancel() }()

	var allRows []map[string]string
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := false
	for !done {
		select {
		case <-streamCtx.Done():
			done = true
		case <-ticker.C:
			done = true
		case res, ok := <-handle.Chan():
			if !ok {
				done = true
				break
			}
			if len(res.Rows) > 0 {
				allRows = append(allRows, res.Rows...)
				if len(allRows) >= maxRows {
					done = true
				}
			}
		}
	}

	if len(allRows) == 0 && handle.Err() != nil {
		return command.Result{}, fmt.Errorf("stream error: %w", handle.Err())
	}

	return command.Result{
		Output: formatStreamOutput(cmd, allRows),
		Rows:   allRows,
	}, nil
}

func formatStreamOutput(cmd command.Command, rows []map[string]string) string {
	if len(rows) == 0 {
		return "ok (0 rows)"
	}

	var sb strings.Builder
	for i, row := range rows {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Ping output formatting
		if host, ok := row["host"]; ok && host != "" {
			seq := row["seq"]
			if seq == "" {
				seq = row["sequence"]
			}
			rtt := row["time"]
			ttl := row["ttl"]
			status := row["status"]
			loss := row["packet-loss"]
			if status != "" && status != "ok" {
				fmt.Fprintf(&sb, "host=%s status=%s", host, status)
			} else if loss != "" && row["sent"] != "" {
				fmt.Fprintf(&sb, "sent=%s received=%s packet-loss=%s min-rtt=%s avg-rtt=%s max-rtt=%s",
					row["sent"], row["received"], loss, row["min-rtt"], row["avg-rtt"], row["max-rtt"])
			} else {
				fmt.Fprintf(&sb, "seq=%s host=%s time=%s ttl=%s", seq, host, rtt, ttl)
			}
			continue
		}

		// Generic key-value formatting
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, row[k]))
		}
		sb.WriteString(strings.Join(parts, " "))
	}
	return sb.String()
}
