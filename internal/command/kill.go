package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/MY-Final/portWatch/internal/process"
)

var (
	// ErrKillFailed marks termination or post-termination verification failures.
	ErrKillFailed = errors.New("process termination failed")
	// ErrProtectedProcess marks a PID that PortWatch refuses to terminate.
	ErrProtectedProcess = errors.New("protected process")
)

// Kill confirms and terminates one PID, then verifies that it no longer exists.
func Kill(ctx context.Context, manager process.Manager, pid int, in io.Reader, out io.Writer) error {
	if manager == nil || in == nil || out == nil {
		return errors.New("kill dependencies are nil")
	}
	if err := validateKillTarget(pid); err != nil {
		return err
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
	// The PID was resolved while the confirmation prompt was displayed. If
	// the process died and the PID was reused in the meantime, terminating
	// now would hit an unrelated process; re-verify the identity first.
	fresh, recheckErr := manager.Info(ctx, pid)
	if recheckErr != nil {
		return fmt.Errorf("%w: re-verify pid %d before termination: %w", ErrKillFailed, pid, recheckErr)
	}
	if fresh.Name != info.Name || fresh.Executable != info.Executable {
		return fmt.Errorf("%w: pid %d changed from %q to %q while waiting for confirmation; refusing to terminate", ErrKillFailed, pid, info.Name, fresh.Name)
	}
	if err := manager.Terminate(ctx, pid); err != nil {
		return fmt.Errorf("%w: terminate pid %d: %w", ErrKillFailed, pid, err)
	}
	exists, err := manager.Exists(ctx, pid)
	if err != nil {
		return fmt.Errorf("%w: verify pid %d termination: %w", ErrKillFailed, pid, err)
	}
	if exists {
		return fmt.Errorf("%w: %w: pid %d", ErrKillFailed, ErrPortStillOccupied, pid)
	}
	_, _ = fmt.Fprintln(out, "Process terminated.")
	return nil
}

func validateKillTarget(pid int) error {
	if pid == 4 || pid == os.Getpid() {
		return fmt.Errorf("%w: refusing to terminate pid %d", ErrProtectedProcess, pid)
	}
	return nil
}
