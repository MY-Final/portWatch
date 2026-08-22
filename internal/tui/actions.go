package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/portwatch/portwatch/pkg/model"
)

func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	if m.Filtering {
		switch key.String() {
		case "esc":
			m.Filtering = false
			m.Filter = ""
		case "enter":
			m.Filtering = false
		case "backspace":
			if len(m.Filter) > 0 {
				runes := []rune(m.Filter)
				m.Filter = string(runes[:len(runes)-1])
			}
		default:
			if key.Type == tea.KeyRunes {
				m.Filter += string(key.Runes)
				m.normalizeSelection()
			}
		}
		return m, nil
	}
	m.normalizeSelection()
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, m.refresh()
	case "/":
		m.Filtering = true
		m.Filter = ""
	case "up":
		indexes := m.visibleIndexes()
		for position, index := range indexes {
			if index == m.Selected && position > 0 {
				m.Selected = indexes[position-1]
				break
			}
		}
	case "down":
		indexes := m.visibleIndexes()
		for position, index := range indexes {
			if index == m.Selected && position+1 < len(indexes) {
				m.Selected = indexes[position+1]
				break
			}
		}
	case "enter":
		if record, ok := m.selectedRecord(); ok {
			info := m.Infos[record.PID]
			m.Detail = fmt.Sprintf("Port %d\nProtocol %s\nState %s\nPID %d\nProcess %s\nCommand %s\nExecutable %s", record.Port, record.Protocol, record.State, record.PID, display(info.Name, record.ProcessName), display(info.Command, "-"), display(info.Executable, "-"))
		}
	case "k":
		if _, ok := m.selectedRecord(); ok {
			m.ConfirmKill = true
		}
	case "n":
		m.ConfirmKill = false
	case "y":
		if m.ConfirmKill {
			record, ok := m.selectedRecord()
			if !ok {
				m.ConfirmKill = false
				return m, nil
			}
			pid := record.PID
			m.ConfirmKill = false
			return m, func() tea.Msg {
				ctx := m.Context
				if ctx == nil {
					ctx = context.Background()
				}
				if m.Manager == nil {
					return killDoneMsg{err: errors.New("process manager is nil")}
				}
				if err := m.Manager.Terminate(ctx, pid); err != nil {
					return killDoneMsg{err: err}
				}
				exists, err := m.Manager.Exists(ctx, pid)
				if err != nil {
					return killDoneMsg{err: fmt.Errorf("verify pid %d termination: %w", pid, err)}
				}
				if exists {
					return killDoneMsg{err: fmt.Errorf("pid %d still exists after termination", pid)}
				}
				return killDoneMsg{}
			}
		}
	}
	return m, nil
}

func (m *Model) normalizeSelection() {
	indexes := m.visibleIndexes()
	if len(indexes) == 0 {
		return
	}
	for _, index := range indexes {
		if index == m.Selected {
			return
		}
	}
	m.Selected = indexes[0]
}

func (m Model) selectedRecord() (model.PortInfo, bool) {
	if m.Selected < 0 || m.Selected >= len(m.Ports) {
		return model.PortInfo{}, false
	}
	for _, index := range m.visibleIndexes() {
		if index == m.Selected {
			return m.Ports[index], true
		}
	}
	return model.PortInfo{}, false
}

func display(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "-"
}
