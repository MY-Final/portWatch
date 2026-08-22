package command

import (
	"encoding/json"
	"errors"
	"io"
	"sort"

	"github.com/portwatch/portwatch/pkg/model"
)

func RenderJSON(w io.Writer, ports []model.PortInfo, infos map[int]model.ProcessInfo) error {
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
	results := make([]model.PortResult, 0, len(sorted))
	for _, port := range sorted {
		results = append(results, model.NewPortResult(port, infos[port.PID]))
	}
	return json.NewEncoder(w).Encode(model.PortsResponse{SchemaVersion: model.JSONSchemaVersion, Ports: results})
}

func RenderFreeJSON(w io.Writer, port int, err error) error {
	if w == nil {
		return errors.New("output writer is nil")
	}
	response := model.FreeResponse{SchemaVersion: model.JSONSchemaVersion, Port: port}
	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
	} else {
		response.Status = "available"
	}
	return json.NewEncoder(w).Encode(response)
}
