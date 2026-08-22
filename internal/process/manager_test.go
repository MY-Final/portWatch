package process

import (
	"context"
	"errors"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

type fakeManager struct {
	infoErr      error
	existsErr    error
	terminateErr error
}

func (f fakeManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, f.infoErr
}

func (f fakeManager) Exists(context.Context, int) (bool, error) {
	return false, f.existsErr
}

func (f fakeManager) Terminate(context.Context, int) error {
	return f.terminateErr
}

func TestFakeManagerImplementsManager(t *testing.T) {
	var _ Manager = fakeManager{}
}

func TestSentinelErrorsCanBeWrapped(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{name: "not found", sentinel: ErrProcessNotFound},
		{name: "access denied", sentinel: ErrAccessDenied},
		{name: "not supported", sentinel: ErrNotSupported},
		{name: "invalid pid", sentinel: ErrInvalidPID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := errors.Join(errors.New("platform context"), tt.sentinel)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", wrapped, tt.sentinel)
			}
		})
	}
}

func TestValidatePID(t *testing.T) {
	for _, pid := range []int{-1, 0} {
		if err := ValidatePID(pid); !errors.Is(err, ErrInvalidPID) {
			t.Errorf("ValidatePID(%d) = %v, want ErrInvalidPID", pid, err)
		}
	}
	if err := ValidatePID(1); err != nil {
		t.Fatalf("ValidatePID(1) = %v, want nil", err)
	}
}
