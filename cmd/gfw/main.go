// Command gfw runs the Great Firewall control plane.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/thomasilliet/gfw/internal/command"
)

// Build metadata is injected with -ldflags by release builds.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// main executes the Cobra root command.
func main() {
	info := command.BuildInfo{Version: version, Commit: commit, Date: date}
	if err := command.New(info).ExecuteContext(context.Background()); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
