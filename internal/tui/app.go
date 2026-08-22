package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
	"github.com/portwatch/portwatch/pkg/model"
)

type Model struct {
	Scanner     port.Scanner
	Manager     process.Manager
	Ports       []model.PortInfo
	Infos       map[int]model.ProcessInfo
	Filter      string
	Filtering   bool
	Err         error
	Selected    int
	Detail      string
	ConfirmKill bool
}

type portsLoadedMsg struct {
	ports []model.PortInfo
	infos map[int]model.ProcessInfo
}
type portsFailedMsg struct{ err error }
type killDoneMsg struct{ err error }

func New(scanner port.Scanner, manager process.Manager) Model {
	return Model{Scanner: scanner, Manager: manager}
}

func (m Model) Init() tea.Cmd { return m.refresh() }

func (m Model) refresh() tea.Cmd {
	return func() tea.Msg {
		ports, err := m.Scanner.List(context.Background())
		if err != nil {
			return portsFailedMsg{err: err}
		}
		infos := make(map[int]model.ProcessInfo, len(ports))
		for i := range ports {
			if m.Manager == nil || ports[i].PID <= 0 {
				continue
			}
			info, infoErr := m.Manager.Info(context.Background(), ports[i].PID)
			if infoErr != nil {
				continue
			}
			infos[ports[i].PID] = info
			ports[i].ProcessName = info.Name
		}
		return portsLoadedMsg{ports: ports, infos: infos}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(value)
	case portsLoadedMsg:
		m.Ports, m.Infos, m.Err = value.ports, value.infos, nil
		if m.Selected >= len(m.Ports) {
			m.Selected = len(m.Ports) - 1
		}
		if m.Selected < 0 {
			m.Selected = 0
		}
		sort.Slice(m.Ports, func(i, j int) bool { return m.Ports[i].Port < m.Ports[j].Port })
	case portsFailedMsg:
		m.Err = value.err
	case killDoneMsg:
		if value.err != nil {
			m.Err = value.err
			m.ConfirmKill = false
			return m, nil
		}
		m.Detail = "Process terminated."
		return m, m.refresh()
	}
	return m, nil
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("PortWatch\n\nError: %v\n\nR refresh   Q quit\n", m.Err)
	}
	var b strings.Builder
	b.WriteString("PortWatch\n\nPORT   PROTOCOL   PID      PROCESS\n")
	if m.Filtering {
		fmt.Fprintf(&b, "Search: %s_\n", m.Filter)
	} else if m.Filter != "" {
		fmt.Fprintf(&b, "Filter: %s\n", m.Filter)
	}
	if m.ConfirmKill {
		b.WriteString("Kill selected process? Y/N\n")
	}
	if m.Detail != "" {
		b.WriteString(m.Detail + "\n")
	}
	for index, record := range m.Ports {
		if m.Filter != "" && !strings.Contains(strings.ToLower(record.ProcessName), strings.ToLower(m.Filter)) {
			continue
		}
		marker := " "
		if index == m.Selected {
			marker = ">"
		}
		fmt.Fprintf(&b, "%s%-5d %-10s %-8d %s\n", marker, record.Port, record.Protocol, record.PID, displayProcessName(record.ProcessName))
	}
	b.WriteString("\nR refresh   / filter   Q quit\n")
	return b.String()
}

func displayProcessName(name string) string {
	if name == "" {
		return "-"
	}
	return name
}

func Run(ctx context.Context, scanner port.Scanner, manager process.Manager) error {
	program := tea.NewProgram(New(scanner, manager), tea.WithContext(ctx))
	_, err := program.Run()
	return err
}
