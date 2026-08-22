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
