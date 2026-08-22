package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/internal/process"
	"github.com/portwatch/portwatch/pkg/model"
)

// Exit codes returned by the command entry point.
const (
	ExitSuccess    = 0
	ExitSystem     = 1
	ExitArguments  = 2
	ExitPermission = 3
	ExitCancelled  = 4
)

// ExitCode maps an operation error to the stable CLI exit code. It never
// terminates the process, allowing callers and tests to decide what to do.
func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if errors.Is(err, context.Canceled) {
		return ExitCancelled
	}
	if errors.Is(err, ErrUserCancelled) {
		return ExitCancelled
	}
	if isArgumentError(err) {
		return ExitArguments
	}
	if errors.Is(err, process.ErrAccessDenied) || errors.Is(err, os.ErrPermission) || os.IsPermission(err) {
		return ExitPermission
	}
	return ExitSystem
}

func isArgumentError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr) ||
		errors.Is(err, model.ErrInvalidPort) ||
		errors.Is(err, model.ErrInvalidPID)
}

// PrintError writes one concise, user-facing error line. It deliberately uses
// Error() rather than formatted stack traces or platform handles.
func PrintError(w io.Writer, err error) {
	if w == nil {
		w = io.Discard
	}
	if err == nil {
		return
	}

	message := errorMessage(err)
	_, _ = fmt.Fprintf(w, "portwatch: %s\n", message)
}

func errorMessage(err error) string {
	var parseErr *ParseError
	if errors.As(err, &parseErr) {
		return cleanMessage(parseErr.Error())
	}
	switch {
	case errors.Is(err, context.Canceled):
		return "operation cancelled"
	case errors.Is(err, ErrUserCancelled):
		return "operation cancelled"
	case errors.Is(err, process.ErrAccessDenied), errors.Is(err, os.ErrPermission), os.IsPermission(err):
		return "permission denied; 请以管理员身份运行"
	case errors.Is(err, model.ErrInvalidPort):
		return "invalid port number"
	case errors.Is(err, model.ErrInvalidPID):
		return "invalid process id"
	case errors.Is(err, process.ErrProcessNotFound):
		return "process not found"
	case errors.Is(err, process.ErrNotSupported), errors.Is(err, port.ErrUnsupported):
		return "operation is not supported on this platform"
	default:
		return "operation failed: " + cleanMessage(err.Error())
	}
}

func cleanMessage(message string) string {
	message = strings.Join(strings.Fields(message), " ")
	if message == "" {
		return "operation failed"
	}
	return message
}
