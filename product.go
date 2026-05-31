package main

import (
	"fmt"
)

type ProductLine struct {
	Code       string
	TextureSeg string
}

// there are also a lot more product lines available
// but i haven't needed them so far!

var Smalto = ProductLine{
	Code:       "SM",
	TextureSeg: "smalto/tex_",
}

var Gloss = ProductLine{
	Code:       "GL",
	TextureSeg: "gloss/tex_",
}

func GetProductLine(code string) ProductLine {
	if code == Smalto.Code {
		return Smalto
	}
	if code == Gloss.Code {
		return Gloss
	}
	panic(fmt.Errorf("product line for '%s' not found", code))
}

type ProductSize struct {
	TileEdgeMm     int
	GapMm          int
	RenderEdgePx   int
	TilesPerRender int
}

func (ps *ProductSize) Seg() string {
	return fmt.Sprintf("%dx%d", ps.TileEdgeMm, ps.TileEdgeMm)
}

func (ps *ProductSize) PxPerMm() float64 {
	// in render file,
	// 15 tiles + 15 gaps (14 between the tiles, 1/2 on each side)
	renderMm := float64(ps.TileEdgeMm*ps.TilesPerRender) + float64(ps.GapMm*ps.TilesPerRender)
	return float64(ps.RenderEdgePx) / renderMm
}

func (ps *ProductSize) MmPerPx() float64 {
	return 1 / ps.PxPerMm()
}

func (ps *ProductSize) PxPerTile() float64 {
	return float64(ps.TileEdgeMm+ps.GapMm) * ps.PxPerMm()
}

// only dealing with the 20mm tiles at the moment
// the 15x15 tiles render comes out as 1000px
var Size20 = ProductSize{
	TileEdgeMm:     20,
	GapMm:          2,
	RenderEdgePx:   1000,
	TilesPerRender: 15,
}
