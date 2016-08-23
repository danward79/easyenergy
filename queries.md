### SQL queries

SELECT date,hour,consumption, cost FROM usage
	WHERE (cost = 0 OR consumption = 0)
	AND date > CURRENT_DATE - INTERVAL '2 months'
	ORDER BY date desc, hour desc LIMIT 24;
