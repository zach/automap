// Command watch is the map's watcher (ADR-0051): for each library in the
// map whose source is on GitHub, it lists the upstream releases and tags
// newer than the map's newest version, by the pattern of that version's
// pinned tag, with each one's commit, release date and whether its notes
// name a security fix. It writes them as JSON for the curation workflow.
//
//	watch -map map -out releases.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zach/automap/ci/internal/gh"
	"github.com/zach/automap/internal/mapmodel"
)

// Release is an upstream release the map doesn't have yet.
type Release struct {
	Library     string `json:"library"`
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Commit      string `json:"commit"`
	Released    string `json:"released"`
	SecurityFix bool   `json:"security_fix"`
	Repo        string `json:"repo"`
}

func main() {
	mapDir := flag.String("map", "map", "the public map")
	out := flag.String("out", "releases.json", "where to write the new releases")
	flag.Parse()
	m, problems := mapmodel.Load(*mapDir, nil)
	if len(problems) > 0 {
		fail(fmt.Errorf("the map has problems: %v", problems))
	}
	found := []Release{}
	for id, lib := range m.Libraries {
		if lib.Source == nil {
			continue
		}
		repo, ok := strings.CutPrefix(lib.Source.URL, "https://github.com/")
		if !ok {
			continue
		}
		v := lib.Newest()
		if v.Pin == nil || v.Pin.Tag == "" {
			continue
		}
		releases, err := upstream(id, repo, v.Pin.Tag, v.Version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", id, err)
			continue
		}
		found = append(found, releases...)
	}
	raw, _ := json.MarshalIndent(found, "", "  ")
	for _, r := range found {
		fmt.Printf("%s %s (%s, released %s, security fix %v)\n", r.Library, r.Version, r.Tag, r.Released, r.SecurityFix)
	}
	if err := os.WriteFile(*out, raw, 0o644); err != nil {
		fail(err)
	}
}

// upstream lists a library's releases newer than the map's newest version.
func upstream(lib, repo, pinTag, pinVersion string) ([]Release, error) {
	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := gh.Get("/repos/"+repo+"/tags?per_page=100", &tags); err != nil {
		return nil, err
	}
	var releases []struct {
		Tag        string `json:"tag_name"`
		Published  string `json:"published_at"`
		Body       string `json:"body"`
		Prerelease bool   `json:"prerelease"`
		Draft      bool   `json:"draft"`
	}
	gh.Get("/repos/"+repo+"/releases?per_page=100", &releases)
	byTag := map[string]string{}
	var versions []string
	for _, t := range tags {
		if v := versionOf(pinTag, pinVersion, t.Name); v != "" {
			byTag[v] = t.Name
			versions = append(versions, v)
			byTag[v+" commit"] = t.Commit.SHA
		}
	}
	var out []Release
	for _, v := range newer(pinVersion, versions) {
		r := Release{Library: lib, Version: v, Tag: byTag[v], Commit: byTag[v+" commit"], Repo: repo}
		for _, rel := range releases {
			if rel.Tag == r.Tag && !rel.Draft && !rel.Prerelease {
				r.Released, r.SecurityFix = rel.Published[:min(10, len(rel.Published))], securityFix(rel.Body)
			}
		}
		if r.Released == "" {
			var c struct {
				Commit struct {
					Committer struct {
						Date string `json:"date"`
					} `json:"committer"`
				} `json:"commit"`
			}
			if err := gh.Get("/repos/"+repo+"/commits/"+r.Commit, &c); err == nil {
				r.Released = c.Commit.Committer.Date[:min(10, len(c.Commit.Committer.Date))]
			}
		}
		out = append(out, r)
	}
	return out, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "watch:", err)
	os.Exit(1)
}
