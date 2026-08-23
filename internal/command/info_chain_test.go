package command

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

type perPIDManager struct {
	infos map[int]model.ProcessInfo
}

func (m perPIDManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	if info, ok := m.infos[pid]; ok {
		return info, nil
	}
	return model.ProcessInfo{}, errors.New("process not found")
}
func (perPIDManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (perPIDManager) Terminate(context.Context, int) error      { return nil }

func chainTestDependencies() Dependencies {
	manager := perPIDManager{infos: map[int]model.ProcessInfo{
		42: {PID: 42, Name: "node.exe", ParentPID: 7},
		7:  {PID: 7, Name: "npm.exe", ParentPID: 4},
	}}
	return Dependencies{Scanner: infoScanner{}, Manager: manager}
}

func TestInfoRendersParentChain(t *testing.T) {
	deps := chainTestDependencies()
	var out strings.Builder
	if err := Info(context.Background(), deps.Scanner, deps.Manager, 42, &out); err != nil {
		t.Fatalf("Info() error = %v", err)
	}
	if !strings.Contains(out.String(), "node.exe (42) ← npm.exe (7)") {
		t.Fatalf("output = %q, want parent chain line", out.String())
	}
	if strings.Contains(out.String(), "← unknown") {
		t.Fatalf("output = %q, chain must stop at PID 4", out.String())
	}
}

func TestInfoJSONIncludesParentFields(t *testing.T) {
	deps := chainTestDependencies()
	var out strings.Builder
	if err := InfoJSON(context.Background(), deps.Scanner, deps.Manager, 42, &out); err != nil {
		t.Fatalf("InfoJSON() error = %v", err)
	}
	var response model.ProcessResponse
	if err := json.Unmarshal([]byte(out.String()), &response); err != nil {
		t.Fatalf("JSON error = %v; output=%q", err, out.String())
	}
	if response.SchemaVersion != model.ProcessSchemaVersion {
		t.Fatalf("schema version = %q, want unchanged %q", response.SchemaVersion, model.ProcessSchemaVersion)
	}
	if response.Process.ParentPID != 7 {
		t.Fatalf("parent_pid = %d, want 7", response.Process.ParentPID)
	}
	if len(response.Process.Ancestors) != 1 || response.Process.Ancestors[0].PID != 7 || response.Process.Ancestors[0].Name != "npm.exe" {
		t.Fatalf("ancestors = %+v", response.Process.Ancestors)
	}
}

func TestInfoParentChainBreaksCycles(t *testing.T) {
	manager := perPIDManager{infos: map[int]model.ProcessInfo{
		42: {PID: 42, Name: "a.exe", ParentPID: 7},
		7:  {PID: 7, Name: "b.exe", ParentPID: 42},
	}}
	var out strings.Builder
	if err := Info(context.Background(), infoScanner{}, manager, 42, &out); err != nil {
		t.Fatalf("Info() error = %v", err)
	}
	if !strings.Contains(out.String(), "a.exe (42) ← b.exe (7)") {
		t.Fatalf("output = %q", out.String())
	}
	if strings.Count(out.String(), "←") != 1 {
		t.Fatalf("output = %q, cycle must stop after one hop", out.String())
	}
}
