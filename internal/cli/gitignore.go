package cli

import (
	"os"
	"strings"
)

// ignoreBuild returns the repo's .gitignore with a line that ignores dir,
// "/build/" or "/build/automap/", unless a line already ignores it, and
// whether the file changes. verify builds into build/automap/, and a
// testbed's own build goes into build/, so neither belongs in git.
func ignoreBuild(dir string) (string, bool) {
	raw, _ := os.ReadFile(".gitignore")
	existing := string(raw)
	covers := map[string]bool{"build": true, "build/": true, "/build": true, "/build/": true, "build/*": true, "/build/*": true}
	if dir != "/build/" {
		trimmed := strings.Trim(dir, "/")
		for _, p := range []string{trimmed, trimmed + "/", "/" + trimmed, "/" + trimmed + "/"} {
			covers[p] = true
		}
	}
	for _, line := range strings.Split(existing, "\n") {
		if covers[strings.TrimSpace(line)] {
			return existing, false
		}
	}
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	return existing + dir + "\n", true
}
