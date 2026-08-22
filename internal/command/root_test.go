package command

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestParseRootArguments(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		action Action
		port   int
	}{
		{name: "empty lists all", action: ActionList},
		{name: "numeric port", args: []string{"8080"}, action: ActionPort, port: 8080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Action != tt.action || got.Port != tt.port {
				t.Fatalf("Parse() = %#v, want action %d port %d", got, tt.action, tt.port)
			}
		})
	}
}

func TestParseRootErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		kind ParseErrorKind
	}{
		{name: "non numeric", args: []string{"abc"}, kind: ParseErrorInvalidPort},
		{name: "out of range", args: []string{"65536"}, kind: ParseErrorInvalidPort},
		{name: "unknown command", args: []string{"wat", "8080"}, kind: ParseErrorUnknownCommand},
		{name: "free without port", args: []string{"free"}, kind: ParseErrorFree},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)
			if err == nil {
				t.Fatal("Parse() error = nil")
			}
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if parseErr.Kind != tt.kind {
				t.Fatalf("error kind = %d, want %d", parseErr.Kind, tt.kind)
			}
		})
	}
}

func TestParseHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--version"}} {
		got, err := Parse(args)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", args, err)
		}
		if got.Action != ActionHelp && got.Action != ActionVersion {
			t.Fatalf("Parse(%v) action = %v", args, got.Action)
		}
	}
}

func TestRunReturnsExitCodeWithoutExit(t *testing.T) {
	var stderr strings.Builder
	if code := Run(context.Background(), []string{"not-a-port"}, Dependencies{}, nil, &stderr); code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: portwatch [port]") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
	deps := Dependencies{
		Scanner: fakeRootScanner{},
		Manager: fakeRootManager{},
	}
	if code := Run(context.Background(), nil, deps, nil, &stderr); code != 0 {
		t.Fatalf("Run() empty args code = %d, want 0", code)
	}
}

func TestParseFreeCommand(t *testing.T) {
	got, err := Parse([]string{"free", "8080"})
	if err != nil {
		t.Fatalf("Parse(free) error = %v", err)
	}
	if got.Action != ActionFree || got.Port != 8080 {
		t.Fatalf("Parse(free) = %#v, want ActionFree port 8080", got)
	}
}

func TestParseTUICommand(t *testing.T) {
	got, err := Parse([]string{"tui"})
	if err != nil {
		t.Fatalf("Parse(tui) error = %v", err)
	}
	if got.Action != ActionTUI {
		t.Fatalf("Parse(tui) action = %v, want ActionTUI", got.Action)
	}
}

func TestParseFindJoinsQueryWords(t *testing.T) {
	got, err := Parse([]string{"find", "spring", "boot"})
	if err != nil {
		t.Fatalf("Parse(find) error = %v", err)
	}
	if got.Action != ActionFind || got.Query != "spring boot" {
		t.Fatalf("Parse(find) = %#v", got)
	}
}

func TestParseWatchPort(t *testing.T) {
	got, err := Parse([]string{"watch", "8080"})
	if err != nil {
		t.Fatalf("Parse(watch) error = %v", err)
	}
	if got.Action != ActionWatch || got.Port != 8080 {
		t.Fatalf("Parse(watch) = %#v", got)
	}
}

func TestRunJSONDoesNotWriteTableOnProcessInfoError(t *testing.T) {
	deps := Dependencies{
		Scanner: onePortScanner{record: model.PortInfo{Port: 8080, Protocol: "TCP", PID: 42}},
		Manager: failingInfoManager{},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code == ExitSuccess {
		t.Fatal("run() code = success, want error")
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") || strings.Contains(stdout.String(), "PORT PROTOCOL") {
		t.Fatalf("stdout = %q, want JSON without table", stdout.String())
	}
	if !strings.Contains(stderr.String(), "operation failed") {
		t.Fatalf("stderr = %q, want error", stderr.String())
	}
}

func TestRunRejectsZeroWatchInterval(t *testing.T) {
	deps := Dependencies{Scanner: fakeRootScanner{}, Manager: fakeRootManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--interval", "0s", "watch"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSystem {
		t.Fatalf("run() code = %d, want %d", code, ExitSystem)
	}
	if !strings.Contains(stderr.String(), "watch interval must be positive") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunFreeJSONKeepsStdoutAsJSON(t *testing.T) {
	deps := Dependencies{Scanner: &freeScanner{}, Manager: &freeManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "free", "8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d, want %d", code, ExitSuccess)
	}
	var response model.FreeResponse
	if err := json.Unmarshal([]byte(stdout.String()), &response); err != nil {
		t.Fatalf("stdout JSON error = %v; stdout=%q; stderr=%q", err, stdout.String(), stderr.String())
	}
	if response.Status != "available" || response.Port != 8080 {
		t.Fatalf("response = %+v", response)
	}
}

type fakeRootScanner struct{}

func (fakeRootScanner) List(context.Context) ([]model.PortInfo, error) { return nil, nil }
func (fakeRootScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return nil, nil
}

type fakeRootManager struct{}

func (fakeRootManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, nil
}
func (fakeRootManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (fakeRootManager) Terminate(context.Context, int) error      { return nil }

type onePortScanner struct{ record model.PortInfo }

func (s onePortScanner) List(context.Context) ([]model.PortInfo, error) {
	return []model.PortInfo{s.record}, nil
}
func (s onePortScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return []model.PortInfo{s.record}, nil
}

type failingInfoManager struct{}

func (failingInfoManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, errors.New("metadata unavailable")
}
func (failingInfoManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (failingInfoManager) Terminate(context.Context, int) error      { return nil }
