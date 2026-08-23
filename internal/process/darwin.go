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
	info, err := model.NewProcessInfo(pid, fields[0], fields[0], strings.TrimSpace(string(output)), "", "")
	if err != nil {
		return model.ProcessInfo{}, err
	}
	return info.WithParent(queryDarwinParentPID(ctx, pid)), nil
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
