package tui

import (
	"context"
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

func TestModelFilterMatchesPortAndKeepsSelectionVisible(t *testing.T) {
	m := Model{Ports: []model.PortInfo{
		{Port: 3000, PID: 11, ProcessName: "node.exe"},
		{Port: 8080, PID: 22, ProcessName: "java.exe"},
	}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'8'}})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.Selected != 1 || !strings.Contains(model.View(), "java.exe") || strings.Contains(model.View(), "node.exe") {
		t.Fatalf("model = %+v, view=%q", model, model.View())
	}
}

type tuiTestManager struct {
	exists bool
}

func (m *tuiTestManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{Name: "demo.exe"}, nil
}
func (m *tuiTestManager) Exists(context.Context, int) (bool, error) { return m.exists, nil }
func (m *tuiTestManager) Terminate(context.Context, int) error      { return nil }

func TestModelKillVerificationFailureIsShown(t *testing.T) {
	manager := &tuiTestManager{exists: true}
	m := Model{Manager: manager, Ports: []model.PortInfo{{Port: 8080, PID: 42}}}
	updated, command := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated, command = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if command == nil {
		t.Fatal("kill confirmation did not produce command")
	}
	updated, _ = updated.(Model).Update(command())
	if updated.(Model).Err == nil {
		t.Fatal("expected verification error")
	}
}

type tuiContextScanner struct {
	seen context.Context
}

func (s *tuiContextScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	s.seen = ctx
	return nil, nil
}
func (s *tuiContextScanner) Port(context.Context, int) ([]model.PortInfo, error) { return nil, nil }

func TestRefreshUsesModelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scanner := &tuiContextScanner{}
	m := New(scanner, nil)
	m.Context = ctx
	msg := m.Init()()
	if _, ok := msg.(portsLoadedMsg); !ok {
		t.Fatalf("Init() message = %T, want portsLoadedMsg", msg)
	}
	if scanner.seen != ctx {
		t.Fatal("refresh did not use model context")
	}
}

func TestModelDoesNotActWhenFilterHasNoMatches(t *testing.T) {
	m := Model{Ports: []model.PortInfo{{Port: 8080, PID: 42, ProcessName: "node.exe"}}, Filter: "missing"}
	updated, command := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(Model).Detail != "" || command != nil {
		t.Fatalf("enter acted on hidden row: model=%+v command=%v", updated, command)
	}
	updated, command = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if updated.(Model).ConfirmKill || command != nil {
		t.Fatalf("kill acted on hidden row: model=%+v command=%v", updated, command)
	}
}
