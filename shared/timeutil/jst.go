package timeutil

import (
	"time"
	_ "time/tzdata"
)

const JSTName = "Asia/Tokyo"

func JST() *time.Location {
	location, err := time.LoadLocation(JSTName)
	if err != nil {
		return time.FixedZone("JST", 9*60*60)
	}
	return location
}
