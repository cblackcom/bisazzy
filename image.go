package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
)

const DirectionHorizontal = "horizontal"
const DirectionVertical = "vertical"

func LoadImage(path string, cropPx int, direction string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	img, err := jpeg.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	if cropPx > 0 {
		type subImager interface {
			SubImage(r image.Rectangle) image.Image
		}
		si, ok := img.(subImager)
		if !ok {
			panic(fmt.Errorf("subImager not ok"))
		}
		b := img.Bounds()
		switch direction {
		case DirectionHorizontal:
			if cropPx < img.Bounds().Dx() {
				img = si.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X+cropPx, b.Min.Y))
			}
		case DirectionVertical:
			if cropPx < img.Bounds().Dy() {
				img = si.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+cropPx))
			}
		default:
			panic(fmt.Errorf("unknown direction %q", direction))
		}
	}
	return img
}

func Stitch(imgs []image.Image, direction string) *image.RGBA {
	totalWidth, totalHeight, maxWidth, maxHeight := 0, 0, 0, 0
	for _, img := range imgs {
		totalWidth += img.Bounds().Dx()
		totalHeight += img.Bounds().Dy()
		if img.Bounds().Dx() > maxWidth {
			maxWidth = img.Bounds().Dx()
		}
		if img.Bounds().Dy() > maxHeight {
			maxHeight = img.Bounds().Dy()
		}
	}
	var canvas *image.RGBA
	switch direction {
	case DirectionVertical:
		canvas = image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
		y := 0
		for _, img := range imgs {
			draw.Draw(canvas, image.Rect(0, y, maxWidth, y+img.Bounds().Dy()), img, img.Bounds().Min, draw.Src)
			y += img.Bounds().Dy()
		}
	case DirectionHorizontal:
		canvas = image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))
		x := 0
		for _, img := range imgs {
			draw.Draw(canvas, image.Rect(x, 0, x+img.Bounds().Dx(), maxHeight), img, img.Bounds().Min, draw.Src)
			x += img.Bounds().Dx()
		}
	default:
		panic(fmt.Errorf("unknown direction %q", direction))
	}
	return canvas
}
