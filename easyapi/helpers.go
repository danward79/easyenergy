package easyapi

import (
	"log"
	"math"
	"time"
)

//Helper funcs...

// DaysSinceDate returns days since a date
func DaysSinceDate(d string) int {
	var date time.Time
	var err error

	if d != "" {
		date, err = time.Parse("02/01/2006", d)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		date = time.Now()
	}

	days := time.Since(date).Hours() / 24

	if days > 0 {
		if math.Mod(days, 1) > 0 {
			days++
		}
	}

	if days > 731 {
		days = 731
	}

	return int(days)
}
