package controllers

import (
	"encoding/json"
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
