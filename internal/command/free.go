package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
	"github.com/MY-Final/portWatch/pkg/model"
)

var (
	// ErrUserCancelled indicates that the user declined process termination.
	ErrUserCancelled = errors.New("operation cancelled")
	// ErrPortStillOccupied indicates that verification found a listener again.
	ErrPortStillOccupied = errors.New("port is still occupied")
)

// Free confirms and terminates the process records currently listening on a
// port, then rescans the port to verify that it has been released.
func Free(ctx context.Context, scanner port.Scanner, manager process.Manager, portNumber int, in io.Reader, out io.Writer) error {
	if scanner == nil {
		return errors.New("port scanner is nil")
	}
	if manager == nil {
		return errors.New("process manager is nil")
	}
	if in == nil {
		return errors.New("confirmation reader is nil")
	}
	if out == nil {
		return errors.New("output writer is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	records, err := scanner.Port(ctx, portNumber)
	if err != nil {
		return fmt.Errorf("scan port %d: %w", portNumber, err)
	}
	if len(records) == 0 {
		_, _ = fmt.Fprintf(out, "Port %d is available.\n", portNumber)
		return nil
	}
	for _, record := range records {
		if err := validateKillTarget(record.PID); err != nil {
			return err
		}
	}

	infos := make([]model.ProcessInfo, 0, len(records))
	for _, record := range records {
		info, infoErr := manager.Info(ctx, record.PID)
		if infoErr != nil {
			return fmt.Errorf("get process info for pid %d: %w", record.PID, infoErr)
		}
		infos = append(infos, info)
		if renderErr := RenderProcess(out, info, record); renderErr != nil {
			return fmt.Errorf("render process %d: %w", record.PID, renderErr)
		}
	}

	_, _ = fmt.Fprintf(out, "Kill %d process(es) listening on port %d? [y/N] ", len(records), portNumber)
	answer, readErr := bufio.NewReader(in).ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fmt.Errorf("read confirmation: %w", readErr)
	}
	if !isConfirmation(answer) {
		_, _ = fmt.Fprintln(out, "Cancelled.")
		return ErrUserCancelled
	}

	terminated := make(map[int]struct{}, len(records))
	for i, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, seen := terminated[record.PID]; seen {
			continue
		}
		// Same reuse guard as kill: the identity was resolved before the
		// confirmation prompt, so re-verify it maps to the same process.
		fresh, recheckErr := manager.Info(ctx, record.PID)
		if recheckErr != nil {
			return fmt.Errorf("%w: re-verify pid %d before termination: %w", ErrKillFailed, record.PID, recheckErr)
		}
		if fresh.Name != infos[i].Name || fresh.Executable != infos[i].Executable {
			return fmt.Errorf("%w: pid %d changed from %q to %q while waiting for confirmation; refusing to terminate", ErrKillFailed, record.PID, infos[i].Name, fresh.Name)
		}
		if err := manager.Terminate(ctx, record.PID); err != nil {
			return fmt.Errorf("%w: terminate pid %d (%s): %w", ErrKillFailed, record.PID, infos[i].Name, err)
		}
		terminated[record.PID] = struct{}{}
	}

	_, _ = fmt.Fprintln(out, "Verifying port release...")
	remaining, err := scanner.Port(ctx, portNumber)
	if err != nil {
		return fmt.Errorf("%w: verify port %d: %w", ErrKillFailed, portNumber, err)
	}
	if len(remaining) != 0 {
		return fmt.Errorf("%w: %w: port %d still has %d listener(s)", ErrKillFailed, ErrPortStillOccupied, portNumber, len(remaining))
	}
	_, _ = fmt.Fprintf(out, "Port %d is now available.\n", portNumber)
	return nil
}

func isConfirmation(answer string) bool {
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
