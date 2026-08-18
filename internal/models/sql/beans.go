package sql

import "time"

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

type Beans struct {
	Id         int        `db:"id"`
	Roaster    *Roaster   `db:"roaster"`
	Name       string     `db:"name"`
	RoastDate  *time.Time `db:"roast_date"`
	RoastLevel RoastLevel `db:"roast_level"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}
