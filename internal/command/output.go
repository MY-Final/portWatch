package command

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/portwatch/portwatch/internal/service"
	"github.com/portwatch/portwatch/pkg/model"
)

const outputHeader = "PORT\tPROTOCOL\tSTATE\tPID\tPROCESS NAME\tCOMMAND\tEXECUTABLE PATH\n"
const outputHeaderWithService = "PORT\tPROTOCOL\tSTATE\tPID\tPROCESS NAME\tCOMMAND\tEXECUTABLE PATH\tSERVICE\n"

// RenderPorts writes a deterministic table of port records to w. The input
// slice is copied before sorting so rendering never changes caller-owned data.
func RenderPorts(w io.Writer, ports []model.PortInfo) error {
	if w == nil {
		return errors.New("output writer is nil")
	}

	sorted := append([]model.PortInfo(nil), ports...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Port != sorted[j].Port {
			return sorted[i].Port < sorted[j].Port
		}
		return sorted[i].PID < sorted[j].PID
	})

	tw := tabwriter.NewWriter(w, 0, 4, 1, ' ', 0)
	if _, err := io.WriteString(tw, outputHeader); err != nil {
		return err
	}
	for _, port := range sorted {
		if err := writeRow(tw, port.Port, port.Protocol, port.State, port.PID, port.ProcessName, "", ""); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// RenderProcess writes one port record together with its resolved process
// details. Empty text fields are represented as "-".
func RenderProcess(w io.Writer, process model.ProcessInfo, port model.PortInfo) error {
	return renderProcess(w, process, port, nil)
}

// RenderProcessWithService writes one port record with an optional service hint.
func RenderProcessWithService(w io.Writer, process model.ProcessInfo, port model.PortInfo, detector service.Detector) error {
	return renderProcess(w, process, port, detector)
}

func renderProcess(w io.Writer, process model.ProcessInfo, port model.PortInfo, detector service.Detector) error {
	if w == nil {
		return errors.New("output writer is nil")
	}

	tw := tabwriter.NewWriter(w, 0, 4, 1, ' ', 0)
	header := outputHeader
	if detector != nil {
		header = outputHeaderWithService
	}
	if _, err := io.WriteString(tw, header); err != nil {
		return err
	}
	name := process.Name
	if name == "" {
		name = port.ProcessName
	}
	if detector != nil {
		info := detector.Detect(port, process)
		_, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
			port.Port, display(port.Protocol), display(port.State), port.PID,
			display(name), display(process.Command), display(process.Executable), display(info.Name))
		if err != nil {
			return err
		}
	} else if err := writeRow(tw, port.Port, port.Protocol, port.State, port.PID, name, process.Command, process.Executable); err != nil {
		return err
	}
	return tw.Flush()
}

func writeRow(w io.Writer, port int, protocol, state string, pid int, name, command, executable string) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		strconv.Itoa(port), display(protocol), display(state), strconv.Itoa(pid),
		display(name), display(command), display(executable))
	return err
}

func display(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
