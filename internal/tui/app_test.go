package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestModelViewAndQuit(t *testing.T) {
	m := Model{Ports: []model.PortInfo{{Port: 8080, Protocol: "TCP", PID: 1, ProcessName: "demo.exe"}}}
	if got := m.View(); got == "" {
		t.Fatal("View() is empty")
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if updated == nil || cmd == nil {
		t.Fatal("q did not return quit command")
	}
}

func TestModelDetailsAndKillConfirmation(t *testing.T) {
	m := Model{Ports: []model.PortInfo{{Port: 8080, Protocol: "TCP", PID: 1, ProcessName: "demo.exe"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(Model).Detail == "" {
		t.Fatal("enter did not show details")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if !updated.(Model).ConfirmKill {
		t.Fatal("k did not request confirmation")
	}
}

func TestModelFilterInput(t *testing.T) {
	m := Model{Ports: []model.PortInfo{
		{Port: 3000, ProcessName: "node.exe"},
		{Port: 8080, ProcessName: "java.exe"},
	}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := updated.(Model).View()
	if !strings.Contains(view, "node.exe") || strings.Contains(view, "java.exe") {
		t.Fatalf("filtered view = %q", view)
	}
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if model := updated.(Model); model.Filtering || model.Filter != "" {
		t.Fatalf("escape did not clear filter: %+v", model)
	}
}
