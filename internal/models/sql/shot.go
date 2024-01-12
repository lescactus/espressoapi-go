package sql

import "time"

type Shot struct {
	Id               int
	GrindSetting     int
	QuantityIn       float64
	QuantityOut      float64
	ShotTime         time.Duration
	WaterTemperature int
}
