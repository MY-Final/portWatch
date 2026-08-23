package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/processinfo"
	"github.com/MY-Final/portWatch/internal/service"
	"github.com/MY-Final/portWatch/internal/tui"
	"github.com/MY-Final/portWatch/pkg/model"
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
	ActionPortRange
	ActionPortSet
	ActionFree
	ActionKill
	ActionFind
	ActionWatch
	ActionTUI
	ActionInfo
	ActionHelp
	ActionVersion
	ActionUninstall
	ActionWait
)

// Command is the parsed root command. Port is set for port-oriented actions,
// including an optional port passed to the TUI.
type Command struct {
	Action  Action
	Port    int
	PortEnd int
	Ports   []int
	PID     int
	Query   string
	Flags   flagOptions
}

// Parse parses root arguments without executing any platform work.
// An empty argument list means list all listening ports. A single decimal
// argument means inspect that port. Known subcommands return their own action;
// malformed or unknown arguments return a structured ParseError.
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
	if _, filterErr := options.queryFilter(); filterErr != nil {
		return Command{}, &ParseError{Kind: ParseErrorInvalidFilter, Argument: filterErr.Error(), Message: filterErr.Error()}
	}
	if len(positional) == 0 {
		if options.PortsSet {
			ports, setErr := parsePortSet(options.Ports)
			if setErr != nil {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: options.Ports, Message: setErr.Error()}
			}
			return Command{Action: ActionPortSet, Ports: ports, Flags: options}, nil
		}
		return Command{Action: ActionList, Flags: options}, nil
	}
	if options.PortsSet {
		return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: options.Ports, Message: "--ports cannot be combined with positional arguments"}
	}
	if len(positional) >= 2 && positional[0] == "find" {
		query := strings.TrimSpace(strings.Join(positional[1:], " "))
		if query != "" {
			return Command{Action: ActionFind, Query: query, Flags: options}, nil
		}
	}
	if len(positional) >= 2 && positional[0] == "uninstall" {
		return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: positional[1], Message: "uninstall does not take arguments"}
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
		case "wait":
			return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: "wait requires a port"}
		case "tui":
			return Command{Action: ActionTUI, Flags: options}, nil
		case "uninstall":
			return Command{Action: ActionUninstall, Flags: options}, nil
		case "info":
			return Command{}, &ParseError{Kind: ParseErrorInvalidPID, Argument: arg, Message: "info requires a pid"}
		case "free":
			return Command{}, &ParseError{Kind: ParseErrorFree, Argument: arg, Message: "free requires a port"}
		}
		if strings.Contains(arg, "-") {
			start, end, rangeErr := parsePortRange(arg)
			if rangeErr != nil {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: rangeErr.Error()}
			}
			return Command{Action: ActionPortRange, Port: start, PortEnd: end, Flags: options}, nil
		}
		port, err := strconv.Atoi(arg)
		if err != nil || port < 1 || port > 65535 {
			return Command{}, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: "port must be a number from 1 to 65535"}
		}
		return Command{Action: ActionPort, Port: port, Flags: options}, nil
	case 2:
		if positional[0] == "tui" {
			port, parseErr := parsePortArg(positional[1])
			if parseErr != nil {
				return Command{}, parseErr
			}
			return Command{Action: ActionTUI, Port: port, Flags: options}, nil
		}
		if positional[0] == "watch" {
			port, parseErr := parsePortArg(positional[1])
			if parseErr != nil {
				return Command{}, parseErr
			}
			return Command{Action: ActionWatch, Port: port, Flags: options}, nil
		}
		if positional[0] == "wait" {
			port, parseErr := parsePortArg(positional[1])
			if parseErr != nil {
				return Command{}, parseErr
			}
			if options.Expect != "free" && options.Expect != "occupied" {
				return Command{}, &ParseError{Kind: ParseErrorInvalidFilter, Argument: options.Expect, Message: "--expect must be free or occupied"}
			}
			if options.Timeout < 0 {
				return Command{}, &ParseError{Kind: ParseErrorInvalidFilter, Argument: options.Timeout.String(), Message: "--timeout must not be negative"}
			}
			return Command{Action: ActionWait, Port: port, Flags: options}, nil
		}
		if positional[0] == "free" {
			port, parseErr := parsePortArg(positional[1])
			if parseErr != nil {
				return Command{}, parseErr
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
		if positional[0] == "info" {
			pid, err := strconv.Atoi(positional[1])
			if err != nil || pid <= 0 {
				return Command{}, &ParseError{Kind: ParseErrorInvalidPID, Argument: positional[1], Message: "pid must be a positive number"}
			}
			return Command{Action: ActionInfo, PID: pid, Flags: options}, nil
		}
	}
	return Command{}, &ParseError{Kind: ParseErrorUnknownCommand, Argument: positional[0], Message: "unknown command or arguments"}
}

