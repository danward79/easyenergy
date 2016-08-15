package easyapi

// ConsumptionData ...
type ConsumptionData struct {
	Day []Readings `json:"peak"`
}

// CostData ...
type CostData struct {
	Day []Readings `json:"peak"`
}

// Readings ...
type Readings struct {
	Reading float64 `json:"total"`
}

//QueryResult ...
type QueryResult struct {
	SelectedPeriod struct {
		ConsumptionData ConsumptionData `json:"consumptionData" bson:"consumptionData"`
		CostData        CostData        `json:"costData" bson:"costData"`
		NetConsumption  float64         `json:"netConsumption" bson:"netConsumption"`
		Date            string          `json:"subtitle" bson:"date"`
	} `json:"selectedPeriod" bson:",inline"`
	ComparisonPeriod struct {
		ConsumptionData ConsumptionData `json:"consumptionData" bson:"consumptionData"`
		CostData        CostData        `json:"costData" bson:"costData"`
		NetConsumption  float64         `json:"netConsumption" bson:"netConsumption"`
		Date            string          `json:"subtitle" bson:"date"`
	} `json:"comparisonPeriod" bson:",inline"`
	LatestInterval string `json:"latestInterval"`
}
