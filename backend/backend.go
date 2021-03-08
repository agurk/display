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

	screen.DrawRect(image.Rect(420, 345, 780, 450), image.White)
	screen.DrawRect(image.Rect(430, 350, 535, 370), image.Black) // 482
	screen.DrawRect(image.Rect(545, 350, 655, 370), image.Black) // 600
	screen.DrawRect(image.Rect(665, 350, 770, 370), image.Black) // 718
	usage := power.WeekUseage()
	screen.Write("Last Week", 482, 360, false, false)
	screen.Write(usage.Amount+"KWh", 482, 385, true, false)
	screen.Write(usage.Cost, 482, 410, true, false)
	screen.Write(usage.Efficiency+"%", 482, 435, true, false)
	usage = power.PrevDayUseage()
	screen.Write(usage.Date, 600, 360, false, false)
	screen.Write(usage.Amount+"KWh", 600, 385, true, false)
	screen.Write(usage.Cost, 600, 410, true, false)
	screen.Write(usage.Efficiency+"%", 600, 435, true, false)
	usage = power.DayUseage()
	screen.Write(usage.Date, 718, 360, false, false)
	screen.Write(usage.Amount+"KWh", 718, 385, true, false)
	screen.Write(usage.Cost, 718, 410, true, false)
	screen.Write(usage.Efficiency+"%", 718, 435, true, false)

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
	bits := screen.OneBitImage()
	_, err = out.Write(bits)
	if err != nil {
		log.Fatal(err)
	}
	/*
		out.Sync()
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

	// allow the point to go one dregree above or below the line
	// otherwise round to the nearest 10 line
	if min%10 == 9 {
		min++
	} else {
		min -= min % 10
	}

	if max%10 == 1 {
		max--
	} else {
		max += 10 - max%10
	}

	if max-min < 10 {
		max += 10
	}

	yMax, yMin := 350, 240
	yDegree := (yMax - yMin) / (max - min)
	x := 50
	for i, v := range hours {
		x += 7
		// split out the days
		if i > 0 && v.Hour == 0 {
			x += 4
			screen.DrawVerticalLine(x-5, 200, 180)
		}

		if v.Precipitation > 0 {
			screen.DrawRect(image.Rect(x-3, 370, x+4, 370-v.Precipitation), image.Black)
		}

		y := yMax - v.Temperature*yDegree
		// white box so visible if lots of precipitation
		screen.DrawRect(image.Rect(x-2, y-2, x+2, y+2), image.White)
		screen.DrawRect(image.Rect(x-2, y-2, x+2, y+2), image.Black)
		screen.DrawRect(image.Rect(x-3, y-1, x+3, y+1), image.Black)
		screen.DrawRect(image.Rect(x-1, y-3, x+1, y+3), image.Black)

		switch v.Symbol {
		case 1:
		case 2, 102:
			screen.DrawRect(image.Rect(x-3, 205, x-2, 215), image.Black)
			screen.DrawRect(image.Rect(x-1, 205, x, 215), image.Black)
			screen.DrawRect(image.Rect(x+1, 205, x+2, 215), image.Black)
			screen.DrawRect(image.Rect(x+3, 205, x+4, 215), image.Black)
		case 3, 103:
			screen.DrawRect(image.Rect(x-3, 205, x+4, 215), image.Black)
		case 45:
			screen.DrawRect(image.Rect(x-3, 205, x+4, 207), image.Black)
			screen.DrawRect(image.Rect(x-3, 209, x+4, 211), image.Black)
			screen.DrawRect(image.Rect(x-3, 213, x+4, 215), image.Black)
		}

	}
	screen.Write("cover", 25, 210, true, false)
	for degrees := min; degrees <= max; degrees += 10 {
		screen.DrawThinBlackLine(yMax-(degrees-min)*yDegree, 50, 350)
		screen.Write(strconv.Itoa(degrees)+"°C", 25, yMax-(degrees-min)*yDegree, true, false)
	}
	screen.Write("precip", 25, 370, true, false)
}

func costGraph(screen *Screen, power *Power) {
	prices, pos := power.CostData()
	// 48 hours shown, each bar has an 8 px slot to fit in with an 8px border
	x := 400
	yScale := 3.0 / 4.0
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
		y := 340 + seperator
		for ; value >= 100; value -= 100 {
			oldy := y - seperator
			y -= int(100.0 * yScale)
			screen.DrawRect(image.Rect(x+offset1, y, x+offset2, oldy), image.Black)
		}
		y -= seperator
		newy := y - int(float64(value)*yScale)
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
