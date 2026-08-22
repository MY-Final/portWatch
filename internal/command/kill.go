package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/portwatch/portwatch/internal/process"
)

// Kill confirms and terminates one PID, then verifies that it no longer exists.
func Kill(ctx context.Context, manager process.Manager, pid int, in io.Reader, out io.Writer) error {
	if manager == nil || in == nil || out == nil {
		return errors.New("kill dependencies are nil")
	}
	info, err := manager.Info(ctx, pid)
	if err != nil {
		return fmt.Errorf("get process info for pid %d: %w", pid, err)
	}
	_, _ = fmt.Fprintf(out, "PID: %d\nProcess: %s\nCommand: %s\nExecutable: %s\n", info.PID, info.Name, display(info.Command), display(info.Executable))
	_, _ = fmt.Fprintf(out, "Kill process %d (%s)? [y/N] ", pid, info.Name)
	answer, readErr := bufio.NewReader(in).ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fmt.Errorf("read confirmation: %w", readErr)
	}
	if !isConfirmation(answer) {
		_, _ = fmt.Fprintln(out, "Cancelled.")
		return ErrUserCancelled
	}
	if err := manager.Terminate(ctx, pid); err != nil {
		return fmt.Errorf("terminate pid %d: %w", pid, err)
	}
	exists, err := manager.Exists(ctx, pid)
	if err != nil {
		return fmt.Errorf("verify pid %d termination: %w", pid, err)
	}
	if exists {
		return fmt.Errorf("%w: pid %d", ErrPortStillOccupied, pid)
	}
	_, _ = fmt.Fprintln(out, "Process terminated.")
	return nil
}
