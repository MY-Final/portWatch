package command

import (
	"encoding/json"
	"errors"
	"io"
	"sort"

	"github.com/MY-Final/portWatch/internal/service"
	"github.com/MY-Final/portWatch/pkg/model"
)

func RenderJSON(w io.Writer, ports []model.PortInfo, infos map[int]model.ProcessInfo) error {
	return RenderJSONWithServices(w, ports, infos, nil)
}

func RenderJSONWithServices(w io.Writer, ports []model.PortInfo, infos map[int]model.ProcessInfo, detector service.Detector) error {
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
		result := model.NewPortResult(port, infos[port.PID])
		if detector != nil {
			info := detector.Detect(port, infos[port.PID])
			result.Service = &model.ServiceResult{Name: info.Name, Type: info.Type, Confidence: info.Confidence}
		}
		results = append(results, result)
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
