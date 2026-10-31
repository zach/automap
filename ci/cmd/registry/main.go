// Command registry publishes the Automap registry (ADR-0034, ADR-0065):
// it copies the ports kept on main in ci/registry/ports onto a checkout of
// the registry branch, records their versions in vcpkg's versions
// database, and prints the branch's new commit for the public map's
// map.json.
//
//	registry -ports ci/registry/ports -branch <registry checkout>
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ports := flag.String("ports", "ci/registry/ports", "the ports to publish")
	branch := flag.String("branch", "", "a checkout of the registry branch")
	flag.Parse()
	commit, err := Publish(*ports, *branch)
	if err != nil {
		fmt.Fprintln(os.Stderr, "registry:", err)
		os.Exit(1)
	}
	fmt.Println(commit)
}
