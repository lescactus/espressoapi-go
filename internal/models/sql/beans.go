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
	Id          int
	RoasterName string
	BeansName   string
	RoastDate   time.Time
	RoastLevel  RoastLevel
}
