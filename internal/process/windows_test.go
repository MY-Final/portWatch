//go:build windows

package process

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	metadata, err := parseMetadata([]byte(`{"Name":"go.exe","CommandLine":"go test","WorkingDirectory":"C:\\src"}`))
	if err != nil {
		t.Fatalf("parseMetadata() error = %v", err)
	}
	if metadata.Name != "go.exe" || metadata.CommandLine != "go test" || metadata.WorkingDir != `C:\src` {
		t.Fatalf("parseMetadata() = %+v", metadata)
	}
}

func TestParseMetadataErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
		want error
	}{
		{name: "null", data: "null", want: ErrProcessNotFound},
		{name: "empty", data: "", want: ErrProcessNotFound},
		{name: "invalid json", data: "{", want: nil},
		{name: "missing name", data: `{"CommandLine":"go test"}`, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMetadata([]byte(tt.data))
			if tt.want != nil {
				if !errors.Is(err, tt.want) {
					t.Fatalf("parseMetadata() error = %v, want %v", err, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatal("parseMetadata() error = nil, want error")
			}
		})
	}
}

func TestWindowsManagerInfoCurrentProcess(t *testing.T) {
	manager := WindowsManager{}
	info, err := manager.Info(context.Background(), os.Getpid())
	if err != nil {
		t.Fatalf("Info(current pid) error = %v", err)
	}
	if info.PID != os.Getpid() || info.Name == "" || info.Executable == "" {
		t.Fatalf("Info(current pid) = %+v", info)
	}
}

func TestWindowsManagerInfoMissingProcess(t *testing.T) {
	manager := WindowsManager{}
	_, err := manager.Info(context.Background(), int(^uint32(0)>>1))
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("Info(missing pid) error = %v, want ErrProcessNotFound", err)
	}
}

func TestWindowsManagerExistsCurrentProcess(t *testing.T) {
	manager := WindowsManager{}
	exists, err := manager.Exists(context.Background(), os.Getpid())
	if err != nil || !exists {
		t.Fatalf("Exists(current pid) = %v, %v", exists, err)
	}
}
