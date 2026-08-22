package command

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestParseQueryFilters(t *testing.T) {
	command, err := Parse([]string{"--process", "Node", "--pid", "42,42,7", "--state", "listening"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	filter, err := command.Flags.queryFilter()
	if err != nil {
		t.Fatalf("queryFilter() error = %v", err)
	}
	if filter.State != "LISTENING" || len(filter.PIDs) != 2 {
		t.Fatalf("filter = %+v", filter)
	}
}

func TestParseQueryFiltersRejectInvalidValues(t *testing.T) {
	for _, args := range [][]string{
		{"--pid", "0"},
		{"--pid", "42,nope"},
		{"--process="},
		{"--state", "unknown"},
	} {
		if _, err := Parse(args); err == nil {
			t.Errorf("Parse(%v) error = nil", args)
		}
	}
}

func TestFilterPortsMatchesStatePIDAndProcess(t *testing.T) {
	manager := &namedInfoManager{infos: map[int]model.ProcessInfo{
		7: {PID: 7, Name: "node.exe"},
		8: {PID: 8, Name: "java.exe"},
	}}
	records := []model.PortInfo{
		{Port: 3000, State: "LISTENING", PID: 7},
		{Port: 8080, State: "LISTENING", PID: 8},
		{Port: 5353, State: "BOUND", PID: 7},
	}
	filtered, infos, errs := filterPorts(context.Background(), manager, records, QueryFilter{Process: "NODE", PIDs: map[int]struct{}{7: {}}, State: "LISTENING"})
	if len(errs) != 0 || len(infos) != 1 || len(filtered) != 1 || filtered[0].Port != 3000 {
		t.Fatalf("filtered=%+v infos=%+v errors=%+v", filtered, infos, errs)
	}
}

func TestRunAppliesProcessFilterToJSON(t *testing.T) {
	deps := Dependencies{
		Scanner: rangeScanner{ports: []model.PortInfo{{Port: 3000, PID: 7}, {Port: 8080, PID: 8}}},
		Manager: &namedInfoManager{infos: map[int]model.ProcessInfo{7: {PID: 7, Name: "node.exe"}, 8: {PID: 8, Name: "java.exe"}}},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "--process", "node"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d stderr=%q", code, stderr.String())
	}
	var response model.PortsResponse
	if err := json.Unmarshal([]byte(stdout.String()), &response); err != nil {
		t.Fatalf("JSON error = %v; output=%q", err, stdout.String())
	}
	if len(response.Ports) != 1 || response.Ports[0].PID != 7 {
		t.Fatalf("ports = %+v", response.Ports)
	}
}

func TestRunRejectsFiltersOnFree(t *testing.T) {
	deps := Dependencies{Scanner: rangeScanner{}, Manager: &namedInfoManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--process", "node", "free", "8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitArguments || !strings.Contains(stderr.String(), "only supported") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

type namedInfoManager struct {
	infos map[int]model.ProcessInfo
}

func (m *namedInfoManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	return m.infos[pid], nil
}

func (m *namedInfoManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (m *namedInfoManager) Terminate(context.Context, int) error      { return nil }
