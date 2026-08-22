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

	"github.com/portwatch/portwatch/pkg/model"
)

type LinuxScanner struct{}

func NewScanner() Scanner { return LinuxScanner{} }

func (LinuxScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := make([]model.PortInfo, 0)
	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		parsed, err := parseProcTCP(path)
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

func parseProcTCP(path string) ([]model.PortInfo, error) {
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
		rows = append(rows, model.PortInfo{Port: port, Protocol: "TCP", LocalAddr: address, State: "LISTENING", PID: findSocketPID(fields[9])})
	}
	return rows, scanner.Err()
}

func findSocketPID(inode string) int {
	target := "socket:[" + inode + "]"
	entries, _ := os.ReadDir("/proc")
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		fds, _ := filepath.Glob(filepath.Join("/proc", entry.Name(), "fd", "*"))
		for _, fd := range fds {
			link, err := os.Readlink(fd)
			if err == nil && link == target {
				pid, _ := strconv.Atoi(entry.Name())
				return pid
			}
		}
	}
	return 0
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
	if len(bytes) == 8 {
		return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String(), int(port64), nil
	}
	if len(bytes) == 32 {
		return net.IP(bytes).String(), int(port64), nil
	}
	return "", 0, errors.New("invalid address length")
}