// parsePortArg validates the positional port argument shared by the tui,
// watch and free subcommands.
func parsePortArg(arg string) (int, *ParseError) {
	port, err := strconv.Atoi(arg)
	if err != nil || port < 1 || port > 65535 {
		return 0, &ParseError{Kind: ParseErrorInvalidPort, Argument: arg, Message: "port must be a number from 1 to 65535"}
	}
	return port, nil
}

// ParseErrorKind identifies the reason a root argument list was rejected.
type ParseErrorKind uint8

const (
	ParseErrorInvalidPort ParseErrorKind = iota + 1
	ParseErrorUnknownCommand
	ParseErrorHelp
	ParseErrorFree
	ParseErrorInvalidPID
	ParseErrorInvalidFilter
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
			_, _ = fmt.Fprintf(stderr, "%s: %s\nusage: %s [flags] [port]\n       %s tui [port]\n", BinaryName, parseErr, BinaryName, BinaryName)
		} else {
			_, _ = fmt.Fprintf(stderr, "%s: %s\n", BinaryName, err)
		}
		return ExitCode(err)
	}
	if command.Action == ActionHelp {
		_, _ = fmt.Fprintln(stdout, "PortWatch - cross-platform port diagnostics")
		usage := []string{
			"[flags] [port]",
			"free <port>",
			"kill <pid>",
			"info <pid>",
			"find <name>",
			"<start-end>",
			"watch",
			"wait <port>",
			"tui [port]",
			"uninstall",
		}
		_, _ = fmt.Fprintf(stdout, "Usage: %s %s\n", BinaryName, usage[0])
		for _, line := range usage[1:] {
			_, _ = fmt.Fprintf(stdout, "       %s %s\n", BinaryName, line)
		}
		_, _ = fmt.Fprintln(stdout, "Flags: --json --yes --ports <p1,p2> --pid <p1,p2> --process <name> --state <state> --interval <duration> --expect free|occupied --timeout <duration> --protocol tcp")
		return ExitSuccess
	}
	if command.Action == ActionVersion {
		_, _ = fmt.Fprintln(stdout, Version)
		return ExitSuccess
	}
	queryFilter, filterErr := command.Flags.queryFilter()
	if filterErr != nil {
		parseErr := &ParseError{Kind: ParseErrorInvalidFilter, Message: filterErr.Error()}
		PrintError(stderr, parseErr)
		return ExitCode(parseErr)
	}
	if !queryFilter.Empty() && !filtersAllowed(command.Action) {
		parseErr := &ParseError{Kind: ParseErrorInvalidFilter, Message: "query filters are only supported for port queries and watch"}
		PrintError(stderr, parseErr)
		return ExitCode(parseErr)
	}
	if command.Action == ActionUninstall {
		err := Uninstall(command.Flags.Yes, stdin, stdout)
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	}
	if command.Action == ActionTUI && command.Flags.Protocol != "tcp" {
		parseErr := &ParseError{Kind: ParseErrorInvalidFilter, Message: "tui currently supports TCP listening ports only"}
		PrintError(stderr, parseErr)
		return ExitCode(parseErr)
	}
	if deps.Scanner == nil || deps.Manager == nil {
		err := errors.New("portwatch dependencies are not configured")
		PrintError(stderr, err)
		return ExitCode(err)
	}
	if command.Action != ActionKill {
		activeScanner, scannerErr := scannerForProtocol(deps.Scanner, command.Flags.Protocol)
		if scannerErr != nil {
			PrintError(stderr, scannerErr)
			return ExitCode(scannerErr)
		}
		deps.Scanner = activeScanner
	}

	switch command.Action {
	case ActionList:
		return runList(ctx, deps, command.Flags.JSON, queryFilter, stdout, stderr)
	case ActionPort:
		return runPort(ctx, deps, command.Port, command.Flags.JSON, queryFilter, stdout, stderr)
	case ActionPortRange:
		return runPortRange(ctx, deps, command.Port, command.PortEnd, command.Flags.JSON, queryFilter, stdout, stderr)
	case ActionPortSet:
		return runPortSet(ctx, deps, command.Ports, command.Flags.JSON, queryFilter, stdout, stderr)
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
		var err error
		if command.Flags.JSON {
			err = FindJSON(ctx, deps.Scanner, deps.Manager, command.Query, stdout)
		} else {
			err = Find(ctx, deps.Scanner, deps.Manager, command.Query, stdout)
		}
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionWatch:
		var err error
		if command.Flags.JSON {
			err = WatchJSONWithFilter(ctx, deps.Scanner, deps.Manager, command.Flags.Interval, command.Port, queryFilter, stdout, stderr)
		} else {
			err = WatchWithFilter(ctx, deps.Scanner, deps.Manager, command.Flags.Interval, command.Port, queryFilter, stdout, stderr)
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionWait:
		err := Wait(ctx, deps.Scanner, command.Port, command.Flags, stdout)
		if err != nil && !errors.Is(err, context.Canceled) {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionInfo:
		var err error
		if command.Flags.JSON {
			err = InfoJSON(ctx, deps.Scanner, deps.Manager, command.PID, stdout)
		} else {
			err = Info(ctx, deps.Scanner, deps.Manager, command.PID, stdout)
		}
		if err != nil {
			PrintError(stderr, err)
		}
		return ExitCode(err)
	case ActionTUI:
		err := tui.RunPort(ctx, deps.Scanner, deps.Manager, command.Port, Version)
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

type protocolScannerAdapter struct {
	scanner  port.ProtocolScanner
	protocol string
}

func (s protocolScannerAdapter) List(ctx context.Context) ([]model.PortInfo, error) {
	return s.scanner.ListProtocol(ctx, s.protocol)
}

func (s protocolScannerAdapter) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
	return s.scanner.PortProtocol(ctx, number, s.protocol)
}

// protocolUnsupportedError explains that the selected --protocol value is a
// Windows-only capability and names the current platform, instead of the
// generic platform message.
type protocolUnsupportedError struct {
	protocol string
	goos     string
}

func (e protocolUnsupportedError) Error() string {
	return fmt.Sprintf("operation is not supported on this platform: only Windows supports --protocol %s; current platform is %s", e.protocol, e.goos)
}

// Unwrap keeps exit-code mapping and errors.Is checks aligned with the plain
// port.ErrUnsupported sentinel.
func (e protocolUnsupportedError) Unwrap() error { return port.ErrUnsupported }

func scannerForProtocol(scanner PortScanner, protocol string) (PortScanner, error) {
	if protocol == "tcp" {
		return scanner, nil
	}
	protocolScanner, ok := scanner.(port.ProtocolScanner)
	if !ok {
		return nil, protocolUnsupportedError{protocol: protocol, goos: runtime.GOOS}
	}
	return protocolScannerAdapter{scanner: protocolScanner, protocol: protocol}, nil
}

func parsePortRange(value string) (int, int, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, fmt.Errorf("port range must use START-END with values from 1 to 65535")
	}
	start, startErr := strconv.Atoi(parts[0])
	end, endErr := strconv.Atoi(parts[1])
	if startErr != nil || endErr != nil || start < 1 || end > 65535 || start > end {
		return 0, 0, fmt.Errorf("port range must use START-END with values from 1 to 65535")
	}
	return start, end, nil
}

func parsePortSet(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	seen := make(map[int]struct{}, len(parts))
	ports := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		port, err := strconv.Atoi(part)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("ports must be a comma-separated list from 1 to 65535")
		}
		if _, exists := seen[port]; exists {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("ports must be a comma-separated list from 1 to 65535")
	}
	sort.Ints(ports)
	return ports, nil
}

