package netwatch

// ParseNetwatchHosts converts reply rows into NetwatchHost slice.
func ParseNetwatchHosts(rows []map[string]string) []NetwatchHost {
	out := make([]NetwatchHost, 0, len(rows))
	for _, r := range rows {
		out = append(out, NetwatchHost{
			ID:         r[".id"],
			Host:       r["host"],
			Interval:   r["interval"],
			Timeout:    r["timeout"],
			Status:     r["status"],
			Since:      r["since"],
			UpScript:   r["up-script"],
			DownScript: r["down-script"],
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}
