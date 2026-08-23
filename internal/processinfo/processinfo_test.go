package processinfo

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MY-Final/portWatch/pkg/model"
)

type recordingManager struct {
	mu    sync.Mutex
	calls map[int]int
	fail  map[int]error
}

func (m *recordingManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[pid]++
	if err, ok := m.fail[pid]; ok {
		return model.ProcessInfo{}, err
	}
	return model.ProcessInfo{PID: pid, Name: "proc"}, nil
}

func TestResolveQueriesEachUniquePIDOnce(t *testing.T) {
	manager := &recordingManager{calls: map[int]int{}}
	records := []model.PortInfo{
		{Port: 80, PID: 11}, {Port: 443, PID: 22}, {Port: 8080, PID: 11},
	}
	infos, errorsByPID := Resolve(context.Background(), manager, records)
	for _, pid := range []int{11, 22} {
		if manager.calls[pid] != 1 {
			t.Fatalf("Info() calls for pid %d = %d, want exactly 1", pid, manager.calls[pid])
		}
		if info, ok := infos[pid]; !ok || info.Name != "proc" {
			t.Fatalf("infos[%d] = %+v, ok=%v", pid, info, ok)
		}
	}
	if len(errorsByPID) != 0 {
		t.Fatalf("errorsByPID = %v, want none", errorsByPID)
	}
	if len(infos) != 2 {
		t.Fatalf("len(infos) = %d, want 2", len(infos))
	}
}

func TestResolveRecordsErrorsByPID(t *testing.T) {
	manager := &recordingManager{
		calls: map[int]int{},
		fail:  map[int]error{22: errors.New("denied")},
	}
	records := []model.PortInfo{{Port: 80, PID: 11}, {Port: 443, PID: 22}}
	infos, errorsByPID := Resolve(context.Background(), manager, records)
	if _, ok := infos[22]; ok {
		t.Fatal("failed PID must not appear in infos")
	}
	if err, ok := errorsByPID[22]; !ok || err.Error() != "denied" {
		t.Fatalf("errorsByPID[22] = %v, ok=%v", err, ok)
	}
	if len(errorsByPID) != 1 || len(infos) != 1 {
		t.Fatalf("infos=%v errorsByPID=%v", infos, errorsByPID)
	}
}

func TestResolveStopsDispatchingOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	manager := &recordingManager{calls: map[int]int{}}
	records := make([]model.PortInfo, 64)
	for i := range records {
		records[i] = model.PortInfo{Port: i + 1, PID: i + 1}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = Resolve(ctx, manager, records)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Resolve did not return after cancellation")
	}
}

func TestApplyNamesBackfillsMatchingRecords(t *testing.T) {
	records := []model.PortInfo{{Port: 80, PID: 11}, {Port: 443, PID: 22}}
	ApplyNames(records, map[int]model.ProcessInfo{11: {PID: 11, Name: "proc"}})
	if records[0].ProcessName != "proc" {
		t.Fatalf("records[0].ProcessName = %q", records[0].ProcessName)
	}
	if records[1].ProcessName != "" {
		t.Fatalf("records[1].ProcessName = %q, want empty", records[1].ProcessName)
	}
}
