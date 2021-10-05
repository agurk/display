package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Power struct {
	Db *sql.DB
}

type Useage struct {
	Amount     string
	Date       string
	Cost       string
	Efficiency string
}

func NewPower(path string) *Power {
	power := new(Power)
	db, err := sql.Open("sqlite3", path)
	power.Db = db
	if err != nil {
		log.Fatal(err)
	}
	return power
}

// Cost returns the cost of electricity in a certain time period
// exact will from the start of the hour until the time given
// otherwise it'll be between exact hours
func (power *Power) Cost(t time.Time, exact bool) float64 {
	lowerBound := t.Format("2006-01-02 15:00:00")
	var upperBound string
	if exact {
		upperBound = t.Format("2006-01-02 15:04:05")
	} else {
		// Hour is as costs are given in 1 hour slots
		upperBound = t.Add(time.Hour).Format("2006-01-02 15:00:00")
	}
	query := "select price, valid_from from prices where valid_from >= $1 and valid_from < $2"
	rows, err := power.Db.Query(query, lowerBound, upperBound)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		var price, validFrom string
		err = rows.Scan(&price, &validFrom)
		if err != nil {
			log.Fatal(err)
		}
		return fmtPrice(price, validFrom)
	}
	fmt.Println("Missing cost for", t)
	return 0
}

// CurrentCost returns the electrical cost right now
func (power *Power) CurrentCost() int {
	return int(math.Round(power.Cost(time.Now(), true)))
}

// CostData returns the latest two days worth of hourly pricing data
func (power *Power) CostData() (prices []int, currentPos int) {
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02 00:00:00")
	tomorrow := time.Now().Add(48 * time.Hour).Format("2006-01-02 00:00:00")
	query := "select price, valid_from from prices where valid_from >= $1 and valid_from < $2"
	rows, err := power.Db.Query(query, yesterday, tomorrow)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var price, validFrom string
		err = rows.Scan(&price, &validFrom)
		if err != nil {
			log.Fatal(err)
		}
		prices = append(prices, int(math.Round(fmtPrice(price, validFrom))))
	}

	pos, err := strconv.Atoi(time.Now().Format("15"))
	if err != nil {
		log.Fatal(err)
	}

	// if tomorrow's data is available, remove yesterday's
	if len(prices) > 48 {
		dif := len(prices) - 48
		return prices[dif:], pos
	}
	return prices, pos + 24
}

func (power *Power) mostRecentDay() time.Time {
	query := `select
				end
			  from
				useage
			  where
				amount is not '0'
			  order by
				start desc
			  limit
				1`
	rows, err := power.Db.Query(query)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var date string
		err = rows.Scan(&date)
		if err != nil {
			log.Fatal(err)
		}

		// iso format date in first 10 chars
		t, err := time.Parse("2006-01-02", date[:10])
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Set this as config
		loc, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			log.Fatal(err)
		}
		t = t.In(loc)
		_, offset := t.Zone()
		t = t.Add(-1 * time.Second * time.Duration(offset))
		return t
	}
	log.Fatal("No latest date available")
	return time.Now()
}

func (power *Power) powerData(offset, days int) (usage Useage) {
	latest := power.mostRecentDay()

	t2 := latest.Add(time.Hour * 24)
	t2 = t2.AddDate(0, 0, -1*offset)
	_, t2off := t2.Zone()
	t2 = t2.Add(-1 * time.Duration(t2off) * time.Second)
	endOfDay := t2.Format("2006-01-02T15:00:00.000Z")

	t2 = latest.AddDate(0, 0, -(offset + days - 1))
	_, t2off = t2.Zone()
	t2 = t2.Add(-1 * time.Duration(t2off) * time.Second)
	startOfDay := t2.Format("2006-01-02T15:00:00.000Z")

	amt, cost, lowestRate := 0.0, 0.0, 0.0
	query := `select
				amount, start, end
			  from
				useage
			  where
				start >= $1
				and end <= $2`
	rows, err := power.Db.Query(query, startOfDay, endOfDay)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var a, start, end string
		err = rows.Scan(&a, &start, &end)
		if err != nil {
			log.Fatal(err)
		}
		a2, err := strconv.ParseFloat(a, 64)
		if err != nil {
			log.Fatal(err)
		}
		amt += a2
		t, err := time.Parse("2006-01-02T15:04:05.000Z", start)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Set this as config
		loc, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			log.Fatal(err)
		}
		t = t.In(loc)
		rate := power.Cost(t, false)
		cost += a2 * rate
		if rate < lowestRate || lowestRate == 0.0 {
			lowestRate = rate
		}
		usage.Date = start[:10]
	}
	usage.Amount = fmt.Sprintf("%0.2f", amt)
	cheapest := amt * lowestRate
	usage.Efficiency = fmt.Sprintf("%0.1f", cost/cheapest*100)
	usage.Cost = fmt.Sprintf("%0.2f", cost/100)
	return
}

// DayUseage returns the total amount of electricity consumed for the most recent
// day that has data
func (power *Power) DayUseage() Useage {
	return power.powerData(0, 1)
}

// PrevDayUseage returns the amount of electricity consumed for the second most recent
// day that has data
func (power *Power) PrevDayUseage() Useage {
	return power.powerData(1, 1)
}

// WeekUseage returns the amount of power consumed in the last 7 days
func (power *Power) WeekUseage() Useage {
	return power.powerData(0, 7)
}

func fmtPrice(price, date string) float64 {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Fatal(err)
	}

	d, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		log.Fatal(err)
	}

	// Distribution costs
	switch d.Month() {
	case 1, 2, 3, 10, 11, 12:
		switch d.Hour() {
		case 17, 18, 19:
			p += 205.37
			//p += 211.28
		default:
			p += 156.07
			//p += 162
		}
	default:
		p += 162
	}
	return p
}
