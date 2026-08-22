package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
		if len(m.Ports) > 0 {
			record := m.Ports[m.Selected]
			info := m.Infos[record.PID]
			m.Detail = fmt.Sprintf("Port %d\nProtocol %s\nState %s\nPID %d\nProcess %s\nCommand %s\nExecutable %s", record.Port, record.Protocol, record.State, record.PID, display(info.Name, record.ProcessName), display(info.Command, "-"), display(info.Executable, "-"))
		}
	case "k":
		if len(m.Ports) > 0 {
			m.ConfirmKill = true
		}
	case "n":
		m.ConfirmKill = false
	case "y":
		if m.ConfirmKill && len(m.Ports) > 0 {
			pid := m.Ports[m.Selected].PID
			m.ConfirmKill = false
			return m, func() tea.Msg {
				if m.Manager == nil {
					return killDoneMsg{err: errors.New("process manager is nil")}
				}
				if err := m.Manager.Terminate(context.Background(), pid); err != nil {
					return killDoneMsg{err: err}
				}
				exists, err := m.Manager.Exists(context.Background(), pid)
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

func display(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "-"
}
