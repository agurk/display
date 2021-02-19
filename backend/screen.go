package main

import (
	"image"
	"image/draw"
	"io/ioutil"
	"log"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Screen struct {
	Font          *truetype.Font
	Image         *image.Gray
	Width, Height int
}

func NewScreen(width, height int) *Screen {
	screen := new(Screen)
	screen.Width = width
	screen.Height = height
	screen.Image = image.NewGray(image.Rect(0, 0, width, height))
	draw.Draw(screen.Image, screen.Image.Bounds(), image.White, image.Point{}, draw.Src)
	return screen
}

func (screen *Screen) LoadFont(path string) {
	ttf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	screen.Font, err = truetype.Parse(ttf)
	if err != nil {
		log.Fatal(err)
	}
}

func (screen *Screen) LargeFace() font.Face {
	return truetype.NewFace(screen.Font, &truetype.Options{
		Size:    28,
		DPI:     72,
		Hinting: 1,
	})
}

func (screen *Screen) Face() font.Face {
	return truetype.NewFace(screen.Font, &truetype.Options{
		Size:    18,
		DPI:     72,
		Hinting: 1,
	})
}

func (screen *Screen) Write(text string, x, y int, black, large bool) {
	face := screen.Face()
	if large {
		face = screen.LargeFace()
	}
	bounds, advance := font.BoundString(face, text)

	x = int(x - advance.Round()/2)
	y = int(y + bounds.Min.Y.Round())

	colour := image.Black
	if !black {
		colour = image.White
	}

	d := &font.Drawer{
		Dst:  screen.Image,
		Src:  colour,
		Face: face,
		Dot:  fixed.P(x, y),
	}

	d.DrawString(text)
}

func (screen *Screen) DrawRect(rect image.Rectangle, colour *image.Uniform) {
	draw.Draw(screen.Image, rect, colour, image.Point{}, draw.Src)
}

func (screen *Screen) DrawHorizontalLine(height, start, length int) {
	end := start + length
	if end > screen.Width {
		end = screen.Width
	}
	height -= 1
	if height < 0 {
		height = 0
	}
	if height > screen.Height+2 {
		height = screen.Height - 2
	}
	screen.DrawRect(image.Rect(start, height, end, height+2), image.Black)
}

func (screen *Screen) DrawVerticalLine(vpos, start, length int) {
	end := start + length
	if end > screen.Height {
		end = screen.Height
	}
	vpos -= 1
	if vpos < 0 {
		vpos = 0
	}
	if vpos > screen.Width+2 {
		vpos = screen.Width - 2
	}
	screen.DrawRect(image.Rect(vpos, start, vpos+2, end), image.Black)
}