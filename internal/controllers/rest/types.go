package rest

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"
)

type RoastDate time.Time

const roastDateLayout = "2006-01-02"

func (r *RoastDate) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}

	t, err := time.Parse(roastDateLayout, value)
	if err != nil {
		return err
	}
	*r = RoastDate(t)
	return nil
}

// Format returns the roast date in its date-only JSON representation.
func (r RoastDate) Format() string {
	return time.Time(r).Format(roastDateLayout)
}

func (r RoastDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Format())
}

// ItemDeletedResponse represents the response when an item is deleted
// swagger:model
type ItemDeletedResponse struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}

// DurationSeconds is the wire representation of a shot duration: a JSON
// number of seconds (25.5 == 25.5s). It stores seconds rounded to the
// nearest millisecond, matching the shots table's storage precision, so a
// value round-trips exactly through Marshal/Unmarshal.
type DurationSeconds float64

// MaxShotTimeSeconds is the maximum plausible shot duration (one hour); it
// also guards against a legacy nanosecond-denominated value being silently
// accepted as a duration spanning centuries.
const MaxShotTimeSeconds = 3600.0

const errShotTimeRange = "shot_time must be a number of seconds greater than 0 and at most 3600"

func (d *DurationSeconds) UnmarshalJSON(b []byte) error {
	var seconds float64
	if err := json.Unmarshal(b, &seconds); err != nil {
		return err
	}
	if seconds <= 0 || seconds > MaxShotTimeSeconds {
		return NewErrorResponse(http.StatusBadRequest, errShotTimeRange)
	}
	*d = DurationSeconds(math.Round(seconds*1000) / 1000)
	return nil
}

func (d DurationSeconds) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(d), 'f', -1, 64)), nil
}

// Duration converts d to a time.Duration for the domain/service layer.
func (d DurationSeconds) Duration() time.Duration {
	return time.Duration(math.Round(float64(d) * float64(time.Second)))
}

// NewDurationSeconds converts a domain time.Duration to its wire
// representation, deriving from the millisecond value (not raw
// nanoseconds) so it matches what UnmarshalJSON would produce for an
// equivalent request.
func NewDurationSeconds(d time.Duration) DurationSeconds {
	return DurationSeconds(float64(d.Milliseconds()) / 1000)
}
