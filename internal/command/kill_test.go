package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
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

func TestKillRefusesProtectedPID(t *testing.T) {
	manager := &freeManager{infos: map[int]model.ProcessInfo{4: sampleProcess(4)}}
	err := Kill(context.Background(), manager, 4, strings.NewReader("yes\n"), &strings.Builder{})
	if !errors.Is(err, ErrProtectedProcess) {
		t.Fatalf("Kill() error = %v, want protected process", err)
	}
	if len(manager.terminated) != 0 {
		t.Fatalf("terminated=%v, want none", manager.terminated)
	}
}

// identitySwapManager simulates PID reuse: the first Info call reports the
// process shown to the user, the second (the pre-termination re-check)
// reports a different process now owning the PID.
type identitySwapManager struct {
	calls      int
	terminated []int
}

func (m *identitySwapManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	m.calls++
	if m.calls == 1 {
		return model.ProcessInfo{PID: pid, Name: "server.exe", Executable: `C:\server.exe`}, nil
	}
	return model.ProcessInfo{PID: pid, Name: "innocent.exe", Executable: `C:\innocent.exe`}, nil
}
func (m *identitySwapManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (m *identitySwapManager) Terminate(_ context.Context, pid int) error {
	m.terminated = append(m.terminated, pid)
	return nil
}

func TestKillAbortsWhenIdentityChanges(t *testing.T) {
	manager := &identitySwapManager{}
	var out strings.Builder
	err := Kill(context.Background(), manager, 4242, strings.NewReader("y\n"), &out)
	if !errors.Is(err, ErrKillFailed) {
		t.Fatalf("Kill() error = %v, want ErrKillFailed", err)
	}
	if !strings.Contains(err.Error(), "changed from") || !strings.Contains(err.Error(), "innocent.exe") {
		t.Fatalf("error = %v, want identity-change detail", err)
	}
	if len(manager.terminated) != 0 {
		t.Fatalf("terminated = %v, must not terminate a reused PID", manager.terminated)
	}
}
