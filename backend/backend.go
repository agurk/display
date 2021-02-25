package main

import (
	"image"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	//"golang.org/x/image/bmp"
)

func main() {
	width, height := 800, 480
	screen := NewScreen(width, height)
	screen.LoadFont("fonts/FontsFree-Net-HelveticaNeueMedium.ttf")

	// Title Box
	screen.DrawRect(image.Rect(0, 0, width, 50), image.Black)
	screen.DrawHorizontalLine(height-20, 0, width)

	// Title text
	screen.Write(dateNow(), width/2, 25, false, true)

	/********* Electricity section ***********/
	power := NewPower("/home/timothy/src/display/electricity.db")
	screen.DrawRect(image.Rect(404, 55, 692, 85), image.Black)
	screen.Write("Current KWh Cost", 548, 70, false, false)
	screen.Write(strconv.Itoa(power.CurrentCost()), 750, 70, true, true)

	costGraph(screen, power)

	screen.DrawRect(image.Rect(440, 300, 760, 340), image.White)
	amount, day := power.DayUseage()
	screen.Write(day+"   "+amount+"KWh", 600, 320, true, true)

	/********* Weather Section ************/
	weather := NewWeather("55.7034", "12.5823")

	screen.Write(weather.Sunrise(), width/8, 25, false, false)
	screen.Write(weather.Sunset(), 7*width/8, 25, false, false)

	screen.Write(weather.Conditions(), 100, 80, true, true)
	screen.DrawHorizontalLine(110, 4, 192)
	screen.Write(weather.WindSpeed()+"m/s ("+weather.WindDirection()+")", 100, 150, true, true)
	screen.DrawRect(image.Rect(4, 168, 196, 188), image.Black)
	screen.Write(weather.WindGust()+"m/s gusts", 100, 178, false, false)

	screen.Write(weather.Temp()+"°C", 250, 75, true, true)
	screen.DrawRect(image.Rect(202, 90, 298, 110), image.Black)
	screen.Write(weather.MaxTemp()+" / "+weather.MinTemp()+"°C", 250, 100, false, false)

	screen.Write(weather.Precipitation()+"mm", 350, 75, true, true)
	screen.DrawRect(image.Rect(302, 90, 398, 110), image.Black)
	screen.Write(weather.DayPrecipitation()+"mm", 350, 100, false, false)

	screen.Write(weather.Humidity()+"%", 250, 135, true, true)
	screen.DrawHorizontalLine(152, 202, 96)
	screen.Write("UV "+weather.UV(), 350, 135, true, true)
	screen.DrawHorizontalLine(152, 302, 96)

	screen.Write(weather.Visibility()+weather.VisibiltyDistance(), 250, 170, true, true)
	screen.DrawHorizontalLine(187, 202, 96)
	screen.Write(weather.Pressure(), 350, 170, true, true)
	screen.DrawHorizontalLine(187, 302, 96)
	//screen.Write("hPa", 400, 173, true, false)

	// next five days
	x := 40
	y := 390
	for _, f := range weather.Forecast() {
		// 80-35-5 = 40
		screen.DrawRect(image.Rect(x-39, y, x+39, y+20), image.Black)
		screen.Write(f.Date, x, y+10, false, false)
		screen.Write(f.TempMax+" / "+f.TempMin+"°C", x, y+30, true, false)
		screen.Write(f.Precipitation+" mm", x, y+50, true, false)
		x += 80
	}

	weatherGraph(screen, weather)

	// when this was created
	screen.Write(time.Now().Format("2006-01-02 15:04:05"), width/2, height-10, true, false)

	out, err := os.Create("out.bmp")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	bits := screen.TwoBitImage()
	_, err = out.Write(bits)
	if err != nil {
		log.Fatal(err)
	}
	out.Sync()
	/*
		err = bmp.Encode(out, screen.Image)
		if err != nil {
			log.Fatal(err)
		}
	*/
}

func weatherGraph(screen *Screen, weather *Weather) {
	hours := weather.HourForecast()
	max, min := 0, 0
	for _, v := range hours {
		if v.Temperature < min {
			min = v.Temperature
		}
		if v.Temperature > max {
			max = v.Temperature
		}
	}

	x := 50
	for i, v := range hours {
		x += 7
		// split out the days
		if i > 0 && v.Hour == 0 {
			x += 4
			screen.DrawVerticalLine(x-5, 220, 160)
		}

		y := 370 - v.Temperature*10
		screen.DrawRect(image.Rect(x-2, y-2, x+2, y+2), image.Black)
		screen.DrawRect(image.Rect(x-3, y-1, x+3, y+1), image.Black)
		screen.DrawRect(image.Rect(x-1, y-3, x+1, y+3), image.Black)

		switch v.Symbol {
		case 1:
		case 2, 102:
			screen.DrawRect(image.Rect(x-3, 220, x-2, 228), image.Black)
			screen.DrawRect(image.Rect(x-1, 220, x, 228), image.Black)
			screen.DrawRect(image.Rect(x+1, 220, x+2, 228), image.Black)
			screen.DrawRect(image.Rect(x+3, 220, x+4, 228), image.Black)
		case 3, 103:
			screen.DrawRect(image.Rect(x-3, 220, x+4, 228), image.Black)
		case 45:
			screen.DrawRect(image.Rect(x-3, 220, x+4, 222), image.Black)
			screen.DrawRect(image.Rect(x-3, 223, x+4, 225), image.Black)
			screen.DrawRect(image.Rect(x-3, 226, x+4, 228), image.Black)
		}
	}
	screen.Write("cover", 25, 224, true, false)
	screen.DrawThinBlackLine(370, 50, 350)
	screen.Write("0°C", 25, 370, true, false)
	screen.DrawThinBlackLine(270, 50, 350)
	screen.Write("10°C", 25, 270, true, false)
}

func costGraph(screen *Screen, power *Power) {
	prices, pos := power.CostData()
	// 48 hours shown, each bar has an 8 px slot to fit in with an 8px border
	x := 400
	seperator := 2
	for i := 0; i < 48; i++ {
		x += 8
		// offsets control width of bar
		offset1 := 0
		offset2 := 7
		if i < pos {
			offset1 = 3
			offset2 = 4
		}
		// seperate out the two day blocks
		if i == 24 {
			x += 4
		}
		value := prices[i]
		y := 420 + seperator
		for ; value >= 100; value -= 100 {
			oldy := y - seperator
			y -= 100
			screen.DrawRect(image.Rect(x+offset1, y, x+offset2, oldy), image.Black)
		}
		y -= seperator
		newy := y - value
		screen.DrawRect(image.Rect(x+offset1, newy, x+offset2, y), image.Black)
	}
}

func dateNow() string {
	today := time.Now()
	suffix := "th"
	switch today.Day() {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return today.Format("Monday 2" + suffix + " January")
}
