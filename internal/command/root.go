package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

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
)

// Command is the parsed root command. Port is set for ActionPort and
// ActionFree; ActionList represents an invocation with no arguments.
type Command struct {
	Action Action
	Port   int
}

// Parse parses the MVP root arguments without executing any platform work.
// An empty argument list means list all listening ports. A single decimal
// argument means inspect that port. Other commands are reserved for later
// wiring and return a structured ParseError.
func Parse(args []string) (Command, error) {
	switch len(args) {
	case 0:
		return Command{Action: ActionList}, nil
	case 1:
		arg := args[0]
		switch arg {
		case "help", "-h", "--help":
			return Command{}, &ParseError{Kind: ParseErrorHelp, Argument: arg, Message: "help is not available yet"}
		case "free":
			return Command{}, &ParseError{Kind: ParseErrorFree, Argument: arg, Message: "free requires a port and is not available yet"}
		}
		port, err := strconv.Atoi(arg)
		if err != nil || port < 1 || port > 65535 {
			return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: "port must be a number from 1 to 65535"}
		}
		return Command{Action: ActionPort, Port: port}, nil
	case 2:
		if args[0] == "free" {
			port, err := strconv.Atoi(args[1])
			if err != nil || port < 1 || port > 65535 {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: args[1], Message: "port must be a number from 1 to 65535"}
			}
			return Command{Action: ActionFree, Port: port}, nil
		}
	}
	return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: args[0], Message: "unknown command or arguments"}
}

// ParseErrorKind identifies the reason a root argument list was rejected.
type ParseErrorKind uint8

const (
	ParseErrorInvalidPort ParseErrorKind = iota + 1
	ParseErrorUnknownCommand
	ParseErrorHelp
	ParseErrorFree
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
			if parseErr.Kind == ParseErrorHelp {
				_, _ = fmt.Fprintln(stdout, "PortWatch - Windows TCP port diagnostics")
				_, _ = fmt.Fprintln(stdout, "Usage: portwatch [port] | portwatch free <port>")
				return ExitSuccess
			}
			_, _ = fmt.Fprintf(stderr, "portwatch: %s\nusage: portwatch [port]\n", parseErr)
		} else {
			_, _ = fmt.Fprintf(stderr, "portwatch: %s\n", err)
		}
		return ExitCode(err)
	}
	if deps.Scanner == nil || deps.Manager == nil {
		err := errors.New("portwatch dependencies are not configured")
		PrintError(stderr, err)
		return ExitCode(err)
	}

	switch command.Action {
	case ActionList:
		return runList(ctx, deps, stdout, stderr)
	case ActionPort:
		return runPort(ctx, deps, command.Port, stdout, stderr)
	case ActionFree:
		err := Free(ctx, deps.Scanner, deps.Manager, command.Port, stdin, stdout)
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	default:
		err := fmt.Errorf("unsupported action %d", command.Action)
		PrintError(stderr, err)
		return ExitCode(err)
	}
}

func runList(ctx context.Context, deps Dependencies, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.List(ctx)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	for i := range ports {
		info, infoErr := deps.Manager.Info(ctx, ports[i].PID)
		if infoErr != nil {
			PrintError(stderr, infoErr)
			continue
		}
		ports[i].ProcessName = info.Name
	}
	if err := RenderPorts(stdout, ports); err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	return ExitSuccess
}

func runPort(ctx context.Context, deps Dependencies, portNumber int, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.Port(ctx, portNumber)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	if len(ports) == 0 {
		_, _ = fmt.Fprintf(stdout, "Port %d is available.\n", portNumber)
		return ExitSuccess
	}
	for _, record := range ports {
		info, infoErr := deps.Manager.Info(ctx, record.PID)
		if infoErr != nil {
			_ = RenderPorts(stdout, []model.PortInfo{record})
			PrintError(stderr, infoErr)
			return ExitCode(infoErr)
		}
		if renderErr := RenderProcess(stdout, info, record); renderErr != nil {
			PrintError(stderr, renderErr)
			return ExitCode(renderErr)
		}
	}
	return ExitSuccess
}
