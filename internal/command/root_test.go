package command

import (
	"context"
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
		{name: "unknown command", args: []string{"find", "8080"}, kind: ParseErrorUnknownCommand},
		{name: "free without port", args: []string{"free"}, kind: ParseErrorFree},
		{name: "help", args: []string{"--help"}, kind: ParseErrorHelp},
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
