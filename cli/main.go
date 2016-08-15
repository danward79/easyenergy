package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/danward79/easyenergy/easyapi"
)

func main() {

	startDate := flag.String("s", "", "Start date to retreive data from (dd/mm/yyyy)")
	flag.Parse()

	firstDayToQuery := easyapi.DaysSinceDate(*startDate)

	c := easyapi.NewClient(os.Getenv("ENERGYEASYUSER"), os.Getenv("ENERGYEASYPASS"))

	err := c.GetCookie()
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login()
	if err != nil {
		log.Fatal(err)
	}

	for i := firstDayToQuery; i >= 0; i-- {

		r, err := c.QueryDay(i)
		if err != nil {
			fmt.Println("Error", err)
		}

		c.UpsertNet(&r)
		c.UpsertConsumption(&r)
	}

	err = c.Logout()
	if err != nil {
		log.Fatal(err)
	}
}
