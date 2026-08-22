package tui

import (
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
