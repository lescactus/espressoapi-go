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

type Beans struct {
	Id         int        `db:"id"`
	Roaster    *Roaster   `db:"roaster"`
	Name       string     `db:"name"`
	RoastDate  *time.Time `db:"roast_date"`
	RoastLevel RoastLevel `db:"roast_level"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}
