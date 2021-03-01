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
	Amount string
	Date   string
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

func (power *Power) CurrentCost() int {
	today := time.Now()
	lowerBound := today.Format("2006-01-02 15:00:00")
	upperBound := today.Format("2006-01-02 15:04:05")
	query := "select price, valid_from from prices where valid_from >= $1 and valid_from < $2"
	rows, err := power.Db.Query(query, lowerBound, upperBound)
	rows.Next()
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	var price, validFrom string
	err = rows.Scan(&price, &validFrom)
	if err != nil {
		log.Fatal(err)
	}
	return fmtPrice(price, validFrom)
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
		prices = append(prices, fmtPrice(price, validFrom))
	}

	pos, err := strconv.Atoi(time.Now().Format("15"))
	if err != nil {
		log.Fatal(err)
	}

	// if tomorrow's data is available, remove yesterday's
	if len(prices) == 72 {
		return prices[24:], pos
	}
	return prices, pos + 24
}

func (power *Power) powerData(limit, offset string) (usage Useage) {
	amt := 0.0
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
		if usage.Date == "" {
			usage.Date = start[:10]
		}
	}
	usage.Amount = fmt.Sprintf("%0.2f", amt)
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

// WeekUseage returns the total amount of electricity consumed for the most
// recent complete week
//func (power *Power) WeekUseage() (amount, week string) {
//}

func fmtPrice(price, date string) int {
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
	return int(math.Round(p))
}
