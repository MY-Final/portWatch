package process

import (
	"context"
	"errors"

	"github.com/portwatch/portwatch/pkg/model"
)

// Sentinel errors returned by process managers. Platform implementations may
// wrap these values with additional context while preserving errors.Is.
var (
	ErrProcessNotFound = errors.New("process not found")
	ErrAccessDenied    = errors.New("process access denied")
	ErrNotSupported    = errors.New("process operation not supported")
	ErrInvalidPID      = model.ErrInvalidPID
)

// Manager provides process lookup and termination operations.
type Manager interface {
	Info(ctx context.Context, pid int) (model.ProcessInfo, error)
	Exists(ctx context.Context, pid int) (bool, error)
	Terminate(ctx context.Context, pid int) error
}

// ValidatePID rejects the zero value and negative process identifiers.
// Platform implementations should call it before issuing an OS operation.
func ValidatePID(pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}
	return nil
}
