package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/MY-Final/portWatch/internal/command"
	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
)

var version = "0.9.0"

func main() {
	command.Version = version
	command.BinaryName = invocationName(os.Args[0])
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	deps := command.Dependencies{
		Scanner: port.NewScanner(),
		Manager: process.NewManager(),
	}
	os.Exit(command.Run(ctx, os.Args[1:], deps, os.Stdout, os.Stderr))
}

// invocationName derives the display name from argv[0] so an installed copy
// named pw.exe prints pw in usage and error prefixes instead of portwatch.
func invocationName(arg0 string) string {
	base := filepath.Base(arg0)
	if ext := filepath.Ext(base); strings.EqualFold(ext, ".exe") {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "" {
		return "portwatch"
	}
	return base
}
