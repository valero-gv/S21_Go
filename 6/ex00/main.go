package main

import (
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA, x, y int, label string, col color.Color) {
	colon := color.RGBA{255, 255, 255, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(colon),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func main() {
	width, height := 300, 300
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bgColor := colornames.Cornflowerblue
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	charColor := colornames.Yellow
	draw.Draw(img, image.Rect(100, 100, 200, 200), &image.Uniform{charColor}, image.Point{}, draw.Src)

	eyeColor := colornames.Black
	draw.Draw(img, image.Rect(120, 130, 140, 150), &image.Uniform{eyeColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(160, 130, 180, 150), &image.Uniform{eyeColor}, image.Point{}, draw.Src)

	draw.Draw(img, image.Rect(130, 170, 170, 175), &image.Uniform{eyeColor}, image.Point{}, draw.Src)

	flagColor := colornames.Red
	draw.Draw(img, image.Rect(210, 120, 280, 150), &image.Uniform{flagColor}, image.Point{}, draw.Src)
	addLabel(img, 215, 135, "Amazing!", colornames.White)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}
	filePath := filepath.Join(cwd, "amazing_logo.png")

	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		log.Fatalf("Error encoding image to PNG: %v", err)
	}

	log.Println("Logo created successfully: amazing_logo.png")
}
