package mikrotik

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// PingResult holds one "!re" row from a /ping command.
type PingResult struct {
	Seq        int
	Host       string
	Sent       int
	Received   int
	PacketLoss int
	MinRTT     string
	AvgRTT     string
	MaxRTT     string
	TTL        string
	Time       string
}

// NewPingCommand builds the command.Command for /ping.
func NewPingCommand(host, count string) command.Command {
	if count == "" {
		count = "4"
	}
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address": host,
			"count":   count,
		},
	}
}

// NewPingStreamCommand builds the command.Command for continuous /ping streaming.
func NewPingStreamCommand(host string) command.Command {
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address":  host,
			"interval": "1s",
		},
	}
}

// ParsePingResults converts command.Result rows from /ping into a slice of PingResult.
func ParsePingResults(result command.Result) []PingResult {
	pings := make([]PingResult, 0, len(result.Rows))
	for _, row := range result.Rows {
		seq, _ := strconv.Atoi(row["seq"])
		sent, _ := strconv.Atoi(row["sent"])
		recv, _ := strconv.Atoi(row["received"])
		loss, _ := strconv.Atoi(row["packet-loss"])

		pings = append(pings, PingResult{
			Seq:        seq,
			Host:       row["host"],
			Sent:       sent,
			Received:   recv,
			PacketLoss: loss,
			MinRTT:     row["min-rtt"],
			AvgRTT:     row["avg-rtt"],
			MaxRTT:     row["max-rtt"],
			TTL:        row["ttl"],
			Time:       row["time"],
		})
	}
	return pings
}
