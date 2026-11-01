package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/release"
)

// FreshnessWindow is how long after an upstream release the map may lack
// it and the library still count as fresh (ADR-0063): the 7-day cooldown
// and a week to curate.
const FreshnessWindow = 14 * 24 * time.Hour

// Measure is the map's part of the v1 measures (ADR-0063): freshness,
// and how much of the map's evidence is current.
type Measure struct {
	Freshness Freshness `json:"freshness"`
	Evidence  Currency  `json:"evidence"`
}

// Freshness counts the watched libraries and those whose latest upstream
// release is in the map, or newer than the window.
type Freshness struct {
	Libraries int      `json:"libraries"`
	Fresh     int      `json:"fresh"`
	Share     float64  `json:"share"`
	Behind    []string `json:"behind,omitempty"` // libraries with an upstream release older than the window the map lacks
}

// Currency counts the map's evidence and the pieces that are current.
type Currency struct {
	Pieces  int     `json:"pieces"`
	Current int     `json:"current"`
	Share   float64 `json:"share"`
}

// freshness judges each watched library by the upstream releases the map
// lacks: fresh when there are none, or when the newest was released less
// than the window ago. It is a pure function.
func freshness(watched []string, missing []Release, now time.Time) Freshness {
	oldest := map[string]time.Time{}
	for _, r := range missing {
		at, err := time.Parse("2006-01-02", r.Released)
		if err != nil {
			continue
		}
		if prev, ok := oldest[r.Library]; !ok || at.Before(prev) {
			oldest[r.Library] = at
		}
	}
	f := Freshness{Libraries: len(watched)}
	for _, lib := range watched {
		if at, ok := oldest[lib]; ok && now.Sub(at) >= FreshnessWindow {
			f.Behind = append(f.Behind, lib)
			continue
		}
		f.Fresh++
	}
	sort.Strings(f.Behind)
	if f.Libraries > 0 {
		f.Share = float64(f.Fresh) / float64(f.Libraries)
	}
	return f
}

// evidenceCurrency reads every piece of evidence under the map and counts
// the current ones.
func evidenceCurrency(mapDir string, now time.Time) Currency {
	var c Currency
	filepath.WalkDir(filepath.Join(mapDir, "evidence"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".json") || strings.HasSuffix(p, ".sigstore.json") {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		_, state, err := release.State(raw, now)
		if err != nil {
			return nil
		}
		c.Pieces++
		if state == "current" {
			c.Current++
		}
		return nil
	})
	if c.Pieces > 0 {
		c.Share = float64(c.Current) / float64(c.Pieces)
	}
	return c
}

func writeMeasure(file string, m Measure) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, append(raw, '\n'), 0o644)
}
