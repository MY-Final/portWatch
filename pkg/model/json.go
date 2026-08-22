package model

const JSONSchemaVersion = "1"

// PortResult is the stable, platform-neutral JSON representation of a port
// and its optional owning process.
type PortResult struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"local_addr"`
	RemoteAddr  string `json:"remote_addr"`
	State       string `json:"state"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	Executable  string `json:"executable"`
	Command     string `json:"command"`
}

type PortsResponse struct {
	SchemaVersion string       `json:"schema_version"`
	Ports         []PortResult `json:"ports"`
}

type FreeResponse struct {
	SchemaVersion string `json:"schema_version"`
	Port          int    `json:"port"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

func NewPortResult(port PortInfo, process ProcessInfo) PortResult {
	return PortResult{
		Port:        port.Port,
		Protocol:    port.Protocol,
		LocalAddr:   port.LocalAddr,
		RemoteAddr:  port.RemoteAddr,
		State:       port.State,
		PID:         port.PID,
		ProcessName: firstNonEmpty(process.Name, port.ProcessName),
		Executable:  process.Executable,
		Command:     process.Command,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
