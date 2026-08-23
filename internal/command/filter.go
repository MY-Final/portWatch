package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

// QueryFilter contains the user-selectable filters shared by port queries.
// Process matching is applied after process metadata has been resolved.
type QueryFilter struct {
	Process string
	PIDs    map[int]struct{}
	State   string
}

func parsePIDSet(value string) (map[int]struct{}, error) {
	parts := strings.Split(value, ",")
	pids := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		pid, err := strconv.Atoi(part)
		if err != nil || pid <= 0 {
			return nil, fmt.Errorf("--pid must be a comma-separated list of positive numbers")
		}
		pids[pid] = struct{}{}
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("--pid must be a comma-separated list of positive numbers")
	}
	return pids, nil
}

func supportedState(state string) bool {
	switch state {
	case "LISTENING", "BOUND", "ESTABLISHED", "SYN_SENT", "SYN_RECEIVED", "FIN_WAIT_1", "FIN_WAIT_2", "TIME_WAIT", "CLOSE_WAIT", "CLOSING", "LAST_ACK", "CLOSED":
		return true
	default:
		return false
	}
}

func (f QueryFilter) Empty() bool {
	return strings.TrimSpace(f.Process) == "" && len(f.PIDs) == 0 && f.State == ""
}

func (f QueryFilter) matchesPort(record model.PortInfo) bool {
	if f.State != "" && !strings.EqualFold(strings.TrimSpace(record.State), f.State) {
		return false
	}
	if len(f.PIDs) > 0 {
		if _, ok := f.PIDs[record.PID]; !ok {
			return false
		}
	}
	return true
}

func (f QueryFilter) matchesProcess(info model.ProcessInfo, record model.PortInfo) bool {
	if !f.matchesPort(record) {
		return false
	}
	if f.Process == "" {
		return true
	}
	return strings.Contains(strings.ToLower(info.Name), strings.ToLower(f.Process)) ||
		strings.Contains(strings.ToLower(record.ProcessName), strings.ToLower(f.Process))
}

func filtersAllowed(action Action) bool {
	switch action {
	case ActionList, ActionPort, ActionPortRange, ActionPortSet, ActionWatch:
		return true
	default:
		return false
	}
}

func filterPorts(ctx context.Context, manager ProcessManager, records []model.PortInfo, filter QueryFilter) ([]model.PortInfo, map[int]model.ProcessInfo, map[int]error) {
	selected := make([]model.PortInfo, 0, len(records))
	for _, record := range records {
		if filter.matchesPort(record) {
			selected = append(selected, record)
		}
	}
	infos, infoErrors := resolveProcessInfos(ctx, manager, selected)
	applyProcessNames(selected, infos)
	if filter.Process == "" {
		return selected, infos, infoErrors
	}
	filtered := make([]model.PortInfo, 0, len(selected))
	for _, record := range selected {
		if info, ok := infos[record.PID]; ok {
			if filter.matchesProcess(info, record) {
				filtered = append(filtered, record)
			}
			continue
		}
		if filter.matchesProcess(model.ProcessInfo{}, record) {
			filtered = append(filtered, record)
		}
	}
	return filtered, infos, infoErrors
}
