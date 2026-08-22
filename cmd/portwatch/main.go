package main

import (
	"context"
	"os"

	"github.com/portwatch/portwatch/internal/command"
	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
)

func main() {
	deps := command.Dependencies{
		Scanner: port.NewWindowsScanner(),
		Manager: process.NewManager(),
	}
	os.Exit(command.Run(context.Background(), os.Args[1:], deps, os.Stdout, os.Stderr))
}
