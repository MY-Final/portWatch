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
			}
		}
		return m, nil
	}
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, m.refresh()
	case "/":
		m.Filtering = true
		m.Filter = ""
	case "up":
		if m.Selected > 0 {
			m.Selected--
		}
	case "down":
		if m.Selected < len(m.Ports)-1 {
			m.Selected++
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
				return killDoneMsg{}
			}
		}
	}
	return m, nil
}

func display(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "-"
}
