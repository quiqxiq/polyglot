package ping

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// ParsePingLatency parses ping output row map from RouterOS or raw ping commands.
// It extracts latency in milliseconds and the status string.
func ParsePingLatency(row map[string]string) (int64, string) {
	status := row["status"]
	if status == "timeout" || status == "host unreachable" || status == "net unreachable" {
		return 0, status
	}
	if status == "" {
		status = "connected"
	}

	timeStr := row["time"]
	if timeStr == "" {
		timeStr = row["avg-rtt"]
	}
	if timeStr == "" {
		timeStr = row["min-rtt"]
	}
	if timeStr == "" {
		timeStr = row["rtt"]
	}
	if timeStr == "" {
		timeStr = row["response-time"]
	}

	if timeStr == "" {
		if _, hasSeq := row["seq"]; hasSeq {
			return 1, status
		}
		if _, hasHost := row["host"]; hasHost {
			return 1, status
		}
		return 0, status
	}

	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.TrimPrefix(timeStr, "<")
	timeStr = strings.TrimPrefix(timeStr, ">")

	// Try Go stdlib time.ParseDuration (handles "15ms", "1ms200us", "230us", etc.)
	if d, err := time.ParseDuration(timeStr); err == nil {
		ms := float64(d) / float64(time.Millisecond)
		if ms > 0 && ms < 1.0 {
			return 1, status
		}
		return int64(math.Round(ms)), status
	}

	// Check HH:MM:SS.microsecond format (e.g. 00:00:00.023000)
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 3 {
			secStr := parts[2]
			if sec, err := strconv.ParseFloat(secStr, 64); err == nil {
				ms := sec * 1000.0
				if ms > 0 && ms < 1.0 {
					return 1, status
				}
				return int64(math.Round(ms)), status
			}
		}
	}

	cleanStr := strings.TrimSuffix(timeStr, "ms")
	cleanStr = strings.TrimSuffix(cleanStr, "s")
	cleanStr = strings.TrimSpace(cleanStr)

	if f, err := strconv.ParseFloat(cleanStr, 64); err == nil {
		if f > 0 && f < 1.0 {
			return 1, status
		}
		return int64(math.Round(f)), status
	}

	return 1, status
}
