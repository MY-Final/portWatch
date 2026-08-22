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
var Version = "0.5.0"

type flagOptions struct {
	Protocol   string
	Ports      string
	PortsSet   bool
	Process    string
	ProcessSet bool
	PIDs       string
	PIDsSet    bool
	State      string
	StateSet   bool
	Help       bool
	Version    bool
	JSON       bool
	Interval   time.Duration
}

func parseFlags(args []string) (flagOptions, []string, error) {
	var options flagOptions
	for _, arg := range args {
		if arg == "--ports" || strings.HasPrefix(arg, "--ports=") {
			options.PortsSet = true
		}
		if arg == "--process" || strings.HasPrefix(arg, "--process=") {
			options.ProcessSet = true
		}
		if arg == "--pid" || strings.HasPrefix(arg, "--pid=") {
			options.PIDsSet = true
		}
		if arg == "--state" || strings.HasPrefix(arg, "--state=") {
			options.StateSet = true
		}
	}
	set := flag.NewFlagSet("portwatch", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	set.StringVar(&options.Protocol, "protocol", "tcp", "network protocol")
	set.StringVar(&options.Ports, "ports", "", "comma-separated ports")
	set.StringVar(&options.Process, "process", "", "filter by process name")
	set.StringVar(&options.PIDs, "pid", "", "filter by comma-separated PIDs")
	set.StringVar(&options.State, "state", "", "filter by port state")
	set.BoolVar(&options.Help, "help", false, "show help")
	set.BoolVar(&options.Help, "h", false, "show help")
	set.BoolVar(&options.Version, "version", false, "show version")
	set.BoolVar(&options.JSON, "json", false, "output JSON")
	set.DurationVar(&options.Interval, "interval", time.Second, "watch interval")
	if err := set.Parse(args); err != nil {
		return flagOptions{}, nil, fmt.Errorf("parse flags: %w", err)
	}
	options.Protocol = strings.ToLower(options.Protocol)
	if options.Protocol != "tcp" && options.Protocol != "udp" && options.Protocol != "all" {
		return flagOptions{}, nil, fmt.Errorf("unsupported protocol %q", options.Protocol)
	}
	return options, set.Args(), nil
}

func (o flagOptions) queryFilter() (QueryFilter, error) {
	filter := QueryFilter{Process: strings.TrimSpace(o.Process)}
	if o.ProcessSet && filter.Process == "" {
		return QueryFilter{}, errors.New("--process cannot be empty")
	}
	if o.PIDsSet {
		pids, err := parsePIDSet(o.PIDs)
		if err != nil {
			return QueryFilter{}, err
		}
		filter.PIDs = pids
	}
	state := strings.ToUpper(strings.TrimSpace(o.State))
	if o.StateSet {
		if state == "" {
			return QueryFilter{}, errors.New("--state cannot be empty")
		}
		if !supportedState(state) {
			return QueryFilter{}, fmt.Errorf("unsupported port state %q", o.State)
		}
		filter.State = state
	}
	return filter, nil
}

func isFlagError(err error) bool {
	return err != nil && !errors.Is(err, flag.ErrHelp)
}
