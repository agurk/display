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
	screen.DrawVerticalLine(400, 0, height)

	// Title text
	screen.Write(dateNow(), 400, 50, false, true)

	power := NewPower("/home/timothy/src/display/electricity.db")
	screen.Write("Current KWh Cost", 550, 100, true, false)
	screen.Write(strconv.Itoa(power.CurrentCost()), 700, 110, true, true)

	costGraph(screen, power)

	screen.Write(time.Now().Format("2006-01-02 15:04:05"), 600, 480, true, false)

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
	pixels := 400
	for i := 0; i < 48; i++ {
		pixels += 8
		value := 400 - prices[i]
		if i < pos {
			screen.DrawRect(image.Rect(pixels+3, value, pixels+4, 400), image.Black)
		} else {
			screen.DrawRect(image.Rect(pixels, value, pixels+7, 400), image.Black)
		}
	}
	screen.DrawVerticalLine(500, 400, 10)
	screen.DrawVerticalLine(600, 400, 15)
	screen.DrawVerticalLine(700, 400, 10)
	screen.DrawThinWhiteLine(200, 408, 384)
	screen.DrawThinWhiteLine(300, 408, 384)
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
