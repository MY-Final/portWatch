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
	Filter      string
	Err         error
	Selected    int
	Detail      string
	ConfirmKill bool
}

type portsLoadedMsg struct{ ports []model.PortInfo }
type portsFailedMsg struct{ err error }

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
		return portsLoadedMsg{ports: ports}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(value)
	case portsLoadedMsg:
		m.Ports, m.Err = value.ports, nil
		sort.Slice(m.Ports, func(i, j int) bool { return m.Ports[i].Port < m.Ports[j].Port })
	case portsFailedMsg:
		m.Err = value.err
	}
	return m, nil
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("PortWatch\n\nError: %v\n\nR refresh   Q quit\n", m.Err)
	}
	var b strings.Builder
	b.WriteString("PortWatch\n\nPORT   PROTOCOL   PID      PROCESS\n")
	if m.ConfirmKill {
		b.WriteString("Kill selected process? Y/N\n")
	}
	if m.Detail != "" {
		b.WriteString(m.Detail + "\n")
	}
	for _, record := range m.Ports {
		if m.Filter != "" && !strings.Contains(strings.ToLower(record.ProcessName), strings.ToLower(m.Filter)) {
			continue
		}
		fmt.Fprintf(&b, "%-6d %-10s %-8d %s\n", record.Port, record.Protocol, record.PID, record.ProcessName)
	}
	b.WriteString("\nR refresh   / filter   Q quit\n")
	return b.String()
}

func Run(ctx context.Context, scanner port.Scanner, manager process.Manager) error {
	program := tea.NewProgram(New(scanner, manager), tea.WithContext(ctx))
	_, err := program.Run()
	return err
}
