package shot

import (
	"math"
	"time"
)

// MaxShotTime is the maximum shot duration accepted by CreateShot and
// UpdateShotById (one hour); it also guards against a legacy nanosecond-
// denominated value being silently accepted as a duration spanning
// centuries. Zero means "not recorded" and is always valid.
const MaxShotTime = 3600 * time.Second

// SecondsToDuration converts a floating-point seconds value (as submitted by
// the REST and web boundaries) to a time.Duration rounded to the nearest
// millisecond, the shots table's storage precision, so a value round-trips
// exactly regardless of which boundary it entered through.
func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(math.Round(seconds*1000)) * time.Millisecond
}
