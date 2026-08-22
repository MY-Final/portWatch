package model

import "errors"

// ErrInvalidPort indicates that a port number is outside the TCP/UDP range.
var ErrInvalidPort = errors.New("invalid port number")

// PortInfo describes one listening or connection endpoint discovered by a
// platform scanner.
type PortInfo struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"local_addr"`
	RemoteAddr  string `json:"remote_addr"`
	State       string `json:"state"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
}

// NewPortInfo constructs a PortInfo after validating its port and PID.
func NewPortInfo(port int, protocol, localAddr, remoteAddr, state string, pid int, processName string) (PortInfo, error) {
	if port < 1 || port > 65535 {
		return PortInfo{}, ErrInvalidPort
	}
	if pid < 0 {
		return PortInfo{}, ErrInvalidPID
	}
	return PortInfo{
		Port:        port,
		Protocol:    protocol,
		LocalAddr:   localAddr,
		RemoteAddr:  remoteAddr,
		State:       state,
		PID:         pid,
		ProcessName: processName,
	}, nil
}
