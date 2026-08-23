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
	return listProcNet(procRoot, os.ReadDir)
}

// listProcNet parses net/tcp and net/tcp6 under root. The socket inode map
// is built once for both tables instead of once per listening row, so the
// whole /proc fd tree is traversed a single time per call. readDir is
// injected so tests can count traversals on a fake root.
func listProcNet(root string, readDir func(string) ([]os.DirEntry, error)) ([]model.PortInfo, error) {
	inodePIDs := buildInodePIDMap(root, readDir)
	rows := make([]model.PortInfo, 0)
	for _, name := range []string{"net/tcp", "net/tcp6"} {
		parsed, err := parseProcTCP(filepath.Join(root, name), inodePIDs)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		rows = append(rows, parsed...)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Port < rows[j].Port })
	return rows, nil
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

func parseProcTCP(path string, inodePIDs map[string]int) ([]model.PortInfo, error) {
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
		if len(fields) < 10 || fields[3] != "0A" {
			continue
		}
		address, port, err := parseProcAddress(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		rows = append(rows, model.PortInfo{Port: port, Protocol: "TCP", LocalAddr: address, State: "LISTENING", PID: findSocketPID(fields[9], inodePIDs)})
	}
	return rows, scanner.Err()
}

// findSocketPID returns the PID that owns the socket inode, or 0 when the
// inode is absent from the map (unreadable fd or exited process).
func findSocketPID(inode string, inodePIDs map[string]int) int {
	return inodePIDs[inode]
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
