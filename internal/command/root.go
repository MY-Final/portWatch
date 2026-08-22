package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/portwatch/portwatch/pkg/model"
)

// PortScanner is the scanner capability used by command handlers. It is kept
// structural so platform implementations can be injected without a framework.
type PortScanner interface {
	List(context.Context) ([]model.PortInfo, error)
	Port(context.Context, int) ([]model.PortInfo, error)
}

// ProcessManager is the process capability used by command handlers.
type ProcessManager interface {
	Info(context.Context, int) (model.ProcessInfo, error)
	Exists(context.Context, int) (bool, error)
	Terminate(context.Context, int) error
}

// Dependencies contains runtime capabilities used by command handlers.
// Keeping them explicit avoids package-level mutable state and makes tests
// independent of the host operating system.
type Dependencies struct {
	Scanner PortScanner
	Manager ProcessManager
}

// Action identifies the command selected by the root argument parser.
type Action uint8

const (
	ActionList Action = iota
	ActionPort
	ActionFree
	ActionKill
	ActionFind
	ActionWatch
	ActionHelp
	ActionVersion
)

// Command is the parsed root command. Port is set for ActionPort and
// ActionFree; ActionList represents an invocation with no arguments.
type Command struct {
	Action Action
	Port   int
	PID    int
	Query  string
	Flags  flagOptions
}

// Parse parses the MVP root arguments without executing any platform work.
// An empty argument list means list all listening ports. A single decimal
// argument means inspect that port. Other commands are reserved for later
// wiring and return a structured ParseError.
func Parse(args []string) (Command, error) {
	options, positional, flagErr := parseFlags(args)
	if flagErr != nil {
		return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: flagErr.Error(), Message: flagErr.Error()}
	}
	if options.Help {
		return Command{Action: ActionHelp, Flags: options}, nil
	}
	if options.Version {
		return Command{Action: ActionVersion, Flags: options}, nil
	}
	if len(positional) == 0 {
		return Command{Action: ActionList, Flags: options}, nil
	}
	switch len(positional) {
	case 0:
	case 1:
		arg := positional[0]
		switch arg {
		case "help":
			return Command{Action: ActionHelp, Flags: options}, nil
		case "watch":
			return Command{Action: ActionWatch, Flags: options}, nil
		case "free":
			return Command{}, &ParseError{Kind: ParseErrorFree, Argument: arg, Message: "free requires a port and is not available yet"}
		}
		port, err := strconv.Atoi(arg)
		if err != nil || port < 1 || port > 65535 {
			return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: "port must be a number from 1 to 65535"}
		}
		return Command{Action: ActionPort, Port: port, Flags: options}, nil
	case 2:
		if positional[0] == "free" {
			port, err := strconv.Atoi(positional[1])
			if err != nil || port < 1 || port > 65535 {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: positional[1], Message: "port must be a number from 1 to 65535"}
			}
			return Command{Action: ActionFree, Port: port, Flags: options}, nil
		}
		if positional[0] == "kill" {
			pid, err := strconv.Atoi(positional[1])
			if err != nil || pid <= 0 {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPID, Argument: positional[1], Message: "pid must be a positive number"}
			}
			return Command{Action: ActionKill, PID: pid, Flags: options}, nil
		}
		if positional[0] == "find" && strings.TrimSpace(positional[1]) != "" {
			return Command{Action: ActionFind, Query: positional[1], Flags: options}, nil
		}
	case 3:
		if positional[0] == "find" && strings.TrimSpace(positional[1]) != "" {
			return Command{Action: ActionFind, Query: strings.Join(positional[1:], " "), Flags: options}, nil
		}
	}
	return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: positional[0], Message: "unknown command or arguments"}
}

// ParseErrorKind identifies the reason a root argument list was rejected.
type ParseErrorKind uint8

const (
	ParseErrorInvalidPort ParseErrorKind = iota + 1
	ParseErrorUnknownCommand
	ParseErrorHelp
	ParseErrorFree
	ParseErrorInvalidPID
)

// ParseError is a machine-readable root argument error.
type ParseError struct {
	Kind     ParseErrorKind
	Argument string
	Message  string
}

