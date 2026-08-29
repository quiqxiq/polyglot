package rosutil

// SetIfNonEmpty assigns value to args[key] only if value is non-empty.
func SetIfNonEmpty(args map[string]string, key, value string) {
	if value != "" {
		args[key] = value
	}
}
