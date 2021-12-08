package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Weather struct {
	Latitude  string
	Longitude string
	weather   data
}

type Forecast struct {
	Date                string
	TempMax             string
	TempMin             string
	PrecipitationAmount string
	Weekend             bool
}

type Hour struct {
	Hour int
	// Temperature is 10x degrees C
	Temperature         int
	Sky                 Cover
	Precipitation       Precipitation
	PrecipitationAmount int
}

type Cover int

const (
	Clear Cover = iota
	Broken
	Cloudy
	Fog
)

type Precipitation int

const (
	None Precipitation = iota
	LightRain
	HeavyRain
	LightSleet
	HeavySleet
	LightSnow
	HeavySnow
)

// Begin DMI Data Struct
type data struct {
	Id         string
	City       string
	Country    string
	Longitude  float64
	Latitude   float64
	Timezone   string
	Lastupdate string
	Sunrise    string
	Sunset     string
	Timeserie  []timeserie
	AggData    []aggdata
}

type timeserie struct {
	Time        string
	Temp        float64
	Symbol      int
	Precip1     float64
	PrecipType  string
	WindDir     string
	WindDegree  float64
	WindSpeed   float64
	WindGust    float64
	Humidity    float64
	Pressure    float64
	Visibility  float64
	Precip3     float64
	Precip6     float64
	Temp10      float64
	Temp50      float64
	Temp90      float64
	Prec10      float64
	Prec50      float64
	Prec75      float64
	Prec90      float64
	Windspeed10 float64
	Windspeed50 float64
	Windspeed90 float64
}

type aggdata struct {
	Time        string
	MinTemp     float64
	MeanTemp    float64
	MaxTemp     float64
	PrecipSum   float64
	UvRadiation float64
}

// End DMI Data Struct

func NewWeather(latitude, longitude string) *Weather {
	w := new(Weather)
	w.Longitude = longitude
	w.Latitude = latitude
	err := w.LoadWeather()
	if err != nil {
		log.Fatal(err)
	}
	return w
}

func (w *Weather) LoadWeather() error {
	resp, err := http.Get(w.url())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(&w.weather)
}

func (w *Weather) url() string {
	return "https://www.dmi.dk/NinJo2DmiDk/ninjo2dmidk?cmd=llj&lon=" + w.Longitude +
		"&lat=" + w.Latitude
}

func (w *Weather) Temp() string {
	return fmt.Sprintf("%.1f", w.weather.Timeserie[0].Temp)
}

func (w *Weather) MaxTemp() string {
	return fmt.Sprintf("%.0f", w.weather.AggData[0].MaxTemp)
}

func (w *Weather) MinTemp() string {
	return fmt.Sprintf("%.0f", w.weather.AggData[0].MinTemp)
}

func (w *Weather) Pressure() string {
	return fmt.Sprintf("%.0f", w.weather.Timeserie[0].Pressure)
}

func (w *Weather) WindSpeed() string {
	return fmt.Sprintf("%.1f", w.weather.Timeserie[0].WindSpeed)
}

func (w *Weather) WindDirection() string {
	return w.weather.Timeserie[0].WindDir
}

func (w *Weather) WindGust() string {
	return fmt.Sprintf("%.1f", w.weather.Timeserie[0].WindGust)
}

func (w *Weather) PrecipitationAmount() string {
	return fmt.Sprintf("%.0f", w.weather.Timeserie[0].Precip1)
}

func (w *Weather) DayPrecipitationAmount() string {
	return fmt.Sprintf("%0.1f", w.weather.AggData[0].PrecipSum)
}

func (w *Weather) PrecipitationType() string {
	if w.weather.Timeserie[0].Precip1 < 0.5 {
		return ""
	}
	return w.weather.Timeserie[0].PrecipType
}

func (w *Weather) UV() string {
	return fmt.Sprintf("%0.1f", w.weather.AggData[0].UvRadiation)
}

func (w *Weather) Humidity() string {
	return fmt.Sprintf("%0.0f", w.weather.Timeserie[0].Humidity)
}

