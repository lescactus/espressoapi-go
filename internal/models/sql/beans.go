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
	Id          int        `db:"id" json:"id"`
	RoasterName string     `db:"roaster_name" json:"roaster_name"`
	BeansName   string     `db:"beans_name" json:"beans_name"`
	RoastDate   time.Time  `db:"roast_date" json:"roast_date"`
	RoastLevel  RoastLevel `db:"roast_level" json:"roast_level"`
}
