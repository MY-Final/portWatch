package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/portwatch/portwatch/internal/command"
	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
)

var version = "0.1.0"

func main() {
	command.Version = version
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	deps := command.Dependencies{
		Scanner: port.NewScanner(),
		Manager: process.NewManager(),
	}
	os.Exit(command.Run(ctx, os.Args[1:], deps, os.Stdout, os.Stderr))
}