func (e *ParseError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Run parses args, executes the selected action, and returns a process exit
// code. It never exits the process itself, which keeps it straightforward to
// test and embed.
func Run(ctx context.Context, args []string, deps Dependencies, stdout io.Writer, stderr io.Writer) int {
	return run(ctx, args, deps, os.Stdin, stdout, stderr)
}

func run(ctx context.Context, args []string, deps Dependencies, stdin io.Reader, stdout, stderr io.Writer) int {
	if stderr == nil {
		stderr = io.Discard
	}
	if stdout == nil {
		stdout = io.Discard
	}
	command, err := Parse(args)
	if err != nil {
		var parseErr *ParseError
		if errors.As(err, &parseErr) {
			_, _ = fmt.Fprintf(stderr, "portwatch: %s\nusage: portwatch [port]\n", parseErr)
		} else {
			_, _ = fmt.Fprintf(stderr, "portwatch: %s\n", err)
		}
		return ExitCode(err)
	}
	if command.Action == ActionHelp {
		_, _ = fmt.Fprintln(stdout, "PortWatch - Windows TCP port diagnostics")
		_, _ = fmt.Fprintln(stdout, "Usage: portwatch [flags] [port]")
		_, _ = fmt.Fprintln(stdout, "       portwatch free <port>")
		_, _ = fmt.Fprintln(stdout, "       portwatch kill <pid>")
		_, _ = fmt.Fprintln(stdout, "       portwatch find <name>")
		return ExitSuccess
	}
	if command.Action == ActionVersion {
		_, _ = fmt.Fprintln(stdout, Version)
		return ExitSuccess
	}
	if deps.Scanner == nil || deps.Manager == nil {
		err := errors.New("portwatch dependencies are not configured")
		PrintError(stderr, err)
		return ExitCode(err)
	}

	switch command.Action {
	case ActionList:
		return runList(ctx, deps, command.Flags.JSON, stdout, stderr)
	case ActionPort:
		return runPort(ctx, deps, command.Port, command.Flags.JSON, stdout, stderr)
	case ActionFree:
		freeOutput := stdout
		if command.Flags.JSON {
			freeOutput = stderr
		}
		err := Free(ctx, deps.Scanner, deps.Manager, command.Port, stdin, freeOutput)
		if command.Flags.JSON {
			if jsonErr := RenderFreeJSON(stdout, command.Port, err); jsonErr != nil && err == nil {
				err = jsonErr
			}
		}
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionKill:
		err := Kill(ctx, deps.Manager, command.PID, stdin, stdout)
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionFind:
		err := Find(ctx, deps.Scanner, deps.Manager, command.Query, stdout)
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionWatch:
		err := Watch(ctx, deps.Scanner, command.Flags.Interval, stdout)
		if err != nil && !errors.Is(err, context.Canceled) {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	default:
		err := fmt.Errorf("unsupported action %d", command.Action)
		PrintError(stderr, err)
		return ExitCode(err)
	}
}

func runList(ctx context.Context, deps Dependencies, asJSON bool, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.List(ctx)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	infos := make(map[int]model.ProcessInfo, len(ports))
	for i := range ports {
		info, infoErr := deps.Manager.Info(ctx, ports[i].PID)
		if infoErr != nil {
			PrintError(stderr, infoErr)
			continue
		}
		infos[ports[i].PID] = info
		ports[i].ProcessName = info.Name
	}
	var renderErr error
	if asJSON {
		renderErr = RenderJSON(stdout, ports, infos)
	} else {
		renderErr = RenderPorts(stdout, ports)
	}
	if renderErr != nil {
		PrintError(stderr, renderErr)
		return ExitCode(renderErr)
	}
	return ExitSuccess
}

func runPort(ctx context.Context, deps Dependencies, portNumber int, asJSON bool, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.Port(ctx, portNumber)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	if len(ports) == 0 {
		if asJSON {
			if renderErr := RenderJSON(stdout, nil, nil); renderErr != nil {
				PrintError(stderr, renderErr)
				return ExitCode(renderErr)
			}
			return ExitSuccess
		}
		_, _ = fmt.Fprintf(stdout, "Port %d is available.\n", portNumber)
		return ExitSuccess
	}
	infos := make(map[int]model.ProcessInfo, len(ports))
	for _, record := range ports {
		info, infoErr := deps.Manager.Info(ctx, record.PID)
		if infoErr != nil {
			_ = RenderPorts(stdout, []model.PortInfo{record})
			PrintError(stderr, infoErr)
			return ExitCode(infoErr)
		}
		infos[record.PID] = info
	}
	if asJSON {
		if renderErr := RenderJSON(stdout, ports, infos); renderErr != nil {
			PrintError(stderr, renderErr)
			return ExitCode(renderErr)
		}
	} else {
		for _, record := range ports {
			if renderErr := RenderProcess(stdout, infos[record.PID], record); renderErr != nil {
				PrintError(stderr, renderErr)
				return ExitCode(renderErr)
			}
		}
	}
	return ExitSuccess
}
