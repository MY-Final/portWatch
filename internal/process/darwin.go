//go:build darwin

package process

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

type DarwinManager struct{}

func NewManager() *DarwinManager { return &DarwinManager{} }

func (DarwinManager) Info(ctx context.Context, pid int) (model.ProcessInfo, error) {
	if err := ValidatePID(pid); err != nil {
		return model.ProcessInfo{}, err
	}
	output, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "comm=,command=").Output()
	if err != nil {
		return model.ProcessInfo{}, ErrProcessNotFound
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return model.ProcessInfo{}, ErrProcessNotFound
	}
	info, err := model.NewProcessInfo(pid, fields[0], fields[0], strings.TrimSpace(string(output)), queryDarwinWorkingDir(ctx, pid), "")
	if err != nil {
		return model.ProcessInfo{}, err
	}
	return info.WithParent(queryDarwinParentPID(ctx, pid)), nil
}

// queryDarwinWorkingDir asks lsof for the process working directory; empty
// on any failure, since lsof may be unavailable or the target protected.
func queryDarwinWorkingDir(ctx context.Context, pid int) string {
	output, err := exec.CommandContext(ctx, "lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return ""
	}
	return parseLsofCwd(string(output))
}

// parseLsofCwd extracts the first n<path> record from lsof -Fn output.
func parseLsofCwd(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if len(line) > 1 && line[0] == 'n' {
			return line[1:]
		}
	}
	return ""
}

// queryDarwinParentPID asks ps for the parent PID; 0 on any failure.
func queryDarwinParentPID(ctx context.Context, pid int) int {
	output, err := exec.CommandContext(ctx, "ps", "-o", "ppid=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0
	}
	ppid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0
	}
	return ppid
}
func (DarwinManager) Exists(_ context.Context, pid int) (bool, error) {
	if err := ValidatePID(pid); err != nil {
		return false, err
	}
	return exec.Command("kill", "-0", strconv.Itoa(pid)).Run() == nil, nil
}
func (DarwinManager) Terminate(_ context.Context, pid int) error {
	if err := ValidatePID(pid); err != nil {
		return err
	}
	return exec.Command("kill", "-TERM", fmt.Sprint(pid)).Run()
}
