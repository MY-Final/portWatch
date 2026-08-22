package model

import (
	"encoding/json"
	"testing"
)

func TestPortResultJSONShape(t *testing.T) {
	data, err := json.Marshal(PortsResponse{
		SchemaVersion: JSONSchemaVersion,
		Ports:         []PortResult{NewPortResult(PortInfo{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 12}, ProcessInfo{Name: "demo.exe", Command: "demo", Executable: `C:\demo.exe`})},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := `{"schema_version":"2","ports":[{"port":8080,"protocol":"TCP","local_addr":"","remote_addr":"","state":"LISTENING","pid":12,"process_name":"demo.exe","executable":"C:\\demo.exe","command":"demo"}]}`
	if string(data) != want {
		t.Fatalf("Marshal() = %s, want %s", data, want)
	}
}