func runList(ctx context.Context, deps Dependencies, asJSON bool, filter QueryFilter, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.List(ctx)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	ports, infos, infoErrors := filterPorts(ctx, deps.Manager, ports, filter)
	reportProcessInfoErrors(stderr, infoErrors)
	var renderErr error
	if asJSON {
		renderErr = RenderJSONWithServices(stdout, ports, infos, service.Rules{})
	} else {
		renderErr = RenderPortsWithServices(stdout, ports, infos, service.Rules{})
	}
	if renderErr != nil {
		PrintError(stderr, renderErr)
		return ExitCode(renderErr)
	}
	return ExitSuccess
}

func runPort(ctx context.Context, deps Dependencies, portNumber int, asJSON bool, filter QueryFilter, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.Port(ctx, portNumber)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	if len(ports) == 0 {
		return renderEmptyPortResult(stdout, stderr, portNumber, asJSON)
	}
	ports, infos, infoErrors := filterPorts(ctx, deps.Manager, ports, filter)
	if len(infoErrors) > 0 {
		for _, infoErr := range infoErrors {
			if asJSON {
				_ = RenderJSONWithServices(stdout, ports, infos, service.Rules{})
			} else if len(ports) > 0 {
				_ = RenderPorts(stdout, ports)
			}
			PrintError(stderr, infoErr)
			return ExitCode(infoErr)
		}
	}
	if len(ports) == 0 {
		return renderEmptyPortResult(stdout, stderr, portNumber, asJSON)
	}
	if asJSON {
		if renderErr := RenderJSONWithServices(stdout, ports, infos, service.Rules{}); renderErr != nil {
			PrintError(stderr, renderErr)
			return ExitCode(renderErr)
		}
	} else {
		for _, record := range ports {
			if renderErr := RenderProcessWithService(stdout, infos[record.PID], record, service.Rules{}); renderErr != nil {
				PrintError(stderr, renderErr)
				return ExitCode(renderErr)
			}
		}
	}
	return ExitSuccess
}

