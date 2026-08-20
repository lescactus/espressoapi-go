package shared

import "time"

// FormatTimestamp renders t as "YYYY-MM-DD HH:MM" UTC, or "" if t is nil.
func FormatTimestamp(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04")
}
