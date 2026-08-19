package sql

import "time"

type ComparisonWithPreviousResult uint8

const (
	Worst ComparisonWithPreviousResult = iota
	Same
	Better
	Unknown
)

// IsValid reports whether r is a supported comparison result.
func (r ComparisonWithPreviousResult) IsValid() bool {
	return r >= Worst && r <= Unknown
}

// String renders a human label for display. JSON encoding stays numeric;
// this is not used by MarshalJSON.
func (r ComparisonWithPreviousResult) String() string {
	switch r {
	case Worst:
		return "Worse"
	case Same:
		return "Same"
	case Better:
		return "Better"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

type Shot struct {
	Id                           int                          `db:"id"`
	Sheet                        *Sheet                       `db:"sheet"`
	Beans                        *Beans                       `db:"beans"`
	GrindSetting                 int                          `db:"grind_setting"`
	QuantityIn                   float64                      `db:"quantity_in"`
	QuantityOut                  float64                      `db:"quantity_out"`
	ShotTime                     time.Duration                `db:"shot_time"`
	WaterTemperature             float64                      `db:"water_temperature"`
	Rating                       float64                      `db:"rating"`
	IsTooBitter                  bool                         `db:"is_too_bitter"`
	IsTooSour                    bool                         `db:"is_too_sour"`
	ComparisonWithPreviousResult ComparisonWithPreviousResult `db:"comparison_with_previous_result"`
	AdditionalNotes              string                       `db:"additional_notes"`
	CreatedAt                    *time.Time                   `db:"created_at"`
	UpdatedAt                    *time.Time                   `db:"updated_at"`
}
