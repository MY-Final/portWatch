package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MY-Final/portWatch/internal/port"
	"github.com/MY-Final/portWatch/internal/process"
	"github.com/MY-Final/portWatch/pkg/model"
)

type pageMode uint8

const (
	pageList pageMode = iota
	pageDetails
	pageConfirm
	pageHelp
	pageView
)

type lookupState uint8

const (
	lookupOK lookupState = iota
	lookupAccessDenied
	lookupExited
	lookupUnknown
)

type rowKey struct {
	Protocol   string
	Port       int
	PID        int
	LocalAddr  string
	RemoteAddr string
}

func keyOf(record model.PortInfo) rowKey {
	return rowKey{Protocol: record.Protocol, Port: record.Port, PID: record.PID, LocalAddr: record.LocalAddr, RemoteAddr: record.RemoteAddr}
}

func classifyLookupError(err error) lookupState {
	switch {
	case err == nil:
		return lookupOK
	case errors.Is(err, process.ErrAccessDenied):
		return lookupAccessDenied
	case errors.Is(err, process.ErrProcessNotFound):
		return lookupExited
	default:
		return lookupUnknown
	}
}

func lookupMessage(state lookupState) string {
	switch state {
	case lookupAccessDenied:
		return "Process information unavailable. Access denied."
	case lookupExited:
		return "Process information unavailable. Process exited."
	case lookupUnknown:
		return "Process information unavailable."
	default:
		return ""
	}
}

func scopeLabel(scope port.Scope) string {
	switch scope {
	case port.ScopeConnections:
		return "CONNECTIONS"
	case port.ScopeAll:
		return "ALL"
	default:
		return "LISTENING"
	}
}

func modeKey(key string) (port.Scope, bool) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "l":
		return port.ScopeListening, true
	case "c":
		return port.ScopeConnections, true
	case "a":
		return port.ScopeAll, true
	default:
		return port.ScopeListening, false
	}
}

func unsupportedScopeError(scope port.Scope) error {
	return fmt.Errorf("%s view is unavailable on this platform: %w", scopeLabel(scope), port.ErrScopeUnsupported)
}
