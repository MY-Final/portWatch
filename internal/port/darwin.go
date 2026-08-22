//go:build darwin

package port

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/portwatch/portwatch/pkg/model"
)

type DarwinScanner struct{}

func NewScanner() Scanner { return DarwinScanner{} }

func (DarwinScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
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
			rows = append(rows, model.PortInfo{Port: port, Protocol: "TCP", LocalAddr: endpoint[:index], State: "LISTENING", PID: pid, ProcessName: name})
		}
	}
	return rows, scanner.Err()
}

func (s DarwinScanner) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
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
