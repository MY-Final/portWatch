//go:build windows

package process

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"
)

func TestWindowsManagerTerminateChildProcess(t *testing.T) {
	child := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 30")
	if err := child.Start(); err != nil {
		t.Fatalf("start child process: %v", err)
	}
	childPID := child.Process.Pid
	waitDone := make(chan error, 1)
	go func() { waitDone <- child.Wait() }()
	t.Cleanup(func() {
		select {
		case <-waitDone:
		default:
			_ = child.Process.Kill()
			select {
			case <-waitDone:
			case <-time.After(time.Second):
			}
		}
	})

	manager := WindowsManager{}
	if err := manager.Terminate(context.Background(), childPID); err != nil {
		t.Fatalf("Terminate(child pid %d) error = %v", childPID, err)
	}
	exists, err := manager.Exists(context.Background(), childPID)
	if err != nil {
		t.Fatalf("Exists(child pid %d) error = %v", childPID, err)
	}
	if exists {
		t.Fatalf("child pid %d still exists after Terminate", childPID)
	}
	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Log("child wait did not complete promptly after termination")
	}
}

func TestWindowsManagerTerminateRejectsUnsafePIDs(t *testing.T) {
	manager := WindowsManager{}
	for _, test := range []struct {
		name string
		pid  int
		want error
	}{
		{name: "zero", pid: 0, want: ErrInvalidPID},
		{name: "system", pid: 4, want: ErrAccessDenied},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := manager.Terminate(ctx, test.pid)
			if !errors.Is(err, test.want) {
				t.Fatalf("Terminate(%d) error = %v, want %v", test.pid, err, test.want)
			}
		})
	}
}

func TestWindowsManagerTerminateMissingProcess(t *testing.T) {
	manager := WindowsManager{}
	err := manager.Terminate(context.Background(), int(^uint32(0)>>1))
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("Terminate(missing pid) error = %v, want ErrProcessNotFound", err)
	}
}
