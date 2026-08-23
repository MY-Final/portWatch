package port

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MY-Final/portWatch/pkg/model"
)

type fakeScanner struct {
	ports []model.PortInfo
}

func (f fakeScanner) List(context.Context) ([]model.PortInfo, error) {
	return f.ports, nil
}

func (f fakeScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return []model.PortInfo{}, nil
}

var _ Scanner = fakeScanner{}

func TestScannerContract(t *testing.T) {
	var scanner Scanner = fakeScanner{}
	ports, err := scanner.Port(context.Background(), 8080)
	if err != nil {
		t.Fatalf("Scanner.Port() unexpected error = %v", err)
	}
	if ports == nil {
		t.Fatal("Scanner.Port() returned nil slice for no matches")
	}
	if len(ports) != 0 {
		t.Fatalf("Scanner.Port() returned %d records, want 0", len(ports))
	}
}

func TestScannerErrors(t *testing.T) {
	if !errors.Is(ErrInvalidPort, model.ErrInvalidPort) {
		t.Fatalf("ErrInvalidPort = %v, want errors.Is(..., model.ErrInvalidPort)", ErrInvalidPort)
	}
	wrapped := fmt.Errorf("scan failed: %w", ErrUnsupported)
	if !errors.Is(wrapped, ErrUnsupported) {
		t.Fatalf("wrapped ErrUnsupported = %v, want errors.Is match", wrapped)
	}
}
