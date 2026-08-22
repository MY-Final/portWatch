package tui

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
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

func TestArrowKeysMoveSelectionByKeyType(t *testing.T) {
	m := Model{Ports: []model.PortInfo{
		{Port: 3000, PID: 11, ProcessName: "node.exe"},
		{Port: 8080, PID: 22, ProcessName: "java.exe"},
		{Port: 9000, PID: 33, ProcessName: "go.exe"},
	}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.Selected != 1 {
		t.Fatalf("down selected index = %d, want 1", m.Selected)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.Selected != 2 {
		t.Fatalf("second down selected index = %d, want 2", m.Selected)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.Selected != 1 {
		t.Fatalf("up selected index = %d, want 1", m.Selected)
	}
}

func TestLetterNavigationFallbackMovesSelection(t *testing.T) {
	m := Model{Ports: []model.PortInfo{{Port: 3000}, {Port: 8080}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.Selected != 1 {
		t.Fatalf("j selected index = %d, want 1", m.Selected)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)
	if m.Selected != 0 {
		t.Fatalf("u selected index = %d, want 0", m.Selected)
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

type countingTUIManager struct {
	infoCalls int
}

func (m *countingTUIManager) Info(context.Context, int) (model.ProcessInfo, error) {
	m.infoCalls++
	return model.ProcessInfo{Name: "demo.exe"}, nil
}
func (m *countingTUIManager) Exists(context.Context, int) (bool, error) { return true, nil }
func (m *countingTUIManager) Terminate(context.Context, int) error      { return nil }

type duplicatePIDScanner struct{}

func (duplicatePIDScanner) List(context.Context) ([]model.PortInfo, error) {
	return []model.PortInfo{
		{Port: 3000, PID: 42},
		{Port: 8080, PID: 42},
	}, nil
}
func (duplicatePIDScanner) Port(context.Context, int) ([]model.PortInfo, error) { return nil, nil }

func TestRefreshCachesProcessInfoByPID(t *testing.T) {
	manager := &countingTUIManager{}
	m := New(duplicatePIDScanner{}, manager)
	message := m.Init()()
	loaded, ok := message.(portsLoadedMsg)
	if !ok {
		t.Fatalf("Init() message = %T, want portsLoadedMsg", message)
	}
	if manager.infoCalls != 1 {
		t.Fatalf("Info() calls = %d, want 1", manager.infoCalls)
	}
	if loaded.ports[0].ProcessName != "demo.exe" || loaded.ports[1].ProcessName != "demo.exe" {
		t.Fatalf("cached process names = %#v", loaded.ports)
	}
}

func TestRunProgramStartsAndExitsFromInput(t *testing.T) {
	scanner := &tuiContextScanner{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runProgram(ctx, scanner, nil,
		tea.WithInput(strings.NewReader("q")),
		tea.WithOutput(io.Discard),
		tea.WithoutRenderer(),
	); err != nil {
		t.Fatalf("runProgram() error = %v", err)
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

func TestV6ListExplainsPrimaryWorkflow(t *testing.T) {
	m := NewWithPort(nil, nil, 8080)
	m.Width = 100
	m.Ports = []model.PortInfo{
		{Port: 3000, Protocol: "TCP", State: "LISTENING", PID: 11, ProcessName: "node.exe"},
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 22, ProcessName: "java.exe"},
	}
	m.Selected = 1
	view := m.View()
	for _, want := range []string{"PortWatch", "LISTENING", "PORT 8080", "> 8080", "Selected: 8080 · PID 22 · java.exe", "Enter Details", "? Help"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q: %s", want, view)
		}
	}
	if strings.Contains(view, "Connections") || strings.Contains(view, "All") {
		t.Fatalf("list view exposes advanced modes: %s", view)
	}
}

func TestSelectionCursorAndSummaryMoveTogether(t *testing.T) {
	m := Model{Ports: []model.PortInfo{
		{Port: 3000, PID: 11, ProcessName: "node.exe"},
		{Port: 8080, PID: 22, ProcessName: "java.exe"},
	}}
	if !strings.Contains(m.View(), "> 3000") || !strings.Contains(m.View(), "Selected: 3000 · PID 11 · node.exe") {
		t.Fatalf("initial cursor/summary missing: %q", m.View())
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if !strings.Contains(m.View(), "> 8080") || !strings.Contains(m.View(), "Selected: 8080 · PID 22 · java.exe") {
		t.Fatalf("moved cursor/summary missing: %q", m.View())
	}
}

func TestV6DetailsAndConfirmAreSeparatePages(t *testing.T) {
	m := Model{Ports: []model.PortInfo{{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 42, ProcessName: "demo.exe"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.Page != pageDetails || !strings.Contains(m.View(), "Process Details") || strings.Contains(m.View(), "Terminate process?") {
		t.Fatalf("details page = page=%d view=%q", m.Page, m.View())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.Page != pageConfirm || !strings.Contains(m.View(), "Terminate process?") || !strings.Contains(m.View(), "PID      42") {
		t.Fatalf("confirm page = page=%d view=%q", m.Page, m.View())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.Page != pageDetails {
		t.Fatalf("cancel returned to page %d, want details", m.Page)
	}
}

func TestV6HelpAndViewMenu(t *testing.T) {
	m := Model{Scope: port.ScopeListening, Page: pageList}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.Page != pageHelp || !strings.Contains(m.View(), "How to use PortWatch") {
		t.Fatalf("help page = page=%d view=%q", m.Page, m.View())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = updated.(Model)
	if m.Page != pageView || !strings.Contains(m.View(), "Choose a view") {
		t.Fatalf("view menu = page=%d view=%q", m.Page, m.View())
	}
}

func TestV6LookupFailureIsDisplayedAsUnknown(t *testing.T) {
	m := Model{
		Ports:        []model.PortInfo{{Port: 8080, PID: 42, ProcessName: "stale.exe"}},
		LookupErrors: map[int]error{42: process.ErrAccessDenied},
	}
	if got := m.processName(m.Ports[0]); got != "Unknown" {
		t.Fatalf("processName() = %q, want Unknown", got)
	}
	if !strings.Contains(m.View(), "Unknown") {
		t.Fatalf("View() = %q, want Unknown", m.View())
	}
}

func TestV6EmptyResultsExplainNextState(t *testing.T) {
	focused := NewWithPort(nil, nil, 8080)
	if !strings.Contains(focused.View(), "Port 8080 is available") {
		t.Fatalf("focused empty view = %q", focused.View())
	}
	filtered := Model{Ports: []model.PortInfo{{Port: 8080}}, Filter: "missing"}
	if !strings.Contains(filtered.View(), `No match for "missing"`) {
		t.Fatalf("filtered empty view = %q", filtered.View())
	}
}

func TestV6KillSuccessStatusSurvivesRefresh(t *testing.T) {
	manager := &tuiTestManager{exists: false}
	scanner := duplicatePIDScanner{}
	m := Model{Scanner: scanner, Manager: manager, Ports: []model.PortInfo{{Port: 8080, Protocol: "TCP", PID: 42}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated, command := updated.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if command == nil {
		t.Fatal("confirm did not produce kill command")
	}
	updated, command = updated.(Model).Update(command())
	if command == nil {
		t.Fatal("successful kill did not schedule refresh")
	}
	updated, _ = updated.(Model).Update(command())
	if !strings.Contains(updated.(Model).Status, "Process terminated") {
		t.Fatalf("status = %q, want termination feedback", updated.(Model).Status)
	}
}
