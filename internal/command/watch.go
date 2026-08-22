package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
	"github.com/portwatch/portwatch/internal/watch"
	"github.com/portwatch/portwatch/pkg/model"
)

func Watch(ctx context.Context, scanner port.Scanner, manager process.Manager, interval time.Duration, portNumber int, out, errOut io.Writer) error {
	if out == nil || errOut == nil {
		return errors.New("watch output writer is nil")
	}
	names := make(map[int]string)
	return (watch.Engine{
		Scanner: scanner, Interval: interval,
		OnScanError: func(err error) error {
			_, writeErr := fmt.Fprintf(errOut, "portwatch: watch scan failed: %v\n", err)
			return writeErr
		},
		Enrich: func(record model.PortInfo) model.PortInfo {
			if record.ProcessName == "" && manager != nil && record.PID > 0 {
				if name := names[record.PID]; name != "" {
					record.ProcessName = name
					return record
				}
				if info, err := manager.Info(ctx, record.PID); err == nil {
					record.ProcessName = info.Name
					names[record.PID] = info.Name
				}
			}
			return record
		},
		Filter: func(record model.PortInfo) bool {
			return portNumber == 0 || record.Port == portNumber
		},
	}).Run(ctx, func(event watch.Event) error {
		processName := event.Port.ProcessName
		if processName == "" {
			processName = "-"
		}
		sign := "+"
		if event.Kind == watch.Removed {
			sign = "-"
		}
		_, err := fmt.Fprintf(out, "%s %s %s %d PID=%d %s\n", sign, event.ObservedAt.Format("15:04:05"), event.Port.Protocol, event.Port.Port, event.Port.PID, processName)
		return err
	})
}
