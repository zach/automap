// Command automap composes C++ game engines from libraries in the Automap map.
package main

import (
	"os"

	"github.com/zach/automap/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