func (w *Weather) Visibility() string {
	if w.weather.Timeserie[0].Visibility >= 1500 {
		return fmt.Sprintf("%0.1f", w.weather.Timeserie[0].Visibility/1000)
	}
	return fmt.Sprintf("%0.0f", w.weather.Timeserie[0].Visibility)

}

func (w *Weather) VisibiltyDistance() string {
	if w.weather.Timeserie[0].Visibility >= 1500 {
		return "km"
	}
	return "m"
}

func (w *Weather) Conditions() string {
	switch w.weather.Timeserie[0].Symbol {
	case 1:
		return "Sunny"
	case 2, 102:
		return "Broken Clouds"
	case 3, 103:
		return "Cloudy"
	case 45:
		return "Fog"
	case 60, 80, 160, 180:
		return "Light Rain"
	case 63, 81, 163, 181:
		return "Heavy Rain"
	case 68, 83, 168, 183:
		return "Light Sleet"
	case 69, 84, 169, 184:
		return "Heavy Sleet"
	case 70, 85, 170, 185:
		return "Light Snow"
	case 73, 86, 173, 186:
		return "Heavy Snow"
	case 101:
		return "Clear"
	default:
		return "Undefined " + strconv.Itoa(w.weather.Timeserie[0].Symbol)
	}
}

func (w *Weather) Sunrise() string {
	return "0" + w.weather.Sunrise[:1] + ":" + w.weather.Sunrise[1:]
}

func (w *Weather) Sunset() string {
	return w.weather.Sunset[:2] + ":" + w.weather.Sunset[2:]
}

// Forecast gives five day-summary forecasts
func (w *Weather) Forecast() []*Forecast {
	var forecasts []*Forecast
	for i := 1; i < 6; i++ {
		f := new(Forecast)
		date := w.weather.AggData[i].Time
		t, err := time.Parse("20060102", date)
		if err != nil {
			fmt.Println(err)
		}
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			f.Weekend = true
		}
		f.Date = date[6:] + "/" + date[4:6]
		f.TempMin = fmt.Sprintf("%.0f", w.weather.AggData[i].MinTemp)
		f.TempMax = fmt.Sprintf("%.0f", w.weather.AggData[i].MaxTemp)
		f.PrecipitationAmount = fmt.Sprintf("%.1f", w.weather.AggData[i].PrecipSum)
		forecasts = append(forecasts, f)
	}
	return forecasts
}

// HourForecast returns 48 hours worth of spot forecast including current conditions
func (w *Weather) HourForecast() []*Hour {
	var hours []*Hour
	for i := 0; i < 48; i++ {
		h := new(Hour)
		hours = append(hours, h)
		hour, err := strconv.Atoi(w.weather.Timeserie[i].Time[8:10])
		h.Hour = hour
		if err != nil {
			log.Fatal(err)
		}
		h.Temperature = int(math.Round(w.weather.Timeserie[i].Temp * 10))
		switch w.weather.Timeserie[i].Symbol {
		case 1, 101:
			h.Sky = Clear
		case 2, 80, 81, 83, 84, 85, 86, 102, 180, 181, 183, 184, 185, 186:
			h.Sky = Broken
		case 3, 60, 63, 68, 69, 70, 73, 103, 160, 163, 168, 169, 170, 173:
			h.Sky = Cloudy
		case 45:
			h.Sky = Fog
		default:
			fmt.Println("Unknown symbol: ", w.weather.Timeserie[i].Symbol)
		}
		switch w.weather.Timeserie[i].Symbol {
		case 60, 80, 160, 180:
			h.Precipitation = LightRain
		case 63, 81, 163, 181:
			h.Precipitation = HeavyRain
		case 68, 83, 168, 183:
			h.Precipitation = LightSleet
		case 69, 84, 169, 184:
			h.Precipitation = HeavySleet
		case 70, 85, 170, 185:
			h.Precipitation = LightSnow
		case 73, 86, 173, 186:
			h.Precipitation = HeavySnow
		}

		h.PrecipitationAmount = int(math.Round(w.weather.Timeserie[i].Precip1))
	}
	return hours
}
