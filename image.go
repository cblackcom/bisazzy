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

func LoadImage(path string, xCropPx int, yCropPx int) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	img, err := jpeg.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	if xCropPx <= 0 && yCropPx <= 0 {
		return img
	}
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	si, ok := img.(subImager)
	if !ok {
		panic(fmt.Errorf("subImager not ok"))
	}
	b := img.Bounds()
	x1 := b.Max.X
	y1 := b.Max.Y
	if xCropPx > 0 {
		x1 = b.Min.X + xCropPx
	}
	if yCropPx > 0 {
		y1 = b.Min.Y + yCropPx
	}
	img = si.SubImage(image.Rect(b.Min.X, b.Min.Y, x1, y1))
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
