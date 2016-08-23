package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/danward79/easyenergy/easyapi"
)

var interval *int

type timeStrings []time.Time

func (t timeStrings) Len() int           { return len(t) }
func (t timeStrings) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t timeStrings) Less(i, j int) bool { return t[i].Before(t[j]) }

func init() {
	interval = flag.Int("interval", 60, "interval beween checks in minutes for new data")
	flag.Parse()
}

func main() {
	log.Println("Energy Easy Started ....")
	// Create new instance of webservice
	c := easyapi.NewClient(os.Getenv("ENERGYEASYUSER"), os.Getenv("ENERGYEASYPASS"), os.Getenv("ENERGYEASYDBUSER"), os.Getenv("ENERGYEASYDBPASS"), os.Getenv("ENERGYEASYDBPATH"), os.Getenv("ENERGYEASYDB"))

	// Main loop... Checks every 60 mins for a new dataset
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-time.After(time.Duration(*interval) * time.Minute):
			doQuery(c)
		case <-signalChan:
			exit()
		}
	}
}

func exit() {
	log.Println("Interrupt Signal received... Exiting")
	os.Exit(0)
}

func doQuery(c *easyapi.EasyClient) {

	//log.Println("EasyEnergy Query") // NOTE: Unnecessary logging!

	// Look for days with empty data and then sort to exract the lowest
	missing := c.FindMissingData()
	sort.Sort(timeStrings(missing))

	firstDayToQuery := easyapi.DaysSinceDate(missing[0].Format("02/01/2006"))

	if err := c.GetCookie(); err != nil {
		log.Println(err)
	}

	if err := c.Login(); err != nil {
		log.Println(err)
	}

	for i := firstDayToQuery; i >= 0; i-- {

		r, err := c.QueryDay(i)
		if err != nil {
			log.Println("Error", err)
		}

		c.UpsertNet(&r)
		c.UpsertConsumption(&r)
	}

	if updates, err := c.PollUpdatesAvailable(); err != nil {
		log.Println("New data available:", updates)
		log.Println(err)
	}

	if err := c.Logout(); err != nil {
		log.Println(err)
	}

}
