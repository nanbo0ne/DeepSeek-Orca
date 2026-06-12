// Command deepseek-orca is a config- and plugin-driven coding agent CLI.
package main

import (
	"os"

	"deepseek-orca/internal/cli"

	// Blank imports wire compile-time built-ins into their registries.
	_ "deepseek-orca/internal/provider/anthropic"
	_ "deepseek-orca/internal/provider/openai"
	_ "deepseek-orca/internal/tool/builtin"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version))
}
