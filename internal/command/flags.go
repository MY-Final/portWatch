package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

// Version is overridden by the release build through main.version.
var Version = "0.1.0"

type flagOptions struct {
	Protocol string
	Help     bool
	Version  bool
	JSON     bool
	Interval time.Duration
}

func parseFlags(args []string) (flagOptions, []string, error) {
	var options flagOptions
	set := flag.NewFlagSet("portwatch", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	set.StringVar(&options.Protocol, "protocol", "tcp", "network protocol")
	set.BoolVar(&options.Help, "help", false, "show help")
	set.BoolVar(&options.Help, "h", false, "show help")
	set.BoolVar(&options.Version, "version", false, "show version")
	set.BoolVar(&options.JSON, "json", false, "output JSON")
	set.DurationVar(&options.Interval, "interval", time.Second, "watch interval")
	if err := set.Parse(args); err != nil {
		return flagOptions{}, nil, fmt.Errorf("parse flags: %w", err)
	}
	if !strings.EqualFold(options.Protocol, "tcp") {
		return flagOptions{}, nil, fmt.Errorf("unsupported protocol %q", options.Protocol)
	}
	return options, set.Args(), nil
}

func isFlagError(err error) bool {
	return err != nil && !errors.Is(err, flag.ErrHelp)
}
