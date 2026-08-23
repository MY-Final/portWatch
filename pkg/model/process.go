package model

import "errors"

// ErrInvalidPID indicates that a process identifier is negative.
var ErrInvalidPID = errors.New("invalid process id")

// ProcessInfo describes the user-visible details of a process.
type ProcessInfo struct {
	PID        int    `json:"pid"`
	Name       string `json:"name"`
	Executable string `json:"executable"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
	User       string `json:"user"`
	// ParentPID is the process that spawned this one; 0 when unknown.
	ParentPID int `json:"parent_pid"`
}

// NewProcessInfo constructs a ProcessInfo after validating its PID.
func NewProcessInfo(pid int, name, executable, command, workingDir, user string) (ProcessInfo, error) {
	if pid < 0 {
		return ProcessInfo{}, ErrInvalidPID
	}
	return ProcessInfo{
		PID:        pid,
		Name:       name,
		Executable: executable,
		Command:    command,
		WorkingDir: workingDir,
		User:       user,
	}, nil
}

// WithParent returns a copy of the process with the parent PID filled in,
// keeping the constructor signature stable for callers without parent data.
func (p ProcessInfo) WithParent(parentPID int) ProcessInfo {
	p.ParentPID = parentPID
	return p
}
