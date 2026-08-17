// Command orca is a config- and plugin-driven coding agent CLI.
package main

import (
	"os"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/cli"

	// Blank imports wire compile-time built-ins into their registries.
	_ "github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider/anthropic"
	_ "github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider/openai"
	_ "github.com/nanbo0ne/O.R.C.A-for-Windows/internal/tool/builtin"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version))
}
