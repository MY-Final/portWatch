package command

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/portwatch/portwatch/internal/service"
	"github.com/portwatch/portwatch/pkg/model"
)

func TestRenderJSONIsValidAndSorted(t *testing.T) {
	var out bytes.Buffer
	ports := []model.PortInfo{{Port: 9000, PID: 20}, {Port: 8080, PID: 10}}
	if err := RenderJSON(&out, ports, map[int]model.ProcessInfo{10: {PID: 10, Name: "a.exe"}}); err != nil {
		t.Fatalf("RenderJSON() error = %v", err)
	}
	var response model.PortsResponse
	if err := json.Unmarshal(out.Bytes(), &response); err != nil {
		t.Fatalf("JSON decode error = %v; output=%q", err, out.String())
	}
	if response.SchemaVersion != model.JSONSchemaVersion || len(response.Ports) != 2 || response.Ports[0].Port != 8080 {
		t.Fatalf("response = %+v", response)
	}
}

func TestRenderFreeJSON(t *testing.T) {
	var out bytes.Buffer
	if err := RenderFreeJSON(&out, 8080, nil); err != nil {
		t.Fatalf("RenderFreeJSON() error = %v", err)
	}
	var response model.FreeResponse
	if err := json.Unmarshal(out.Bytes(), &response); err != nil {
		t.Fatalf("JSON decode error = %v", err)
	}
	if response.Status != "available" || response.Port != 8080 {
		t.Fatalf("response = %+v", response)
	}
}

func TestFindJSONIsValidAndSorted(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{{Port: 8080, PID: 12}, {Port: 3000, PID: 12}}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{12: {
		PID: 12, Name: "Node.exe", Executable: `C:\\node.exe`, Command: "node server.js",
	}}}
	var out bytes.Buffer
	if err := FindJSON(context.Background(), scanner, manager, " node ", &out); err != nil {
		t.Fatalf("FindJSON() error = %v", err)
	}
	var response model.FindResponse
	if err := json.Unmarshal(out.Bytes(), &response); err != nil {
		t.Fatalf("JSON decode error = %v; output=%q", err, out.String())
	}
	if response.SchemaVersion != model.JSONSchemaVersion || response.Query != "node" || len(response.Processes) != 1 {
		t.Fatalf("response = %+v", response)
	}
	if got := response.Processes[0].Ports; len(got) != 2 || got[0] != 3000 || got[1] != 8080 {
		t.Fatalf("ports = %v", got)
	}
}

func TestRenderJSONWithServicesIncludesDetection(t *testing.T) {
	var out bytes.Buffer
	ports := []model.PortInfo{{Port: 5173, Protocol: "TCP", PID: 12}}
	infos := map[int]model.ProcessInfo{12: {PID: 12, Name: "node.exe", Command: "vite --host"}}
	if err := RenderJSONWithServices(&out, ports, infos, service.Rules{}); err != nil {
		t.Fatalf("RenderJSONWithServices() error = %v", err)
	}
	var response model.PortsResponse
	if err := json.Unmarshal(out.Bytes(), &response); err != nil {
		t.Fatalf("JSON decode error = %v", err)
	}
	if len(response.Ports) != 1 || response.Ports[0].Service == nil || response.Ports[0].Service.Name != "Vite" {
		t.Fatalf("response = %+v", response)
	}
}
