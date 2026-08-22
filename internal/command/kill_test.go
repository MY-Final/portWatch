package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestKillRequiresConfirmation(t *testing.T) {
	manager := &freeManager{infos: map[int]model.ProcessInfo{12: sampleProcess(12)}}
	var out strings.Builder
	err := Kill(context.Background(), manager, 12, strings.NewReader("n\n"), &out)
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("Kill() error = %v, want cancellation", err)
	}
	if len(manager.terminated) != 0 {
		t.Fatalf("terminated=%v, want none", manager.terminated)
	}
}

func TestKillTerminatesAndVerifies(t *testing.T) {
	manager := &freeManager{infos: map[int]model.ProcessInfo{12: sampleProcess(12)}}
	var out strings.Builder
	if err := Kill(context.Background(), manager, 12, strings.NewReader("yes\n"), &out); err != nil {
		t.Fatalf("Kill() error = %v", err)
	}
	if len(manager.terminated) != 1 || !strings.Contains(out.String(), "Process terminated") {
		t.Fatalf("terminated=%v output=%q", manager.terminated, out.String())
	}
}
