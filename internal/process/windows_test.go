//go:build windows

package process

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPointer(t *testing.T) {
	if sizePointer == 8 {
		raw := []byte{0x78, 0x56, 0x34, 0x12, 0xEF, 0xCD, 0xAB, 0x89}
		if got, want := readPointer(raw), uintptr(0x89ABCDEF12345678); got != want {
			t.Fatalf("readPointer() = %#x, want %#x", got, want)
		}
		return
	}
	raw := []byte{0x78, 0x56, 0x34, 0x12}
	if got, want := readPointer(raw), uintptr(0x12345678); got != want {
		t.Fatalf("readPointer() = %#x, want %#x", got, want)
	}
}

func TestWindowsManagerInfoCurrentProcess(t *testing.T) {
	manager := WindowsManager{}
	info, err := manager.Info(context.Background(), os.Getpid())
	if err != nil {
		t.Fatalf("Info(current pid) error = %v", err)
	}
	if info.PID != os.Getpid() || info.Name == "" || info.Executable == "" {
		t.Fatalf("Info(current pid) = %+v", info)
	}
	if want := filepath.Base(info.Executable); info.Name != want {
		t.Fatalf("Info(current pid).Name = %q, want %q", info.Name, want)
	}
	if !strings.Contains(strings.ToLower(info.Command), "process.test") {
		t.Fatalf("Info(current pid).Command = %q, want substring process.test", info.Command)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if !strings.EqualFold(trimTrailingSeparators(info.WorkingDir), trimTrailingSeparators(cwd)) {
		t.Fatalf("Info(current pid).WorkingDir = %q, want %q", info.WorkingDir, cwd)
	}
}

func trimTrailingSeparators(path string) string {
	return strings.TrimRight(path, `\/`)
}

func TestWindowsManagerInfoMissingProcess(t *testing.T) {
	manager := WindowsManager{}
	_, err := manager.Info(context.Background(), int(^uint32(0)>>1))
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("Info(missing pid) error = %v, want ErrProcessNotFound", err)
	}
}

func TestWindowsManagerExistsCurrentProcess(t *testing.T) {
	manager := WindowsManager{}
	exists, err := manager.Exists(context.Background(), os.Getpid())
	if err != nil || !exists {
		t.Fatalf("Exists(current pid) = %v, %v", exists, err)
	}
}
