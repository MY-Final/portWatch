//go:build linux

package port

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

type LinuxScanner struct{}

// procRoot anchors /proc lookups so tests can inject a fake tree.
const procRoot = "/proc"

func NewScanner() Scanner { return LinuxScanner{} }

func (LinuxScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return listProcNet(procRoot, os.ReadDir, "tcp")
}

// ListProtocol serves the --protocol flag: tcp, udp or all.
func (LinuxScanner) ListProtocol(ctx context.Context, protocol string) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch protocol {
	case "tcp", "udp":
		return listProcNet(procRoot, os.ReadDir, protocol)
	case "all":
		tcp, err := listProcNet(procRoot, os.ReadDir, "tcp")
		if err != nil {
			return nil, err
		}
		udp, err := listProcNet(procRoot, os.ReadDir, "udp")
		if err != nil {
			return nil, err
		}
		return append(tcp, udp...), nil
	default:
		return nil, fmt.Errorf("%w: protocol %q", ErrUnsupported, protocol)
	}
}

func (s LinuxScanner) PortProtocol(ctx context.Context, number int, protocol string) ([]model.PortInfo, error) {
	if number < 1 || number > 65535 {
		return nil, ErrInvalidPort
	}
	rows, err := s.ListProtocol(ctx, protocol)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.PortInfo, 0)
	for _, row := range rows {
		if row.Port == number {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

// ListScope returns the TCP records needed by the interactive TUI, mirroring
// the Windows semantics: Listening keeps the CLI contract, Connections and
// All read every state from /proc/net/tcp{,6}.
func (s LinuxScanner) ListScope(ctx context.Context, scope Scope) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch scope {
	case ScopeListening:
		return s.List(ctx)
	case ScopeConnections, ScopeAll:
		rows, err := listProcNetAllStates(procRoot, os.ReadDir)
		if err != nil {
			return nil, err
		}
		if scope == ScopeConnections {
			filtered := rows[:0]
			for _, row := range rows {
				if row.State != "LISTENING" {
					filtered = append(filtered, row)
				}
			}
			rows = filtered
		}
		return rows, nil
	default:
		return nil, ErrScopeUnsupported
	}
}

func (s LinuxScanner) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
	if number < 1 || number > 65535 {
		return nil, ErrInvalidPort
	}
	rows, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	filtered := rows[:0]
	for _, row := range rows {
		if row.Port == number {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

// listProcNet parses the tcp or udp tables under root. The socket inode map
// is built once for both tables of the protocol instead of once per row, so
// the whole /proc fd tree is traversed a single time per call. readDir is
// injected so tests can count traversals on a fake root.
func listProcNet(root string, readDir func(string) ([]os.DirEntry, error), protocol string) ([]model.PortInfo, error) {
	inodePIDs := buildInodePIDMap(root, readDir)
	rows := make([]model.PortInfo, 0)
	for _, name := range procTables(protocol) {
		parsed, err := parseProcTable(filepath.Join(root, name), inodePIDs, protocol)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		rows = append(rows, parsed...)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Port < rows[j].Port })
	return rows, nil
}

// listProcNetAllStates parses the TCP tables keeping every state, for the
// TUI Connections and All views.
func listProcNetAllStates(root string, readDir func(string) ([]os.DirEntry, error)) ([]model.PortInfo, error) {
	inodePIDs := buildInodePIDMap(root, readDir)
	rows := make([]model.PortInfo, 0)
	for _, name := range []string{"net/tcp", "net/tcp6"} {
		parsed, err := parseProcTCP(filepath.Join(root, name), inodePIDs, true)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		rows = append(rows, parsed...)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Port < rows[j].Port })
	return rows, nil
}

func procTables(protocol string) []string {
	if protocol == "udp" {
		return []string{"net/udp", "net/udp6"}
	}
	return []string{"net/tcp", "net/tcp6"}
}

// parseProcTable dispatches to the per-protocol row parser; both table
// families share the same column layout.
func parseProcTable(path string, inodePIDs map[string]int, protocol string) ([]model.PortInfo, error) {
	if protocol == "udp" {
		return parseProcRows(path, inodePIDs, parseUDPRow)
	}
	return parseProcRows(path, inodePIDs, parseTCPListenerRow)
}

// parseProcTCP parses one TCP table with an explicit all-states switch.
func parseProcTCP(path string, inodePIDs map[string]int, includeAllStates bool) ([]model.PortInfo, error) {
	if includeAllStates {
		return parseProcRows(path, inodePIDs, parseAnyTCPRow)
	}
	return parseProcRows(path, inodePIDs, parseTCPListenerRow)
}

// procRowParser turns one /proc/net table line into a record; keep=false
// drops the row without an error.
type procRowParser func(fields []string, inodePIDs map[string]int) (model.PortInfo, bool, error)

func parseProcRows(path string, inodePIDs map[string]int, parse procRowParser) ([]model.PortInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var rows []model.PortInfo
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		row, keep, parseErr := parse(fields, inodePIDs)
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, parseErr)
		}
		if keep {
			rows = append(rows, row)
		}
	}
	return rows, scanner.Err()
}

