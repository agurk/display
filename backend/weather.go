package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Weather struct {
	Latitude  string
	Longitude string
	weather   data
}

type Forecast struct {
	Date          string
	TempMax       string
	TempMin       string
	Precipitation string
}

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

func (w *Weather) Conditions() string {
	switch w.weather.Timeserie[0].Symbol {
	case 1:
		return "Sunny"
	case 2, 102:
		return "Broken Clouds"
	case 3:
		return "Cloudy"
	case 45:
		return "Fog"
	case 60, 180:
		return "Light Rain"
	case 63, 181:
		return "Heavy Rain"
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

func (w *Weather) Forecast() []*Forecast {
	var forecasts []*Forecast
	for i := 1; i < 6; i++ {
		f := new(Forecast)
		date := w.weather.AggData[i].Time
		f.Date = date[4:6] + "/" + date[6:]
		f.TempMin = fmt.Sprintf("%.0f", w.weather.AggData[i].MinTemp)
		f.TempMax = fmt.Sprintf("%.0f", w.weather.AggData[i].MaxTemp)
		f.Precipitation = fmt.Sprintf("%.1f", w.weather.AggData[i].PrecipSum)
		forecasts = append(forecasts, f)
	}
	return forecasts
}
