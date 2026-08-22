package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
	"github.com/portwatch/portwatch/pkg/model"
)

type processPorts struct {
	Info  model.ProcessInfo
	Ports []int
}

// Find lists processes whose name contains query, case-insensitively.
func Find(ctx context.Context, scanner port.Scanner, manager process.Manager, query string, out io.Writer) error {
	if scanner == nil || manager == nil || out == nil {
		return errors.New("find dependencies are nil")
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return errors.New("find query is empty")
	}
	ports, err := scanner.List(ctx)
	if err != nil {
		return fmt.Errorf("scan ports: %w", err)
	}
	byPID := make(map[int]*processPorts)
	for _, record := range ports {
		entry := byPID[record.PID]
		if entry == nil {
			info, infoErr := manager.Info(ctx, record.PID)
			if infoErr != nil {
				continue
			}
			entry = &processPorts{Info: info}
			byPID[record.PID] = entry
		}
		entry.Ports = append(entry.Ports, record.Port)
	}
	rows := make([]processPorts, 0, len(byPID))
	for _, entry := range byPID {
		if strings.Contains(strings.ToLower(entry.Info.Name), query) {
			sort.Ints(entry.Ports)
			rows = append(rows, *entry)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Info.PID < rows[j].Info.PID })
	tw := tabwriter.NewWriter(out, 0, 4, 1, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PID\tPROCESS\tPORTS")
	for _, row := range rows {
		portsText := make([]string, len(row.Ports))
		for i, number := range row.Ports {
			portsText[i] = fmt.Sprint(number)
		}
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\n", row.Info.PID, display(row.Info.Name), strings.Join(portsText, ","))
	}
	return tw.Flush()
}
