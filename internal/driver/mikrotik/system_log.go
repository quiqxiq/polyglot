package mikrotik

import (
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// LogEntry represents one row returned by /log/print.
type LogEntry struct {
	RosID   string
	Time    string
	Topics  string
	Message string
}

// NewPrintLogsCommand builds the command.Command for /log/print.
func NewPrintLogsCommand(topicsFilter string) command.Command {
	args := map[string]string{}
	if topicsFilter != "" {
		args["?topics~"] = topicsFilter
	}
	return command.Command{
		Raw:  "/log/print",
		Args: args,
	}
}

// NewPrintLogCommand is an alias for NewPrintLogsCommand.
func NewPrintLogCommand(topicsFilter string) command.Command {
	return NewPrintLogsCommand(topicsFilter)
}

// NewStreamLogsCommand builds the command.Command for streaming /log/print
// follow, which makes RouterOS push each new log line as it is written.
// topicsFilter is optional (?topics~=).
func NewStreamLogsCommand(topicsFilter string) command.Command {
	args := map[string]string{"follow": ""}
	if topicsFilter != "" {
		args["?topics~"] = topicsFilter
	}
	return command.Command{
		Raw:  "/log/print",
		Args: args,
	}
}

// ParseLogs converts command.Result rows from /log/print into a slice of LogEntry.
// It accepts entries even if .id is omitted by RouterOS event stream sentences.
func ParseLogs(result command.Result) []LogEntry {
	logs := make([]LogEntry, 0, len(result.Rows))
	for i, row := range result.Rows {
		msg := row["message"]
		id := row[".id"]
		timeStr := row["time"]
		topics := row["topics"]

		// Ignore completely empty sentences
		if msg == "" && timeStr == "" && topics == "" {
			continue
		}

		if id == "" {
			id = fmt.Sprintf("log-%d-%d", time.Now().UnixNano(), i)
		}

		logs = append(logs, LogEntry{
			RosID:   id,
			Time:    timeStr,
			Topics:  topics,
			Message: msg,
		})
	}
	return logs
}

// ParseLogEntries is an alias for ParseLogs.
func ParseLogEntries(result command.Result) []LogEntry {
	return ParseLogs(result)
}
