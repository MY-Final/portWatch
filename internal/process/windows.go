//go:build windows

package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/portwatch/portwatch/pkg/model"
)

// WindowsManager reads process information using the Windows process API and
// WMI. It has no mutable state and is safe to share between callers.
type WindowsManager struct{}

// NewManager returns the native process manager for Windows.
func NewManager() *WindowsManager {
	return &WindowsManager{}
}

// Info returns the executable path from the process API and the remaining
// metadata from Win32_Process.
func (WindowsManager) Info(ctx context.Context, pid int) (model.ProcessInfo, error) {
	if err := ValidatePID(pid); err != nil {
		return model.ProcessInfo{}, err
	}

	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return model.ProcessInfo{}, mapProcessError("open process", pid, err)
	}
	defer windows.CloseHandle(handle)

	executable, err := queryExecutable(handle)
	if err != nil {
		return model.ProcessInfo{}, fmt.Errorf("query executable for pid %d: %w", pid, err)
	}

	metadata, err := queryMetadata(ctx, pid)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	return model.NewProcessInfo(pid, metadata.Name, executable, metadata.CommandLine, metadata.WorkingDir, "")
}

// Exists reports whether a process can be opened for query. The handle is
// always closed before this method returns.
func (WindowsManager) Exists(_ context.Context, pid int) (bool, error) {
	if err := ValidatePID(pid); err != nil {
		return false, err
	}

	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
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

type processMetadata struct {
	Name        string `json:"Name"`
	CommandLine string `json:"CommandLine"`
	WorkingDir  string `json:"WorkingDirectory"`
}

const metadataScript = `& {
  param([int]$ProcessId)
  $ErrorActionPreference = 'Stop'
  $process = Get-CimInstance -ClassName Win32_Process -Filter ("ProcessId = {0}" -f $ProcessId) | Select-Object -First 1
  if ($null -eq $process) {
    'null'
    return
  }
  [pscustomobject]@{
    Name = $process.Name
    CommandLine = $process.CommandLine
    WorkingDirectory = $process.WorkingDirectory
  } | ConvertTo-Json -Compress
}`

func queryMetadata(ctx context.Context, pid int) (processMetadata, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", metadataScript, "-Args", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return processMetadata{}, ctx.Err()
		}
		if strings.Contains(strings.ToLower(string(output)), "access is denied") {
			return processMetadata{}, fmt.Errorf("query WMI metadata for pid %d: %w", pid, ErrAccessDenied)
		}
		return processMetadata{}, fmt.Errorf("query WMI metadata for pid %d: %w", pid, err)
	}
	metadata, err := parseMetadata(output)
	if err != nil {
		if errors.Is(err, ErrProcessNotFound) {
			return processMetadata{}, fmt.Errorf("query WMI metadata for pid %d: %w", pid, err)
		}
		return processMetadata{}, fmt.Errorf("parse WMI metadata for pid %d: %w", pid, err)
	}
	return metadata, nil
}

func parseMetadata(data []byte) (processMetadata, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return processMetadata{}, ErrProcessNotFound
	}
	var metadata processMetadata
	if err := json.Unmarshal([]byte(trimmed), &metadata); err != nil {
		return processMetadata{}, fmt.Errorf("decode JSON: %w", err)
	}
	if metadata.Name == "" {
		return processMetadata{}, errors.New("decoded process metadata has no name")
	}
	return metadata, nil
}

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
	default:
		return fmt.Errorf("%s%s: %w", operation, suffix, err)
	}
}
