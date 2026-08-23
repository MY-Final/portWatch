package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
	"github.com/MY-Final/portWatch/pkg/model"
)

type processPorts struct {
	Info  model.ProcessInfo
	Ports []int
}

// Find lists processes whose name contains query, case-insensitively.
func Find(ctx context.Context, scanner port.Scanner, manager process.Manager, query string, out io.Writer) error {
	if out == nil {
		return errors.New("find output writer is nil")
	}
	rows, normalized, err := findRecords(ctx, scanner, manager, query)
	if err != nil {
		return err
	}
	return renderFindTable(out, normalized, rows)
}

// FindJSON writes the stable JSON representation of process search results.
func FindJSON(ctx context.Context, scanner port.Scanner, manager process.Manager, query string, out io.Writer) error {
	if out == nil {
		return errors.New("find output writer is nil")
	}
	rows, normalized, err := findRecords(ctx, scanner, manager, query)
	if err != nil {
		return err
	}
	results := make([]model.ProcessResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, model.ProcessResult{
			PID: row.Info.PID, ProcessName: row.Info.Name, Executable: row.Info.Executable,
			Command: row.Info.Command, WorkingDir: row.Info.WorkingDir, User: row.Info.User,
			Ports: append([]int(nil), row.Ports...),
		})
	}
	return json.NewEncoder(out).Encode(model.FindResponse{
		SchemaVersion: model.JSONSchemaVersion,
		Query:         normalized,
		Processes:     results,
	})
}

func findRecords(ctx context.Context, scanner port.Scanner, manager process.Manager, query string) ([]processPorts, string, error) {
	if scanner == nil || manager == nil {
		return nil, "", errors.New("find dependencies are nil")
	}
	normalized := strings.TrimSpace(query)
	queryLower := strings.ToLower(normalized)
	if queryLower == "" {
		return nil, "", errors.New("find query is empty")
	}
	ports, err := scanner.List(ctx)
	if err != nil {
		return nil, normalized, fmt.Errorf("scan ports: %w", err)
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
		if strings.Contains(strings.ToLower(entry.Info.Name), queryLower) {
			sort.Ints(entry.Ports)
			rows = append(rows, *entry)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Info.PID < rows[j].Info.PID })
	return rows, normalized, nil
}

func renderFindTable(out io.Writer, _ string, rows []processPorts) error {
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
