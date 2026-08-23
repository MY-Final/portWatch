package command

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

type infoScanner struct {
	ports []model.PortInfo
}

func (s infoScanner) List(context.Context) ([]model.PortInfo, error) { return s.ports, nil }
func (s infoScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return nil, nil
}

type infoManager struct {
	info model.ProcessInfo
}

func (m infoManager) Info(context.Context, int) (model.ProcessInfo, error) { return m.info, nil }
func (infoManager) Exists(context.Context, int) (bool, error)              { return false, nil }
func (infoManager) Terminate(context.Context, int) error                   { return nil }

func TestInfoJSONIncludesSortedUniquePorts(t *testing.T) {
	var out strings.Builder
	err := InfoJSON(context.Background(), infoScanner{ports: []model.PortInfo{
		{Port: 8081, PID: 42}, {Port: 8080, PID: 7}, {Port: 8080, PID: 42},
	}}, infoManager{info: model.ProcessInfo{PID: 42, Name: "demo.exe", Command: "demo --serve"}}, 42, &out)
	if err != nil {
		t.Fatalf("InfoJSON() error = %v", err)
	}
	var response model.ProcessResponse
	if err := json.Unmarshal([]byte(out.String()), &response); err != nil {
		t.Fatalf("JSON error = %v; output=%q", err, out.String())
	}
	if response.SchemaVersion != model.ProcessSchemaVersion || response.Process.PID != 42 || response.Process.Name != "demo.exe" || len(response.Process.Ports) != 2 || response.Process.Ports[0] != 8080 || response.Process.Ports[1] != 8081 {
		t.Fatalf("response = %+v", response)
	}
}

func TestParseInfoCommand(t *testing.T) {
	command, err := Parse([]string{"info", "42"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if command.Action != ActionInfo || command.PID != 42 {
		t.Fatalf("command = %+v", command)
	}
}

func TestRunInfoJSON(t *testing.T) {
	deps := Dependencies{
		Scanner: infoScanner{ports: []model.PortInfo{{Port: 8080, PID: 42}}},
		Manager: infoManager{info: model.ProcessInfo{PID: 42, Name: "demo.exe"}},
	}
	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"--json", "info", "42"}, deps, strings.NewReader(""), &stdout, &stderr); code != ExitSuccess {
		t.Fatalf("run() code = %d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"schema_version":"1"`) || !strings.Contains(stdout.String(), `"pid":42`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
