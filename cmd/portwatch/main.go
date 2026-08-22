package main

import (
	"context"
	"os"

	"github.com/portwatch/portwatch/internal/command"
)

func main() {
	os.Exit(command.Run(context.Background(), os.Args[1:], command.Dependencies{}, os.Stdout, os.Stderr))
}
