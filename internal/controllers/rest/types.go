package rest

import (
	"encoding/json"
	"strings"
	"time"
)

type RoastDate time.Time

// Implement Marshaler and Unmarshaler interface
func (r *RoastDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*r = RoastDate(t)
	return nil
}

func (r RoastDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(r))
}

// ItemDeletedResponse represents the response when an item is deleted
// swagger:model
type ItemDeletedResponse struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}
