package sql

import "time"

type ComparisonWithPreviousResult uint8

const (
	Worst ComparisonWithPreviousResult = iota
	Same
	Better
	Unknown
)

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
