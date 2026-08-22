package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, m.refresh()
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
			m.Detail = fmt.Sprintf("Port %d\nPID %d\nProcess %s", record.Port, record.PID, record.ProcessName)
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
					return portsFailedMsg{err: errors.New("process manager is nil")}
				}
				if err := m.Manager.Terminate(context.Background(), pid); err != nil {
					return portsFailedMsg{err: err}
				}
				return portsLoadedMsg{ports: m.Ports}
			}
		}
	}
	return m, nil
}
