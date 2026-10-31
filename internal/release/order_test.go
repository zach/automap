package release_test

import (
	"testing"

	"github.com/zach/automap/internal/release"
)

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"map-2026.10.01.10", "map-2026.10.01.2", true},
		{"map-2026.10.01.1", "map-2026.10.01", true},
		{"map-2026.10.02", "map-2026.10.01.9", true},
		{"map-2026.10.01", "map-2026.10.01", false},
		{"map-2026.09.30.11", "map-2026.10.01", false},
	} {
		if got := release.Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
	for _, tag := range []string{"map-2026.10.01", "map-2026.10.01.1", "map-2026.10.01.12"} {
		if !release.Tag.MatchString(tag) {
			t.Errorf("%s isn't a release tag", tag)
		}
	}
	for _, tag := range []string{"map-2026.10.01.0", "map-2026.10.01.01", "map-2026.10.1"} {
		if release.Tag.MatchString(tag) {
			t.Errorf("%s is a release tag", tag)
		}
	}
}
