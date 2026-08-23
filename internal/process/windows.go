//go:build windows

package process

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/sys/windows"

	"github.com/MY-Final/portWatch/pkg/model"
)

// WindowsManager reads process information through the Windows process API:
// the executable path via QueryFullProcessImageName and the command line and
// working directory from the process parameters in the PEB. No child
// processes are spawned. It has no mutable state and is safe to share
// between callers.
type WindowsManager struct{}

// NewManager returns the native process manager for Windows.
func NewManager() *WindowsManager {
	return &WindowsManager{}
}

// Info reads the executable path, command line and working directory from
// the process itself.
func (WindowsManager) Info(ctx context.Context, pid int) (model.ProcessInfo, error) {
	if err := ValidatePID(pid); err != nil {
		return model.ProcessInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return model.ProcessInfo{}, err
	}

	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_VM_READ, false, uint32(pid))
	if err != nil {
		return model.ProcessInfo{}, mapProcessError("open process", pid, err)
	}
	defer windows.CloseHandle(handle)

	executable, err := queryExecutable(handle)
	if err != nil {
		return model.ProcessInfo{}, fmt.Errorf("query executable for pid %d: %w", pid, err)
	}

	parameters, err := queryProcessParameters(handle, pid)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	info, err := model.NewProcessInfo(pid, filepath.Base(executable), executable, parameters.CommandLine, parameters.CurrentDirectory, "")
	if err != nil {
		return model.ProcessInfo{}, err
	}
	return info.WithParent(parameters.ParentPID), nil
}

// Exists reports whether a process can be opened for query. The handle is
// always closed before this method returns.
func (WindowsManager) Exists(_ context.Context, pid int) (bool, error) {
	if err := ValidatePID(pid); err != nil {
		return false, err
	}

	// GetExitCodeProcess only needs the limited query right; requesting
	// PROCESS_QUERY_INFORMATION would ask for more than the check requires.
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) || errors.Is(err, windows.ERROR_INVALID_HANDLE) {
			return false, nil
		}
		return false, mapProcessError("open process", pid, err)
	}
	var exitCode uint32
	exitErr := windows.GetExitCodeProcess(handle, &exitCode)
	if err := windows.CloseHandle(handle); err != nil {
		return false, fmt.Errorf("close process handle for pid %d: %w", pid, err)
	}
	if exitErr != nil {
		return false, fmt.Errorf("check process state for pid %d: %w", pid, exitErr)
	}
	return exitCode == 259, nil
}

// Terminate requests process termination and waits until the process can no
// longer be opened. Critical system processes are rejected before any API
// call so a caller cannot accidentally target the system process.
func (m WindowsManager) Terminate(ctx context.Context, pid int) error {
	if err := ValidatePID(pid); err != nil {
		return err
	}
	if pid == 4 {
		return fmt.Errorf("refusing to terminate critical pid %d: %w", pid, ErrAccessDenied)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return mapProcessError("open process for termination", pid, err)
	}
	if err := windows.TerminateProcess(handle, 1); err != nil {
		_ = windows.CloseHandle(handle)
		return mapProcessError("terminate process", pid, err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		if err := ctx.Err(); err != nil {
			_ = windows.CloseHandle(handle)
			return err
		}
		event, waitErr := windows.WaitForSingleObject(handle, pollIntervalMilliseconds)
		if waitErr != nil {
			_ = windows.CloseHandle(handle)
			return fmt.Errorf("wait for pid %d to exit: %w", pid, waitErr)
		}
		switch event {
		case windows.WAIT_OBJECT_0:
			if err := windows.CloseHandle(handle); err != nil {
				return fmt.Errorf("close process handle for pid %d: %w", pid, err)
			}
			exists, err := m.Exists(ctx, pid)
			if err != nil {
				return fmt.Errorf("verify termination for pid %d: %w", pid, err)
			}
			if exists {
				return fmt.Errorf("process pid %d still exists after termination", pid)
			}
			return nil
		case uint32(windows.WAIT_TIMEOUT):
			if time.Now().After(deadline) {
				_ = windows.CloseHandle(handle)
				return fmt.Errorf("timed out waiting for pid %d to exit", pid)
			}
		default:
			_ = windows.CloseHandle(handle)
			return fmt.Errorf("wait for pid %d to exit: unexpected wait result %#x", pid, event)
		}
	}
}

const pollIntervalMilliseconds = 50

func queryExecutable(handle windows.Handle) (string, error) {
	for size := uint32(260); size <= 32768; size *= 2 {
		buffer := make([]uint16, size)
		length := size
		err := windows.QueryFullProcessImageName(handle, 0, &buffer[0], &length)
		if err == nil {
			return windows.UTF16ToString(buffer[:length]), nil
		}
		if errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			continue
		}
		return "", mapProcessError("query executable", 0, err)
	}
	return "", errors.New("executable path exceeds maximum supported length")
}

func mapProcessError(operation string, pid int, err error) error {
	suffix := ""
	if pid > 0 {
		suffix = " for pid " + strconv.Itoa(pid)
	}
	switch {
	case errors.Is(err, windows.ERROR_ACCESS_DENIED):
		return fmt.Errorf("%s%s: %w", operation, suffix, ErrAccessDenied)
	case errors.Is(err, windows.ERROR_INVALID_PARAMETER), errors.Is(err, windows.ERROR_INVALID_HANDLE):
		return fmt.Errorf("%s%s: %w", operation, suffix, ErrProcessNotFound)
	case errors.Is(err, windows.ERROR_PARTIAL_COPY):
		// ReadProcessMemory reports a partial copy when the target exits
		// between opening the handle and reading its memory.
		return fmt.Errorf("%s%s: %w", operation, suffix, ErrProcessNotFound)
	default:
		return fmt.Errorf("%s%s: %w", operation, suffix, err)
	}
}
