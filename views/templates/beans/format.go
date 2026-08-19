package beans

import "time"

// dateOnly renders t as "YYYY-MM-DD" UTC, or "" if t is nil.
func dateOnly(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}
