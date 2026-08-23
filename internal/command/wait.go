package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MY-Final/portWatch/internal/port"
)

// waitDefaultInterval is faster than the 1s watch default: wait polls a
// single port, so a tighter loop is cheap and reacts quickly.
const waitDefaultInterval = 200 * time.Millisecond

// Wait blocks until the port reaches the state selected by --expect (free by
// default, occupied for waiting on a server to start). --timeout caps the
// wait; giving up reports errWaitTimeout, which maps to exit code 124.
func Wait(ctx context.Context, scanner port.Scanner, portNumber int, flags flagOptions, out io.Writer) error {
	if scanner == nil {
		return errors.New("port scanner is nil")
	}
	if out == nil {
		return errors.New("wait output writer is nil")
	}
	interval := flags.Interval
	if !flags.IntervalSet {
		interval = waitDefaultInterval
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be positive, got %s", interval)
	}
	expectOccupied := flags.Expect == "occupied"

	var deadline time.Time
	if flags.Timeout > 0 {
		deadline = time.Now().Add(flags.Timeout)
	}
	for {
		records, err := scanner.Port(ctx, portNumber)
		if err != nil {
			return fmt.Errorf("scan port %d: %w", portNumber, err)
		}
		occupied := len(records) > 0
		if occupied == expectOccupied {
			state := "free"
			event := "port_free"
			if occupied {
				state = "occupied"
				event = "port_occupied"
			}
			if flags.JSON {
				return emitWaitEvent(out, event, portNumber, time.Now())
			}
			_, _ = fmt.Fprintf(out, "Port %d is %s.\n", portNumber, state)
			return nil
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return fmt.Errorf("%w: port %d is still %s after %s", errWaitTimeout, portNumber, stateWord(expectOccupied), flags.Timeout)
		}
		if err := waitInterval(ctx, interval); err != nil {
			return err
		}
	}
}

func stateWord(expectOccupied bool) string {
	if expectOccupied {
		return "not occupied"
	}
	return "occupied"
}

// waitInterval sleeps for the poll interval, aborting promptly on context
// cancellation using the same timer pattern as the watch engine.
func waitInterval(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// emitWaitEvent prints the single wait event in the same style as the watch
// JSON Lines events. It is one terminal event, so no schema_version is
// involved.
func emitWaitEvent(out io.Writer, event string, portNumber int, observedAt time.Time) error {
	_, err := fmt.Fprintf(out, "{\"event\":%q,\"observed_at\":%q,\"port\":%d,\"protocol\":\"TCP\"}\n", event, observedAt.Format(time.RFC3339), portNumber)
	return err
}
