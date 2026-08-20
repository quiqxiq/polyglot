package mikrotik

import (
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
func ParseLogs(result command.Result) []LogEntry {
	logs := make([]LogEntry, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		logs = append(logs, LogEntry{
			RosID:   id,
			Time:    row["time"],
			Topics:  row["topics"],
			Message: row["message"],
		})
	}
	return logs
}

// ParseLogEntries is an alias for ParseLogs.
func ParseLogEntries(result command.Result) []LogEntry {
	return ParseLogs(result)
}
