package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/watch"
)

func Watch(ctx context.Context, scanner port.Scanner, interval time.Duration, out io.Writer) error {
	if out == nil {
		return errors.New("watch output writer is nil")
	}
	return (watch.Engine{Scanner: scanner, Interval: interval}).Run(ctx, func(event watch.Event) error {
		sign := "+"
		if event.Kind == watch.Removed {
			sign = "-"
		}
		_, err := fmt.Fprintf(out, "%s %s %d %s\n", sign, event.ObservedAt.Format("15:04:05"), event.Port.Port, event.Port.ProcessName)
		return err
	})
}
