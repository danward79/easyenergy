package easyapi

import (
	"fmt"
	"log"
	"time"
)

// UpsertConsumption Data update or insert new data.
func (c *EasyClient) UpsertConsumption(j *QueryResult) {
	/*
		UPDATE c SET consumption = 999, cost=8
		WHERE (date='2016-08-14' AND hour=23);
		INSERT INTO c (date, hour, consumption, cost)
		SELECT '2016-08-14', 23, 999, 7
		WHERE NOT EXISTS (SELECT * FROM c WHERE (date='2016-08-14' AND hour=23));
	*/
	date := j.SelectedPeriod.Date

	for k, v := range j.SelectedPeriod.ConsumptionData.Day {
		s := prepUpsertConsumptionStatement(date, k, v.Reading, j.SelectedPeriod.CostData.Day[k].Reading)
		err := c.db.Execute(s)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// prepUpsertConsumptionStatement helper to prepare database statement
func prepUpsertConsumptionStatement(date string, hour int, consumption, cost float64) string {
	// HACK: Method used to update or insert is a bit of a hack.
	return fmt.Sprintf(`UPDATE usage SET consumption = %f, cost=%f WHERE (date='%s' AND hour=%d);
	INSERT INTO usage (date, hour, consumption, cost)
	SELECT '%s', %d, %f, %f
	WHERE NOT EXISTS (SELECT * FROM usage WHERE (date='%s' AND hour=%d));`, consumption, cost, date, hour, date, hour, consumption, cost, date, hour)
}

// UpsertNet update or insert net total consumption data
func (c *EasyClient) UpsertNet(j *QueryResult) {
	/*
			UPDATE netconsumption SET consumption = 999
			WHERE date='2016-08-14';
			INSERT INTO netconsumption (date, consumption)
		  SELECT '2016-08-14', 123
		  WHERE NOT EXISTS (SELECT * FROM netconsumption WHERE date='2016-08-14');
	*/
	consumption := j.SelectedPeriod.NetConsumption
	date := j.SelectedPeriod.Date

	statement := fmt.Sprintf(`UPDATE netconsumption SET consumption = %f
	WHERE date='%s';
	INSERT INTO netconsumption (date, consumption)
  SELECT '%s', %f
  WHERE NOT EXISTS (SELECT * FROM netconsumption WHERE date='%s');`, consumption, date, date, consumption, date)

	err := c.db.Execute(statement)
	if err != nil {
		log.Fatal(err)
	}
}

// InsertNet add net consumption data to the db
func (c *EasyClient) InsertNet(j *QueryResult) {

	statement := "INSERT INTO netconsumption(date, consumption) VALUES($1, $2);"

	err := c.db.Execute(statement, j.SelectedPeriod.Date, j.SelectedPeriod.NetConsumption)
	if err != nil {
		log.Fatal(err)
	}
}

// InsertConsumption add net consumption data to the db
func (c *EasyClient) InsertConsumption(j *QueryResult) {

	statement := "INSERT INTO usage(date, hour, consumption, cost) VALUES($1, $2, $3, $4);"

	date := j.SelectedPeriod.Date

	for k, v := range j.SelectedPeriod.ConsumptionData.Day {
		err := c.db.Execute(statement, date, k, v.Reading, j.SelectedPeriod.CostData.Day[k].Reading)
		if err != nil {
			log.Fatal(err)
		}
	}

}

// consumption ...
type consumption struct {
	date        time.Time
	hour        int
	consumption float64
	cost        float64
}

// FindMissingData returns dates with missing data
func (c *EasyClient) FindMissingData() []time.Time {
	/*
	  SELECT date,hour,consumption, cost FROM usage
	  WHERE (cost = 0 OR consumption = 0)
	  AND date > CURRENT_DATE - INTERVAL '2 months'
	  ORDER BY date desc, hour desc LIMIT 24;
	*/

	statement := `SELECT date,hour,consumption, cost FROM usage
	WHERE (cost = 0 OR consumption = 0)
	AND date > CURRENT_DATE - INTERVAL '2 months'
	ORDER BY date desc, hour desc LIMIT 24;`

	rows, err := c.db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var dateWithMissingData []time.Time
	for rows.Next() {
		var c consumption

		if err := rows.Scan(&c.date, &c.hour, &c.consumption, &c.cost); err != nil {
			log.Fatal(err)
		}

		dateWithMissingData = append(dateWithMissingData, c.date)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return dateWithMissingData
}
