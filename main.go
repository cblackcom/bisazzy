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
	var imgs []image.Image
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
		stepTilesAdded := 0
		for {
			tilesRequired := step.Tiles - stepTilesAdded
			if tilesRequired <= 0 {
				break
			}
			var path string
			if len(cached) > 0 {
				path = cached[0]
				cached = cached[1:]
			} else {
				path = mix.Render()
			}
			log.Printf("path for %q: %s", mix.Name, path)
			var cropPx int
			if tilesRequired < Size20.TilesPerRender {
				cropPx = int(math.Ceil(Size20.PxPerTile() * float64(tilesRequired)))
			}
			img := LoadImage(path, cropPx, cd.Direction)
			imgs = append(imgs, img)
			stepTilesAdded += tilesRequired
		}
	}
	comp := Stitch(imgs, cd.Direction)
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
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bisazzy <composition.yaml>")
		os.Exit(1)
	}
	cfg := loadCompositionDef(os.Args[1])
	cfg.Compose(os.Args[1])
}