// parseTCPListenerRow keeps LISTENING rows only, preserving the default
// List contract.
func parseTCPListenerRow(fields []string, inodePIDs map[string]int) (model.PortInfo, bool, error) {
	if fields[3] != "0A" {
		return model.PortInfo{}, false, nil
	}
	row, err := tcpRow(fields, inodePIDs, "LISTENING")
	if err != nil {
		return model.PortInfo{}, false, err
	}
	return row, true, nil
}

// parseAnyTCPRow keeps every row and maps the hex state code to the same
// state names the Windows scanner reports.
func parseAnyTCPRow(fields []string, inodePIDs map[string]int) (model.PortInfo, bool, error) {
	state, ok := tcpStateNames[fields[3]]
	if !ok {
		state = "UNKNOWN"
	}
	row, err := tcpRow(fields, inodePIDs, state)
	if err != nil {
		return model.PortInfo{}, false, err
	}
	return row, true, nil
}

func tcpRow(fields []string, inodePIDs map[string]int, state string) (model.PortInfo, error) {
	address, port, err := parseProcAddress(fields[1])
	if err != nil {
		return model.PortInfo{}, err
	}
	remote, _, err := parseProcAddress(fields[2])
	if err != nil {
		return model.PortInfo{}, err
	}
	return model.PortInfo{
		Port:       port,
		Protocol:   "TCP",
		LocalAddr:  address,
		RemoteAddr: remote,
		State:      state,
		PID:        inodePIDs[fields[9]],
	}, nil
}

// parseUDPRow mirrors the Windows UDP view: State is BOUND because /proc
// UDP sockets have no connection state.
func parseUDPRow(fields []string, inodePIDs map[string]int) (model.PortInfo, bool, error) {
	address, port, err := parseProcAddress(fields[1])
	if err != nil {
		return model.PortInfo{}, false, err
	}
	return model.PortInfo{
		Port:      port,
		Protocol:  "UDP",
		LocalAddr: address,
		State:     "BOUND",
		PID:       inodePIDs[fields[9]],
	}, true, nil
}

// tcpStateNames maps the /proc/net/tcp hex state codes to the state strings
// used across platforms.
var tcpStateNames = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTENING",
	"0B": "CLOSING",
	"0C": "NEW_SYN_RECV",
}

// buildInodePIDMap walks the pid directories under root once and maps every
// socket inode reachable through /proc/<pid>/fd to its owning PID. fd
// entries that cannot be listed or read (permission denied, process exited)
// are skipped silently, matching the previous per-socket lookup. When the
// same inode appears under several PIDs the first one found wins. Only
// Readlink is issued per fd entry; no stat calls are made.
func buildInodePIDMap(root string, readDir func(string) ([]os.DirEntry, error)) map[string]int {
	inodePIDs := make(map[string]int)
	entries, err := readDir(root)
	if err != nil {
		return inodePIDs
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, pidErr := strconv.Atoi(entry.Name())
		if pidErr != nil {
			continue
		}
		fdDir := filepath.Join(root, entry.Name(), "fd")
		fds, fdErr := readDir(fdDir)
		if fdErr != nil {
			continue
		}
		for _, fd := range fds {
			link, linkErr := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if linkErr != nil {
				continue
			}
			inode, ok := socketInode(link)
			if !ok {
				continue
			}
			if _, exists := inodePIDs[inode]; !exists {
				inodePIDs[inode] = pid
			}
		}
	}
	return inodePIDs
}

// socketInode extracts the inode number from an fd link such as
// "socket:[12345]".
func socketInode(link string) (string, bool) {
	const prefix = "socket:["
	if !strings.HasPrefix(link, prefix) || !strings.HasSuffix(link, "]") {
		return "", false
	}
	inode := link[len(prefix) : len(link)-1]
	if inode == "" {
		return "", false
	}
	return inode, true
}

func parseProcAddress(value string) (string, int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("invalid endpoint")
	}
	port64, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return "", 0, err
	}
	bytes, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", 0, err
	}
	// /proc renders IPv4 as 8 hex characters (little-endian, 4 decoded bytes)
	// and IPv6 as 32 characters (network order, 16 decoded bytes).
	if len(bytes) == 4 {
		return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String(), int(port64), nil
	}
	if len(bytes) == 16 {
		return net.IP(bytes).String(), int(port64), nil
	}
	return "", 0, errors.New("invalid address length")
}
