package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type MixComponent struct {
	Product string `yaml:"product"`
	Percent int    `yaml:"percent"`
}

func (mc *MixComponent) ProductLineCode() string {
	return mc.Product[:2]
}

type Mix struct {
	Name       string         `yaml:"name"`
	Components []MixComponent `yaml:"components"`
	FillGel    string         `yaml:"fillgel"`
}

func (m *Mix) UnmarshalYAML(value *yaml.Node) error {
	type mixAlias struct {
		Name       string    `yaml:"name"`
		FillGel    string    `yaml:"fillgel"`
		Components yaml.Node `yaml:"components"`
	}
	var alias mixAlias
	if err := value.Decode(&alias); err != nil {
		return err
	}
	m.Name = alias.Name
	m.FillGel = alias.FillGel
	nodes := alias.Components.Content
	for i := 0; i < len(nodes); i += 2 {
		var percent int
		if err := nodes[i+1].Decode(&percent); err != nil {
			return err
		}
		m.Components = append(m.Components, MixComponent{Product: nodes[i].Value, Percent: percent})
	}
	return nil
}

func (m *Mix) SortedComponents() []MixComponent {
	mcs := make([]MixComponent, 0, len(m.Components))
	for _, mc := range m.Components {
		if mc.Percent == 0 {
			continue
		}
		mcs = append(mcs, mc)
	}
	sort.Slice(mcs, func(i, j int) bool {
		return mcs[i].Product < mcs[j].Product
	})
	return mcs
}

func (m *Mix) RenderUrl() string {
	total := 0
	taParts := make([]string, 0, len(m.Components))
	idx := 0
	for _, mc := range m.SortedComponents() {
		line := GetProductLine(mc.ProductLineCode())
		mixSeg := fmt.Sprintf(
			"ta%d=%d;%s_%s%s",
			idx,
			mc.Percent,
			Size20.Seg(),
			line.TextureSeg,
			strings.ReplaceAll(mc.Product, " ", "%20"),
		)
		taParts = append(taParts, mixSeg)
		total += mc.Percent
		idx++
	}
	if total != 100 {
		panic(fmt.Errorf("mix is at %d%%, must be 100%%", total))
	}
	return fmt.Sprintf(
		"https://app.bisazza.com/process.php?req=miscela&lang=en&fillgel=%s&layout=%s&altezza=1&%s",
		m.FillGel,
		Size20.Seg(),
		strings.Join(taParts, "&"),
	)
}

func (m *Mix) Hash() string {
	h := sha256.Sum256([]byte(m.RenderUrl()))
	return fmt.Sprintf("%x", h[:4])
}

func (m *Mix) CachedRenders() []string {
	matches, _ := filepath.Glob(fmt.Sprintf("cache/%s_*.jpg", m.Hash()))
	rand.Shuffle(len(matches), func(i, j int) { matches[i], matches[j] = matches[j], matches[i] })
	return matches
}

func (m *Mix) NextCachePath() string {
	h := m.Hash()
	if err := os.MkdirAll("cache", 0755); err != nil {
		panic(err)
	}
	matches, err := filepath.Glob(fmt.Sprintf("cache/%s_*.jpg", h))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("cache/%s_%d.jpg", h, len(matches)+1)
}

func (m *Mix) Render() string {
	url := m.RenderUrl()
	path := m.NextCachePath()
	fmt.Printf("downloading %s -> %s\n", m.Name, path)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("server returned %s", resp.Status))
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		panic(err)
	}
	return path
}
