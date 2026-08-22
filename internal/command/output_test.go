package command

import (
	"bytes"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestRenderPortsGolden(t *testing.T) {
	ports := []model.PortInfo{
		{Port: 9000, Protocol: "TCP", State: "LISTENING", PID: 22, ProcessName: "second"},
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 17, ProcessName: "first"},
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 9, ProcessName: "earlier"},
	}
	original := append([]model.PortInfo(nil), ports...)
	var got bytes.Buffer
	if err := RenderPorts(&got, ports); err != nil {
		t.Fatalf("RenderPorts() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID PROCESS NAME COMMAND EXECUTABLE PATH\n" +
		"8080 TCP      LISTENING 9   earlier      -       -\n" +
		"8080 TCP      LISTENING 17  first        -       -\n" +
		"9000 TCP      LISTENING 22  second       -       -\n"
	if got.String() != want {
		t.Fatalf("RenderPorts() = %q, want %q", got.String(), want)
	}
	if !equalPorts(ports, original) {
		t.Fatal("RenderPorts() changed input slice")
	}
}

func TestRenderProcessGolden(t *testing.T) {
	var got bytes.Buffer
	process := model.ProcessInfo{PID: 42, Name: "go", Command: "go run main.go", Executable: `C:\Go\bin\go.exe`}
	port := model.PortInfo{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 42}
	if err := RenderProcess(&got, process, port); err != nil {
		t.Fatalf("RenderProcess() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID PROCESS NAME COMMAND        EXECUTABLE PATH\n" +
		"8080 TCP      LISTENING 42  go           go run main.go C:\\Go\\bin\\go.exe\n"
	if got.String() != want {
		t.Fatalf("RenderProcess() = %q, want %q", got.String(), want)
	}
}

func TestRenderPortsEmptyAndMultilineCommand(t *testing.T) {
	var empty bytes.Buffer
	if err := RenderPorts(&empty, nil); err != nil {
		t.Fatalf("RenderPorts(empty) error = %v", err)
	}
	if !strings.HasPrefix(empty.String(), "PORT PROTOCOL STATE") {
		t.Fatalf("RenderPorts(empty) = %q, want header", empty.String())
	}

	var multiline bytes.Buffer
	if err := RenderProcess(&multiline, model.ProcessInfo{Name: "proc", Command: "first\nsecond"}, model.PortInfo{Port: 80, Protocol: "TCP", State: "LISTENING", PID: 1}); err != nil {
		t.Fatalf("RenderProcess(multiline) error = %v", err)
	}
	if !strings.Contains(multiline.String(), "first\nsecond") {
		t.Fatalf("RenderProcess(multiline) = %q, want command preserved", multiline.String())
	}
}

func equalPorts(a, b []model.PortInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
