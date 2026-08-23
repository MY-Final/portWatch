package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MY-Final/portWatch/pkg/model"
)

func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	if m.Filtering {
		return m.handleFilterKey(key)
	}
	if key.String() == "q" || key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.Page == pageConfirm {
		return m.handleConfirmKey(key)
	}
	if m.Page == pageHelp {
		if key.String() == "esc" || key.String() == "?" {
			m.Page = m.HelpReturn
		}
		return m, nil
	}
	if m.Page == pageView {
		return m.handleViewKey(key)
	}
	if m.Page == pageDetails {
		return m.handleDetailsKey(key)
	}
	return m.handleListKey(key)
}

func (m Model) handleListKey(key tea.KeyMsg) (Model, tea.Cmd) {
	// Use the typed key in addition to String() because Windows console input
	// adapters can normalize arrow sequences differently across terminals.
	if key.Type == tea.KeyUp {
		m.moveSelection(-1)
		return m, nil
	}
	if key.Type == tea.KeyDown {
		m.moveSelection(1)
		return m, nil
	}
	switch key.String() {
	case "u":
		m.moveSelection(-1)
	case "j":
		m.moveSelection(1)
	case "r":
		m.Status = "Refreshing..."
		return m, m.refresh()
	case "/":
		m.Filtering = true
		return m, nil
	case "enter":
		if record, ok := m.selectedRecord(); ok {
			m.DetailRecord = record
			m.Detail = "Process details"
			m.Page = pageDetails
		}
	case "k":
		if record, ok := m.selectedRecord(); ok {
			m.DetailRecord = record
			m.ConfirmReturn = pageList
			m.Page = pageConfirm
			m.ConfirmKill = true
		}
	case "?":
		m.HelpReturn = pageList
		m.Page = pageHelp
	case "v", "V":
		m.ViewSelection = m.Scope
		m.Page = pageView
	}
	return m, nil
}

func (m Model) handleViewKey(key tea.KeyMsg) (Model, tea.Cmd) {
	if key.String() == "esc" {
		m.Page = pageList
		return m, nil
	}
	scope, ok := modeKey(key.String())
	if !ok {
		return m, nil
	}
	if scope == m.Scope {
		m.Page = pageList
		return m, nil
	}
	m.Scope = scope
	m.ViewSelection = scope
	m.Page = pageList
	m.Status = "Loading " + scopeLabel(scope) + "..."
	return m, m.refresh()
}

func (m Model) handleDetailsKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.Page = pageList
		m.Detail = ""
		m.ConfirmKill = false
	case "k":
		if _, ok := m.selectedRecord(); ok {
			m.ConfirmReturn = pageDetails
			m.Page = pageConfirm
			m.ConfirmKill = true
		}
	case "r":
		m.Status = "Refreshing..."
		return m, m.refresh()
	case "?":
		m.HelpReturn = pageDetails
		m.Page = pageHelp
	}
	return m, nil
}

func (m Model) handleConfirmKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc", "n":
		m.Page = m.ConfirmReturn
		m.ConfirmKill = false
	case "enter", "y":
		return m, m.killSelected()
	}
	return m, nil
}

func (m Model) handleFilterKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.Filtering = false
		m.Filter = ""
	case "enter":
		m.Filtering = false
		m.normalizeSelection()
	case "backspace":
		if len(m.Filter) > 0 {
			runes := []rune(m.Filter)
			m.Filter = string(runes[:len(runes)-1])
			m.normalizeSelection()
		}
	default:
		if key.Type == tea.KeyRunes {
			m.Filter += string(key.Runes)
			m.normalizeSelection()
		}
	}
	return m, nil
}

func (m *Model) moveSelection(delta int) {
	indexes := m.visibleIndexes()
	if len(indexes) == 0 {
		return
	}
	position := 0
	for i, index := range indexes {
		if index == m.Selected {
			position = i
			break
		}
	}
	position += delta
	if position < 0 {
		position = 0
	}
	if position >= len(indexes) {
		position = len(indexes) - 1
	}
	m.Selected = indexes[position]
	m.SelectedKey = m.currentKey()
}

func (m *Model) normalizeSelection() {
	indexes := m.visibleIndexes()
	if len(indexes) == 0 {
		return
	}
	for _, index := range indexes {
		if index == m.Selected {
			m.SelectedKey = m.currentKey()
			return
		}
	}
	m.Selected = indexes[0]
	m.SelectedKey = m.currentKey()
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

func (m Model) killSelected() tea.Cmd {
	record := m.DetailRecord
	if record.Port == 0 {
		var ok bool
		record, ok = m.selectedRecord()
		if !ok {
			return nil
		}
	}
	m.DetailRecord = record
	m.Page = pageConfirm
	m.ConfirmKill = false
	return func() tea.Msg {
		ctx := m.Context
		if ctx == nil {
			ctx = context.Background()
		}
		if m.Manager == nil {
			return killDoneMsg{err: errors.New("process manager is nil")}
		}
		if record.PID == 4 || record.PID == os.Getpid() {
			return killDoneMsg{err: fmt.Errorf("refusing to terminate protected pid %d", record.PID)}
		}
		if err := m.Manager.Terminate(ctx, record.PID); err != nil {
			return killDoneMsg{err: fmt.Errorf("terminate pid %d: %w", record.PID, err)}
		}
		exists, err := m.Manager.Exists(ctx, record.PID)
		if err != nil {
			return killDoneMsg{err: fmt.Errorf("verify pid %d termination: %w", record.PID, err)}
		}
		if exists {
			return killDoneMsg{err: fmt.Errorf("pid %d still exists after termination", record.PID)}
		}
		if m.Scanner != nil {
			remaining, scanErr := m.Scanner.Port(ctx, record.Port)
			if scanErr != nil {
				return killDoneMsg{err: fmt.Errorf("verify port %d release: %w", record.Port, scanErr)}
			}
			for _, current := range remaining {
				if current.PID == record.PID && strings.EqualFold(current.Protocol, record.Protocol) {
					return killDoneMsg{err: fmt.Errorf("port %d is still occupied by pid %d", record.Port, record.PID)}
				}
			}
		}
		return killDoneMsg{}
	}
}
