package model

import (
	"errors"
	"testing"
)

func TestNewPortInfoPortValidation(t *testing.T) {
	tests := []struct {
		name string
		port int
		want error
	}{
		{name: "below minimum", port: 0, want: ErrInvalidPort},
		{name: "above maximum", port: 65536, want: ErrInvalidPort},
		{name: "minimum", port: 1},
		{name: "maximum", port: 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPortInfo(tt.port, "tcp", "127.0.0.1:8080", "0.0.0.0:0", "LISTENING", 42, "example")
			if tt.want != nil {
				if !errors.Is(err, tt.want) {
					t.Fatalf("NewPortInfo() error = %v, want %v", err, tt.want)
				}
				if got != (PortInfo{}) {
					t.Fatalf("NewPortInfo() returned non-zero value on error: %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPortInfo() unexpected error = %v", err)
			}
			if got.Port != tt.port {
				t.Errorf("NewPortInfo().Port = %d, want %d", got.Port, tt.port)
			}
		})
	}
}

func TestNewPortInfoPIDValidation(t *testing.T) {
	if _, err := NewPortInfo(8080, "tcp", "", "", "LISTENING", -1, ""); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("NewPortInfo() error = %v, want %v", err, ErrInvalidPID)
	}
	if got, err := NewPortInfo(8080, "tcp", "", "", "LISTENING", 0, ""); err != nil || got.PID != 0 {
		t.Fatalf("NewPortInfo() with PID 0 = %#v, %v; want valid PID 0", got, err)
	}
}

func TestNewProcessInfoPIDValidation(t *testing.T) {
	if _, err := NewProcessInfo(-1, "", "", "", "", ""); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("NewProcessInfo() error = %v, want %v", err, ErrInvalidPID)
	}
	got, err := NewProcessInfo(42, "go", "/usr/bin/go", "go test", "/tmp", "user")
	if err != nil {
		t.Fatalf("NewProcessInfo() unexpected error = %v", err)
	}
	want := ProcessInfo{PID: 42, Name: "go", Executable: "/usr/bin/go", Command: "go test", WorkingDir: "/tmp", User: "user"}
	if got != want {
		t.Errorf("NewProcessInfo() = %#v, want %#v", got, want)
	}
}