// renderEmptyPortResult reports an unoccupied port identically for the text
// and JSON outputs, both before and after process filtering.
func renderEmptyPortResult(stdout, stderr io.Writer, portNumber int, asJSON bool) int {
	if asJSON {
		if renderErr := RenderJSONWithServices(stdout, nil, nil, service.Rules{}); renderErr != nil {
			PrintError(stderr, renderErr)
			return ExitCode(renderErr)
		}
		return ExitSuccess
	}
	_, _ = fmt.Fprintf(stdout, "Port %d is available.\n", portNumber)
	return ExitSuccess
}

func runPortRange(ctx context.Context, deps Dependencies, start, end int, asJSON bool, filter QueryFilter, stdout, stderr io.Writer) int {
	ports, err := deps.Scanner.List(ctx)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	filtered := make([]model.PortInfo, 0)
	for _, record := range ports {
		if record.Port >= start && record.Port <= end {
			filtered = append(filtered, record)
		}
	}
	filtered, infos, infoErrors := filterPorts(ctx, deps.Manager, filtered, filter)
	reportProcessInfoErrors(stderr, infoErrors)
	if asJSON {
		if err := RenderJSONWithServices(stdout, filtered, infos, service.Rules{}); err != nil {
			PrintError(stderr, err)
			return ExitCode(err)
		}
		return ExitSuccess
	}
	if err := RenderPortsWithServices(stdout, filtered, infos, service.Rules{}); err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	return ExitSuccess
}

func runPortSet(ctx context.Context, deps Dependencies, requested []int, asJSON bool, filter QueryFilter, stdout, stderr io.Writer) int {
	wanted := make(map[int]struct{}, len(requested))
	for _, port := range requested {
		wanted[port] = struct{}{}
	}
	ports, err := deps.Scanner.List(ctx)
	if err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	filtered := make([]model.PortInfo, 0)
	for _, record := range ports {
		if _, ok := wanted[record.Port]; ok {
			filtered = append(filtered, record)
		}
	}
	filtered, infos, infoErrors := filterPorts(ctx, deps.Manager, filtered, filter)
	reportProcessInfoErrors(stderr, infoErrors)
	if asJSON {
		if err := RenderJSONWithServices(stdout, filtered, infos, service.Rules{}); err != nil {
			PrintError(stderr, err)
			return ExitCode(err)
		}
		return ExitSuccess
	}
	if err := RenderPortsWithServices(stdout, filtered, infos, service.Rules{}); err != nil {
		PrintError(stderr, err)
		return ExitCode(err)
	}
	return ExitSuccess
}

// resolveProcessInfos looks up process metadata once per unique PID using a
// bounded worker pool (at most 8 concurrent Info calls). Cancellation stops
// dispatching new lookups; already-started calls are left to their context.
func resolveProcessInfos(ctx context.Context, manager ProcessManager, records []model.PortInfo) (map[int]model.ProcessInfo, map[int]error) {
	return processinfo.Resolve(ctx, manager, records)
}

func applyProcessNames(records []model.PortInfo, infos map[int]model.ProcessInfo) {
	processinfo.ApplyNames(records, infos)
}

func reportProcessInfoErrors(stderr io.Writer, errorsByPID map[int]error) {
	if len(errorsByPID) == 0 {
		return
	}
	// Map iteration order is randomized; pick the lowest PID's error so the
	// same input always produces the same stderr line.
	pids := make([]int, 0, len(errorsByPID))
	for pid := range errorsByPID {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	PrintError(stderr, fmt.Errorf("process information unavailable for %d PID(s): %w", len(errorsByPID), errorsByPID[pids[0]]))
}
