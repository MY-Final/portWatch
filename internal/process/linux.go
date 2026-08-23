//go:build linux

package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

type LinuxManager struct{}

func NewManager() *LinuxManager { return &LinuxManager{} }

func (LinuxManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	if err := ValidatePID(pid); err != nil {
		return model.ProcessInfo{}, err
	}
	base := fmt.Sprint("/proc/", pid)
	name, err := os.ReadFile(filepath.Join(base, "comm"))
	if err != nil {
		return model.ProcessInfo{}, mapLinuxError(err)
	}
	executable, err := os.Readlink(filepath.Join(base, "exe"))
	if err != nil {
		return model.ProcessInfo{}, mapLinuxError(err)
	}
	command, err := os.ReadFile(filepath.Join(base, "cmdline"))
	if err != nil {
		return model.ProcessInfo{}, mapLinuxError(err)
	}
	info, err := model.NewProcessInfo(pid, strings.TrimSpace(string(name)), executable, strings.ReplaceAll(string(command), "\x00", " "), readCwd(base), "")
	if err != nil {
		return model.ProcessInfo{}, err
	}
	return info.WithParent(readStatusParentPID(filepath.Join(base, "status"))), nil
}

// readCwd resolves the /proc/<pid>/cwd link; empty when unreadable, which is
// common for processes owned by other users.
func readCwd(base string) string {
	target, err := os.Readlink(filepath.Join(base, "cwd"))
	if err != nil {
		return ""
	}
	return target
}

// readStatusParentPID extracts the PPid line from /proc/<pid>/status,
// returning 0 when the file is unreadable or malformed.
func readStatusParentPID(statusPath string) int {
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "PPid:") {
			continue
		}
		ppid, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "PPid:")))
		if err != nil {
			return 0
		}
		return ppid
	}
	return 0
}

func (LinuxManager) Exists(_ context.Context, pid int) (bool, error) {
	if err := ValidatePID(pid); err != nil {
		return false, err
	}
	_, err := os.Stat(fmt.Sprint("/proc/", pid))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (LinuxManager) Terminate(_ context.Context, pid int) error {
	if err := ValidatePID(pid); err != nil {
		return err
	}
	return exec.Command("kill", "-TERM", fmt.Sprint(pid)).Run()
}

func mapLinuxError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return ErrProcessNotFound
	}
	if errors.Is(err, os.ErrPermission) {
		return ErrAccessDenied
	}
	return err
}
