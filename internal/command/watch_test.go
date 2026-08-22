package command

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/portwatch/portwatch/pkg/model"
)

type watchCommandScanner struct {
	record model.PortInfo
}

func (s watchCommandScanner) List(context.Context) ([]model.PortInfo, error) {
	return []model.PortInfo{s.record}, nil
}

func (s watchCommandScanner) Port(context.Context, int) ([]model.PortInfo, error) { return nil, nil }

type watchCommandManager struct {
	cancel context.CancelFunc
}

func (m watchCommandManager) Info(context.Context, int) (model.ProcessInfo, error) {
	if m.cancel != nil {
		m.cancel()
	}
	return model.ProcessInfo{PID: 12, Name: "node.exe"}, nil
}

func (watchCommandManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (watchCommandManager) Terminate(context.Context, int) error      { return nil }

func TestWatchEnrichesProcessName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out, errOut strings.Builder
	err := Watch(ctx, watchCommandScanner{record: model.PortInfo{Port: 8080, Protocol: "TCP", PID: 12}}, watchCommandManager{cancel: cancel}, time.Millisecond, 0, &out, &errOut)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Watch() error = %v, want context.Canceled", err)
	}
	if !strings.Contains(out.String(), "TCP 8080 PID=12 node.exe") {
		t.Fatalf("output = %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("error output = %q", errOut.String())
	}
}

func TestWatchJSONEmitsSchemaVersionedEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out, errOut strings.Builder
	err := WatchJSON(ctx, watchCommandScanner{record: model.PortInfo{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 12}}, watchCommandManager{cancel: cancel}, time.Millisecond, 0, &out, &errOut)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WatchJSON() error = %v, want context.Canceled", err)
	}
	var event model.WatchEventResponse
	if err := json.Unmarshal([]byte(out.String()), &event); err != nil {
		t.Fatalf("event JSON error = %v; output=%q", err, out.String())
	}
	if event.SchemaVersion != model.JSONSchemaVersion || event.Event != "added" || event.Port != 8080 || event.ProcessName != "node.exe" {
		t.Fatalf("event = %+v", event)
	}
	if errOut.Len() != 0 {
		t.Fatalf("error output = %q", errOut.String())
	}
}
