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
		loc, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			fmt.Println(err)
		}
		t = t.In(loc)
		return t
	}
	log.Fatal("No latest date available")
	return time.Now()
}

// todo: deal with daylight savings
func (power *Power) powerData(limit, offset string) (usage Useage) {
	amt, cost, lowestRate := 0.0, 0.0, 0.0
	query := `select
				amount, start, end
			  from
				useage
			  where
				amount is not '0'
			  order by
				start desc
			  limit
				$1
			  offset
				$2`
	rows, err := power.Db.Query(query, limit, offset)
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
			fmt.Println(err)
		}
		t = t.In(loc)
		rate := power.Cost(t, false)
		cost += a2 * rate
		if rate < lowestRate || lowestRate == 0.0 {
			lowestRate = rate
		}
		if usage.Date == "" {
			usage.Date = start[:10]
		}
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
	return power.powerData("24", "0")
}

// PrevDayUseage returns the amount of electricity consumed for the second most recent
// day that has data
func (power *Power) PrevDayUseage() Useage {
	return power.powerData("24", "24")
}

// WeekUseage returns the amount of power consumed in the last 7 days
func (power *Power) WeekUseage() Useage {
	return power.powerData("168", "0")
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
	switch d.Hour() {
	case 17, 18, 19:
		p += 211.28
	default:
		p += 162
	}
	return p
}
