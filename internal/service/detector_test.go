package service

import (
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestRulesDetectKnownServices(t *testing.T) {
	tests := []struct {
		command string
		name    string
		want    string
	}{
		{"vite --host", "node.exe", "Vite"},
		{"uvicorn main:app", "python.exe", "FastAPI / Uvicorn"},
		{"java -jar spring-app.jar", "java.exe", "Spring Boot"},
	}
	for _, tt := range tests {
		got := (Rules{}).Detect(model.PortInfo{}, model.ProcessInfo{Name: tt.name, Command: tt.command})
		if got.Name != tt.want {
			t.Errorf("Detect(%q) = %+v, want %s", tt.command, got, tt.want)
		}
	}
}

func TestRulesUsePortAndWorkingDirectoryHints(t *testing.T) {
	tests := []struct {
		port       int
		process    model.ProcessInfo
		want       string
		confidence int
	}{
		{5173, model.ProcessInfo{Name: "node.exe"}, "Vite", 80},
		{8848, model.ProcessInfo{Name: "java.exe"}, "Nacos", 70},
		{3000, model.ProcessInfo{Name: "node.exe", WorkingDir: `D:\\projects\\frontend`}, "Node.js", 70},
	}
	for _, tt := range tests {
		got := (Rules{}).Detect(model.PortInfo{Port: tt.port}, tt.process)
		if got.Name != tt.want || got.Confidence != tt.confidence {
			t.Errorf("Detect(port=%d, process=%+v) = %+v, want %s/%d", tt.port, tt.process, got, tt.want, tt.confidence)
		}
	}
}

func TestRulesKeepUnknownWhenHintsDoNotMatch(t *testing.T) {
	got := (Rules{}).Detect(model.PortInfo{Port: 5173}, model.ProcessInfo{Name: "python.exe"})
	if got.Name != "Unknown" || got.Confidence != 0 {
		t.Fatalf("Detect() = %+v, want Unknown/0", got)
	}
}
