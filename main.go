package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type MixSequenceStep struct {
	Name  string `yaml:"name"`
	Tiles int    `yaml:"tiles"`
}

type CompositionDef struct {
	Direction   string            `yaml:"direction"`
	Mixes       []Mix             `yaml:"mixes"`
	MixSequence []MixSequenceStep `yaml:"mixSequence"`
	WidthTiles  int               `yaml:"widthTiles"`
}

func loadCompositionDef(path string) CompositionDef {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var cfg CompositionDef
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return cfg
}

func (cd *CompositionDef) Compose(inputPath string) {
	nextRenderAt := time.Now()
	// default composition direction is vertical
	var flipXy bool
	switch cd.Direction {
	case DirectionVertical:
		flipXy = false
	case DirectionHorizontal:
		flipXy = true
	default:
		panic(fmt.Errorf("unknown direction %q", cd.Direction))
	}
	var yImgs []image.Image
	for _, step := range cd.MixSequence {
		var mix *Mix
		for i := range cd.Mixes {
			if cd.Mixes[i].Name == step.Name {
				mix = &cd.Mixes[i]
				break
			}
		}
		if mix == nil {
			panic(fmt.Sprintf("mix %q not found", step.Name))
		}
		cached := mix.CachedRenders()
		yRemain := step.Tiles
		for {
			if yRemain <= 0 {
				break
			}
			yCropTiles := Size20.TilesPerRender
			var yCropPx int
			if yRemain < Size20.TilesPerRender {
				yCropTiles = yRemain
				yCropPx = int(math.Ceil(Size20.PxPerTile() * float64(yRemain)))
			}
			var xImgs []image.Image
			xRemain := cd.WidthTiles
			for {
				if xRemain <= 0 {
					break
				}
				var path string
				if len(cached) > 0 {
					path = cached[0]
					cached = cached[1:]
				} else {
					// be nice; don't peg their API
					if wait := time.Until(nextRenderAt); wait > 0 {
						time.Sleep(wait)
					}
					path = mix.Render()
					nextRenderAt = time.Now().Add(time.Second)
				}
				log.Printf("path for %q: %s", mix.Name, path)
				xCropTiles := Size20.TilesPerRender
				var xCropPx int
				if xRemain < Size20.TilesPerRender {
					xCropTiles = xRemain
					xCropPx = int(math.Ceil(Size20.PxPerTile() * float64(xRemain)))
				}
				var img image.Image
				if flipXy {
					img = LoadImage(path, yCropPx, xCropPx)
				} else {
					img = LoadImage(path, xCropPx, yCropPx)
				}
				xImgs = append(xImgs, img)
				xRemain -= xCropTiles
			}
			var xComp *image.RGBA
			if flipXy {
				xComp = Stitch(xImgs, DirectionVertical)
			} else {
				xComp = Stitch(xImgs, DirectionHorizontal)
			}
			yImgs = append(yImgs, xComp)
			yRemain -= yCropTiles
		}
	}
	comp := Stitch(yImgs, cd.Direction)
	outFilename := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".jpg"
	out, err := os.Create(outFilename)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	err = jpeg.Encode(out, comp, &jpeg.Options{Quality: 80})
	if err != nil {
		panic(err)
	}
	log.Printf("saved as %s", outFilename)
	w := comp.Bounds().Dx()
	h := comp.Bounds().Dy()
	log.Printf(
		"measurements: %d x %d px; %d x %d mm",
		w,
		h,
		int(math.Round(float64(comp.Bounds().Dx())*Size20.MmPerPx())),
		int(math.Round(float64(comp.Bounds().Dy())*Size20.MmPerPx())),
	)
	if w > 4096 || h > 4096 {
		log.Printf("warning: dimensions > 4096px, sketchup will downsample")
		log.Printf("(although 2-3x downsample may look sufficient depending on use case)")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bisazzy <composition.yaml>")
		os.Exit(1)
	}
	cfg := loadCompositionDef(os.Args[1])
	cfg.Compose(os.Args[1])
}
