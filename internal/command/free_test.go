package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

type freeScanner struct {
	initial   []model.PortInfo
	remaining []model.PortInfo
	calls     int
	err       error
}

func (s *freeScanner) List(context.Context) ([]model.PortInfo, error) {
	return append([]model.PortInfo(nil), s.initial...), nil
}

func (s *freeScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if s.calls == 1 {
		return append([]model.PortInfo(nil), s.initial...), nil
	}
	return append([]model.PortInfo(nil), s.remaining...), nil
}

type freeManager struct {
	infos      map[int]model.ProcessInfo
	terminated []int
	infoErr    error
	termErr    error
}

func (m *freeManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	if m.infoErr != nil {
		return model.ProcessInfo{}, m.infoErr
	}
	return m.infos[pid], nil
}

func (m *freeManager) Exists(context.Context, int) (bool, error) { return false, nil }

func (m *freeManager) Terminate(_ context.Context, pid int) error {
	if m.termErr != nil {
		return m.termErr
	}
	m.terminated = append(m.terminated, pid)
	return nil
}

func samplePort(pid int) model.PortInfo {
	return model.PortInfo{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: pid, ProcessName: "demo.exe"}
}

func sampleProcess(pid int) model.ProcessInfo {
	return model.ProcessInfo{PID: pid, Name: "demo.exe", Executable: `C:\demo.exe`, Command: "demo --serve"}
}

func TestFreeAvailablePort(t *testing.T) {
	scanner := &freeScanner{}
	manager := &freeManager{}
	var out strings.Builder
	if err := Free(context.Background(), scanner, manager, 8080, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Free() error = %v", err)
	}
	if scanner.calls != 1 || !strings.Contains(out.String(), "available") {
		t.Fatalf("calls=%d output=%q", scanner.calls, out.String())
	}
}

func TestFreeRejectsWithoutTermination(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10)}}
	var out strings.Builder
	err := Free(context.Background(), scanner, manager, 8080, strings.NewReader("n\n"), &out)
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("Free() error = %v, want ErrUserCancelled", err)
	}
	if len(manager.terminated) != 0 || scanner.calls != 1 {
		t.Fatalf("terminated=%v scanner calls=%d", manager.terminated, scanner.calls)
	}
}

func TestFreeTerminatesAndVerifies(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10), samplePort(20)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10), 20: sampleProcess(20)}}
	var out strings.Builder
	if err := Free(context.Background(), scanner, manager, 8080, strings.NewReader("YES\n"), &out); err != nil {
		t.Fatalf("Free() error = %v", err)
	}
	if len(manager.terminated) != 2 || scanner.calls != 2 {
		t.Fatalf("terminated=%v scanner calls=%d", manager.terminated, scanner.calls)
	}
	if !strings.Contains(out.String(), "now available") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestFreeReportsVerificationFailure(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10)}, remaining: []model.PortInfo{samplePort(10)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10)}}
	err := Free(context.Background(), scanner, manager, 8080, strings.NewReader("y\n"), &strings.Builder{})
	if !errors.Is(err, ErrPortStillOccupied) {
		t.Fatalf("Free() error = %v, want ErrPortStillOccupied", err)
	}
}

func TestFreePropagatesTerminationError(t *testing.T) {
	want := errors.New("access denied")
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10)}, termErr: want}
	err := Free(context.Background(), scanner, manager, 8080, strings.NewReader("y\n"), &strings.Builder{})
	if !errors.Is(err, want) {
		t.Fatalf("Free() error = %v, want wrapped termination error", err)
	}
}
