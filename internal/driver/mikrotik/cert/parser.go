package cert

// ParseCertificates converts reply rows into Certificate slice.
func ParseCertificates(rows []map[string]string) []Certificate {
	out := make([]Certificate, 0, len(rows))
	for _, r := range rows {
		out = append(out, Certificate{
			ID:            r[".id"],
			Name:          r["name"],
			CommonName:    r["common-name"],
			Fingerprint:   r["fingerprint"],
			InvalidBefore: r["invalid-before"],
			InvalidAfter:  r["invalid-after"],
			ExpiresAfter:  r["expires-after"],
			CA:            r["ca"] == "true",
			Trusted:       r["trusted"] == "true",
			Issued:        r["issued"] == "true",
		})
	}
	return out
}
