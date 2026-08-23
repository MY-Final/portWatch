package processinfo

import (
	"context"
	"errors"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

type chainManager struct {
	infos  map[int]model.ProcessInfo
	failed map[int]error
}

func (m chainManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	if err, ok := m.failed[pid]; ok {
		return model.ProcessInfo{}, err
	}
	if info, ok := m.infos[pid]; ok {
		return info, nil
	}
	return model.ProcessInfo{}, errors.New("process not found")
}

func TestAncestorsWalksToTop(t *testing.T) {
	manager := chainManager{infos: map[int]model.ProcessInfo{
		100: {PID: 100, Name: "node", ParentPID: 90},
		90:  {PID: 90, Name: "npm", ParentPID: 80},
		80:  {PID: 80, Name: "shell", ParentPID: 0},
	}}
	ancestors := Ancestors(context.Background(), manager, manager.infos[100], MaxAncestorHops)
	if len(ancestors) != 2 || ancestors[0].PID != 90 || ancestors[1].PID != 80 {
		t.Fatalf("ancestors = %+v", ancestors)
	}
}

func TestAncestorsStopsAtSystemBoundary(t *testing.T) {
	manager := chainManager{infos: map[int]model.ProcessInfo{
		100: {PID: 100, Name: "svc", ParentPID: 4},
	}}
	if ancestors := Ancestors(context.Background(), manager, manager.infos[100], MaxAncestorHops); len(ancestors) != 0 {
		t.Fatalf("ancestors = %+v, want none for parent <= 4", ancestors)
	}
}

func TestAncestorsBreaksCycles(t *testing.T) {
	manager := chainManager{infos: map[int]model.ProcessInfo{
		100: {PID: 100, Name: "a", ParentPID: 90},
		90:  {PID: 90, Name: "b", ParentPID: 100},
	}}
	ancestors := Ancestors(context.Background(), manager, manager.infos[100], MaxAncestorHops)
	if len(ancestors) != 1 || ancestors[0].PID != 90 {
		t.Fatalf("ancestors = %+v, want the cycle cut after one hop", ancestors)
	}
}

func TestAncestorsSelfReferenceStops(t *testing.T) {
	manager := chainManager{infos: map[int]model.ProcessInfo{
		100: {PID: 100, Name: "a", ParentPID: 100},
	}}
	if ancestors := Ancestors(context.Background(), manager, manager.infos[100], MaxAncestorHops); len(ancestors) != 0 {
		t.Fatalf("ancestors = %+v, want none for a self reference", ancestors)
	}
}

func TestAncestorsTruncatesOnFailure(t *testing.T) {
	manager := chainManager{
		infos: map[int]model.ProcessInfo{
			100: {PID: 100, Name: "a", ParentPID: 90},
			90:  {PID: 90, Name: "b", ParentPID: 80},
		},
		failed: map[int]error{80: errors.New("access denied")},
	}
	ancestors := Ancestors(context.Background(), manager, manager.infos[100], MaxAncestorHops)
	if len(ancestors) != 1 || ancestors[0].PID != 90 {
		t.Fatalf("ancestors = %+v, want truncation at the failed hop", ancestors)
	}
}

func TestAncestorsRespectsHopsCap(t *testing.T) {
	infos := map[int]model.ProcessInfo{}
	for pid := 100; pid < 100+32; pid++ {
		infos[pid] = model.ProcessInfo{PID: pid, Name: "p", ParentPID: pid + 1}
	}
	ancestors := Ancestors(context.Background(), chainManager{infos: infos}, infos[100], MaxAncestorHops)
	if len(ancestors) != MaxAncestorHops || ancestors[len(ancestors)-1].PID != 100+MaxAncestorHops {
		t.Fatalf("ancestors = %d entries, want exactly %d", len(ancestors), MaxAncestorHops)
	}
}

func TestFormatAncestors(t *testing.T) {
	if got := FormatAncestors("node", 100, nil); got != "" {
		t.Fatalf("FormatAncestors(nil) = %q, want empty", got)
	}
	got := FormatAncestors("node", 100, []model.ProcessAncestor{
		{PID: 90, Name: "npm"},
		{PID: 80, Name: ""},
	})
	want := "node (100) ← npm (90) ← unknown (80)"
	if got != want {
		t.Fatalf("FormatAncestors() = %q, want %q", got, want)
	}
}
