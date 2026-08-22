package model

// JSONSchemaVersion is bumped when the machine-readable response shape changes.
const JSONSchemaVersion = "2"

// PortResult is the stable, platform-neutral JSON representation of a port
// and its optional owning process.
type PortResult struct {
	Port        int            `json:"port"`
	Protocol    string         `json:"protocol"`
	LocalAddr   string         `json:"local_addr"`
	RemoteAddr  string         `json:"remote_addr"`
	State       string         `json:"state"`
	PID         int            `json:"pid"`
	ProcessName string         `json:"process_name"`
	Executable  string         `json:"executable"`
	Command     string         `json:"command"`
	Service     *ServiceResult `json:"service,omitempty"`
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

// ProcessResult is the stable JSON representation returned by process search.
type ProcessResult struct {
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	Executable  string `json:"executable"`
	Command     string `json:"command"`
	WorkingDir  string `json:"working_dir"`
	User        string `json:"user"`
	Ports       []int  `json:"ports"`
}

type FindResponse struct {
	SchemaVersion string          `json:"schema_version"`
	Query         string          `json:"query"`
	Processes     []ProcessResult `json:"processes"`
}

type ServiceResult struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Confidence int    `json:"confidence"`
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
