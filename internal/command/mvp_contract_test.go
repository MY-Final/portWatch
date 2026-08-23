package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

func TestFreeHonorsCancelledContextBeforeScanning(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10)}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Free(ctx, scanner, manager, 8080, strings.NewReader("y\n"), &strings.Builder{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Free() error = %v, want context.Canceled", err)
	}
	if scanner.calls != 0 || len(manager.terminated) != 0 {
		t.Fatalf("cancelled Free() performed work: scanner calls=%d terminated=%v", scanner.calls, manager.terminated)
	}
}

func TestFreeTerminatesDuplicatePIDOnce(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{samplePort(10), samplePort(10)}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{10: sampleProcess(10)}}
	if err := Free(context.Background(), scanner, manager, 8080, strings.NewReader("y\n"), &strings.Builder{}); err != nil {
		t.Fatalf("Free() error = %v", err)
	}
	if len(manager.terminated) != 1 || manager.terminated[0] != 10 {
		t.Fatalf("terminated=%v, want one termination for PID 10", manager.terminated)
	}
}
