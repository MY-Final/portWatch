//go:build linux

package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	return model.NewProcessInfo(pid, strings.TrimSpace(string(name)), executable, strings.ReplaceAll(string(command), "\x00", " "), "", "")
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
