//go:build darwin

package port

import (
	"bufio"
	"context"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

type DarwinScanner struct{}

func NewScanner() Scanner { return DarwinScanner{} }

func (s DarwinScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	return s.ListProtocol(ctx, "tcp")
}

// ListProtocol serves the --protocol flag: tcp, udp or all. UDP rows carry
// the same BOUND state the Windows and Linux scanners report.
func (s DarwinScanner) ListProtocol(ctx context.Context, protocol string) ([]model.PortInfo, error) {
	switch protocol {
	case "tcp":
		rows, err := runLsof(ctx, "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn")
		if err != nil {
			return nil, err
		}
		return parseLsofRows(rows, "TCP", "LISTENING"), nil
	case "udp":
		rows, err := runLsof(ctx, "-nP", "-iUDP", "-F", "pcn")
		if err != nil {
			return nil, err
		}
		return parseLsofRows(rows, "UDP", "BOUND"), nil
	case "all":
		tcp, err := s.ListProtocol(ctx, "tcp")
		if err != nil {
			return nil, err
		}
		udp, err := s.ListProtocol(ctx, "udp")
		if err != nil {
			return nil, err
		}
		return append(tcp, udp...), nil
	default:
		return nil, ErrUnsupported
	}
}

func (s DarwinScanner) PortProtocol(ctx context.Context, number int, protocol string) ([]model.PortInfo, error) {
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

func (s DarwinScanner) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
	return s.PortProtocol(ctx, number, "tcp")
}

func runLsof(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "lsof", args...).Output()
}

// parseLsofRows reads lsof -F pcn output: p<pid>, c<command> and
// n<local->remote> records grouped per file descriptor.
func parseLsofRows(output []byte, protocol, state string) []model.PortInfo {
	var rows []model.PortInfo
	var pid int
	var name string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
		case 'c':
			name = line[1:]
		case 'n':
			parts := strings.Split(line[1:], "->")
			endpoint := parts[0]
			index := strings.LastIndex(endpoint, ":")
			if index < 0 {
				continue
			}
			port, parseErr := strconv.Atoi(endpoint[index+1:])
			if parseErr != nil {
				continue
			}
			rows = append(rows, model.PortInfo{Port: port, Protocol: protocol, LocalAddr: endpoint[:index], State: state, PID: pid, ProcessName: name})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Port < rows[j].Port })
	return rows
}
