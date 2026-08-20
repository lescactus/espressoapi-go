package shots

import (
	"strconv"
	"time"
)

// secondsString renders a shot_time duration as seconds with one decimal,
// e.g. "28.5".
func secondsString(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 1, 64)
}

// secondsDisplay renders a shot_time duration for display, e.g. "28.5 s".
func secondsDisplay(d time.Duration) string {
	return secondsString(d) + " s"
}
