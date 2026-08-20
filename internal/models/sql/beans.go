package sql

import "time"

// RoastLevel represents how dark the beans were roasted.
//
// 0 = light, 1 = light to medium, 2 = medium, 3 = medium to dark, 4 = dark.
//
// enum: 0,1,2,3,4
type RoastLevel uint8

const (
	RoastLevelLight RoastLevel = iota
	RoastLevelLightToMedium
	RoastLevelMedium
	RoastLevelMediumToDark
	RoastLevelDark
)

// IsValid reports whether r is a supported roast level.
func (r RoastLevel) IsValid() bool {
	return r >= RoastLevelLight && r <= RoastLevelDark
}

// String renders a human label for display. JSON encoding stays numeric;
// this is not used by MarshalJSON.
func (r RoastLevel) String() string {
	switch r {
	case RoastLevelLight:
		return "Light"
	case RoastLevelLightToMedium:
		return "Light to medium"
	case RoastLevelMedium:
		return "Medium"
	case RoastLevelMediumToDark:
		return "Medium to dark"
	case RoastLevelDark:
		return "Dark"
	default:
		return "Unknown"
	}
}

type Beans struct {
	Id         int        `db:"id"`
	Roaster    *Roaster   `db:"roaster"`
	Name       string     `db:"name"`
	RoastDate  *time.Time `db:"roast_date"`
	RoastLevel RoastLevel `db:"roast_level"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}
