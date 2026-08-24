package shots

import (
	"strconv"
	"time"

	"github.com/lescactus/espressoapi-go/internal/services/bean"
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

func dateOnly(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

func beanName(b *bean.Bean) string {
	if b == nil {
		return ""
	}
	return b.Name
}

func beanRoasterName(b *bean.Bean) string {
	if b == nil || b.Roaster == nil {
		return ""
	}
	return b.Roaster.Name
}

func beanRoastDate(b *bean.Bean) string {
	if b == nil {
		return ""
	}
	return dateOnly(b.RoastDate)
}

func beanOptionLabel(b bean.Bean) string {
	label := b.Name
	if b.Roaster != nil && b.Roaster.Name != "" {
		label += " — " + b.Roaster.Name
	}
	if b.RoastDate != nil {
		label += " (" + dateOnly(b.RoastDate) + ")"
	}
	return label
}
