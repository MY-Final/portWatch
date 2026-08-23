package command

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/MY-Final/portWatch/internal/service"
	"github.com/MY-Final/portWatch/pkg/model"
)

// RenderPortsWithServices is an opt-in service-aware table that preserves the
// original MVP table contract for callers that do not need detection.
func RenderPortsWithServices(w io.Writer, ports []model.PortInfo, infos map[int]model.ProcessInfo, detector service.Detector) error {
	if w == nil {
		return fmt.Errorf("output writer is nil")
	}
	sorted := append([]model.PortInfo(nil), ports...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Port != sorted[j].Port {
			return sorted[i].Port < sorted[j].Port
		}
		return sorted[i].PID < sorted[j].PID
	})
	tw := tabwriter.NewWriter(w, 0, 4, 1, ' ', 0)
	if _, err := fmt.Fprintln(tw, "PORT\tPROTOCOL\tSTATE\tPID\tPROCESS NAME\tCOMMAND\tEXECUTABLE PATH\tSERVICE"); err != nil {
		return err
	}
	for _, record := range sorted {
		info := infos[record.PID]
		serviceInfo := service.Info{Name: "Unknown"}
		if detector != nil {
			serviceInfo = detector.Detect(record, info)
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n", record.Port, display(record.Protocol), display(record.State), record.PID, display(info.Name), display(info.Command), display(info.Executable), display(serviceInfo.Name)); err != nil {
			return err
		}
	}
	return tw.Flush()
}
