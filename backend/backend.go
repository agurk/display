package main

import (
	"image"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/image/bmp"
)

func main() {
	width, height := 800, 480
	screen := NewScreen(width, height)
	screen.LoadFont("fonts/FontsFree-Net-HelveticaNeueMedium.ttf")

	// Title Box
	screen.DrawRect(image.Rect(0, 0, width, 50), image.Black)
	// Divide screen in half
	screen.DrawVerticalLine(width/2, 0, height-20)
	screen.DrawHorizontalLine(height-20, 0, width)

	// Title text
	screen.Write(dateNow(), width/2, 60, false, true)

	power := NewPower("/home/timothy/src/display/electricity.db")
	screen.Write("Current KWh Cost", 550, 100, true, false)
	screen.Write(strconv.Itoa(power.CurrentCost()), 700, 110, true, true)

	costGraph(screen, power)

	screen.Write(time.Now().Format("2006-01-02 15:04:05"), width/2, 489, true, false)

	weather := NewWeather("55.7034", "12.5823")
	screen.Write(weather.Conditions(), 60, 140, true, true)
	screen.Write(weather.Temp()+"° C", 200, 120, true, true)
	screen.Write(weather.Pressure()+" hPa", 320, 120, true, true)
	screen.Write(weather.WindSpeed()+" m/s", 200, 160, true, true)
	screen.Write(weather.WindDirection(), 350, 160, true, true)

	out, err := os.Create("out.bmp")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	err = bmp.Encode(out, screen.Image)
	if err != nil {
		log.Fatal(err)
	}
}

func costGraph(screen *Screen, power *Power) {
	prices, pos := power.CostData()
	// 48 hours shown, each bar has an 8 px slot to fit in with an 8px border
	x := 400
	seperator := 2
	for i := 0; i < 48; i++ {
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
		x += 8
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
