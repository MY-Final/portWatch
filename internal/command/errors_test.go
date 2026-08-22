package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
	"github.com/portwatch/portwatch/pkg/model"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "success", want: ExitSuccess},
		{name: "invalid port", err: model.ErrInvalidPort, want: ExitArguments},
		{name: "invalid pid", err: fmtWrap(model.ErrInvalidPID), want: ExitArguments},
		{name: "permission", err: fmtWrap(process.ErrAccessDenied), want: ExitPermission},
		{name: "os permission", err: fmtWrap(os.ErrPermission), want: ExitPermission},
		{name: "unsupported scanner", err: fmtWrap(port.ErrUnsupported), want: ExitSystem},
		{name: "process not found", err: fmtWrap(process.ErrProcessNotFound), want: ExitSystem},
		{name: "not supported", err: fmtWrap(process.ErrNotSupported), want: ExitSystem},
		{name: "cancelled", err: fmtWrap(context.Canceled), want: ExitCancelled},
		{name: "kill failure", err: fmtWrap(ErrKillFailed), want: ExitKill},
		{name: "protected process", err: fmtWrap(ErrProtectedProcess), want: ExitKill},
		{name: "unknown", err: errors.New("scan failed"), want: ExitSystem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}

	parseErr := &ParseError{Kind: ParseErrorInvalidPort, Message: "port must be a number"}
	if got := ExitCode(fmtWrap(parseErr)); got != ExitArguments {
		t.Fatalf("ExitCode(wrapped ParseError) = %d, want %d", got, ExitArguments)
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		want   string
		forbid string
	}{
		{name: "parse", err: &ParseError{Message: "bad port"}, want: "portwatch: bad port\n"},
		{name: "permission", err: fmtWrap(process.ErrAccessDenied), want: "请以管理员身份运行"},
		{name: "cancelled", err: context.Canceled, want: "operation cancelled"},
		{name: "unknown", err: errors.New("wrapped context\nwith details"), want: "portwatch: operation failed: wrapped context with details\n", forbid: "stack trace"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			PrintError(&output, tt.err)
			got := output.String()
			if !strings.Contains(got, tt.want) {
				t.Fatalf("PrintError() = %q, want substring %q", got, tt.want)
			}
			if tt.forbid != "" && strings.Contains(got, tt.forbid) {
				t.Fatalf("PrintError() contains forbidden text %q: %q", tt.forbid, got)
			}
		})
	}
}

func fmtWrap(err error) error {
	return fmt.Errorf("operation context: %w", err)
}
