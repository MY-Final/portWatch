package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
	"github.com/MY-Final/portWatch/internal/processinfo"
	"github.com/MY-Final/portWatch/pkg/model"
)

// Info prints process metadata and the ports currently associated with pid.
func Info(ctx context.Context, scanner port.Scanner, manager process.Manager, pid int, out io.Writer) error {
	result, err := inspectProcess(ctx, scanner, manager, pid)
	if err != nil {
		return err
	}
	if out == nil {
		return errors.New("info output writer is nil")
	}
	tw := tabwriter.NewWriter(out, 0, 4, 1, ' ', 0)
	if _, err := fmt.Fprintln(tw, "FIELD\tVALUE"); err != nil {
		return err
	}
	fields := [][2]string{
		{"PID", fmt.Sprint(result.PID)},
		{"PROCESS NAME", display(result.Name)},
		{"PARENT CHAIN", display(processinfo.FormatAncestors(result.Name, result.PID, result.Ancestors))},
		{"EXECUTABLE PATH", display(result.Executable)},
		{"COMMAND", display(result.Command)},
		{"WORKING DIRECTORY", display(result.WorkingDir)},
		{"USER", display(result.User)},
		{"PORTS", formatPorts(result.Ports)},
	}
	for _, field := range fields {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", field[0], field[1]); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// InfoJSON writes the stable machine-readable info response.
func InfoJSON(ctx context.Context, scanner port.Scanner, manager process.Manager, pid int, out io.Writer) error {
	result, err := inspectProcess(ctx, scanner, manager, pid)
	if err != nil {
		return err
	}
	if out == nil {
		return errors.New("info output writer is nil")
	}
	return json.NewEncoder(out).Encode(model.ProcessResponse{
		SchemaVersion: model.ProcessSchemaVersion,
		Process:       result,
	})
}

func inspectProcess(ctx context.Context, scanner port.Scanner, manager process.Manager, pid int) (model.InfoProcessResult, error) {
	if scanner == nil || manager == nil {
		return model.InfoProcessResult{}, errors.New("info dependencies are nil")
	}
	if err := process.ValidatePID(pid); err != nil {
		return model.InfoProcessResult{}, err
	}
	info, err := manager.Info(ctx, pid)
	if err != nil {
		return model.InfoProcessResult{}, fmt.Errorf("get process info for pid %d: %w", pid, err)
	}
	ports, err := scanner.List(ctx)
	if err != nil {
		return model.InfoProcessResult{}, fmt.Errorf("scan ports for pid %d: %w", pid, err)
	}
	portNumbers := make([]int, 0)
	seen := make(map[int]struct{})
	for _, record := range ports {
		if record.PID != pid {
			continue
		}
		if _, ok := seen[record.Port]; ok {
			continue
		}
		seen[record.Port] = struct{}{}
		portNumbers = append(portNumbers, record.Port)
	}
	sort.Ints(portNumbers)
	ancestors := processinfo.Ancestors(ctx, manager, info, processinfo.MaxAncestorHops)
	chain := make([]model.ProcessAncestor, 0, len(ancestors))
	for _, ancestor := range ancestors {
		chain = append(chain, model.ProcessAncestor{PID: ancestor.PID, Name: ancestor.Name})
	}
	return model.InfoProcessResult{
		PID:        pid,
		Name:       info.Name,
		Executable: info.Executable,
		Command:    info.Command,
		WorkingDir: info.WorkingDir,
		User:       info.User,
		Ports:      portNumbers,
		ParentPID:  info.ParentPID,
		Ancestors:  chain,
	}, nil
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}
	text := ""
	for i, port := range ports {
		if i > 0 {
			text += ","
		}
		text += fmt.Sprint(port)
	}
	return text
}
