package port

import (
	"context"
	"errors"

	"github.com/portwatch/portwatch/pkg/model"
)

// ErrUnsupported indicates that port scanning is not available on the
// current platform.
var ErrUnsupported = errors.New("port scanning is unsupported")

// ErrInvalidPort indicates that a requested port number is outside the valid
// TCP/UDP range of 1 through 65535.
//
// It aliases model.ErrInvalidPort so callers can use errors.Is consistently
// regardless of which package performed the validation.
var ErrInvalidPort = model.ErrInvalidPort

// Scope identifies the records requested by an interactive view.
type Scope uint8

const (
	ScopeListening Scope = iota
	ScopeConnections
	ScopeAll
)

func (s Scope) String() string {
	switch s {
	case ScopeListening:
		return "LISTENING"
	case ScopeConnections:
		return "CONNECTIONS"
	case ScopeAll:
		return "ALL"
	default:
		return "UNKNOWN"
	}
}

// ScopedScanner is an optional extension used by TUI views. Scanner.List
// remains the stable default listener query for CLI callers.
type ScopedScanner interface {
	Scanner
	ListScope(context.Context, Scope) ([]model.PortInfo, error)
}

// Scanner discovers listening port records.
//
// Port returns every listening record matching number. Implementations return
// an empty, non-nil slice when no record matches; absence is not an error.
type Scanner interface {
	List(ctx context.Context) ([]model.PortInfo, error)
	Port(ctx context.Context, number int) ([]model.PortInfo, error)
}

// ProtocolScanner is an optional extension for scanners that support more
// than the default TCP listener view. Protocol names are tcp, udp, or all.
type ProtocolScanner interface {
	Scanner
	ListProtocol(context.Context, string) ([]model.PortInfo, error)
	PortProtocol(context.Context, int, string) ([]model.PortInfo, error)
}
