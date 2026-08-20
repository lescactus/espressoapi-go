package rest

import (
	"encoding/json"
	"math"
	"time"

	"github.com/lescactus/espressoapi-go/internal/services/shot"
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
// value round-trips exactly through Marshal/Unmarshal. Range validation
// (0 <= seconds <= 3600) happens once, in the service layer, so it applies
// identically regardless of which boundary (REST or web) a value came from.
type DurationSeconds float64

// UnmarshalJSON parses a JSON number into rounded seconds. A JSON null is
// treated as a no-op, matching encoding/json's convention for an omitted
// value (both leave the field at its zero value).
func (d *DurationSeconds) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var seconds float64
	if err := json.Unmarshal(b, &seconds); err != nil {
		return err
	}
	*d = DurationSeconds(math.Round(seconds*1000) / 1000)
	return nil
}

// Duration converts d to a time.Duration for the domain/service layer.
func (d DurationSeconds) Duration() time.Duration {
	return shot.SecondsToDuration(float64(d))
}

// NewDurationSeconds converts a domain time.Duration to its wire
// representation, deriving from the millisecond value (not raw
// nanoseconds) so it matches what UnmarshalJSON would produce for an
// equivalent request.
func NewDurationSeconds(d time.Duration) DurationSeconds {
	return DurationSeconds(float64(d.Milliseconds()) / 1000)
}
