package command

import (
	"strings"
	"testing"
	"time"
)

func TestParseFlagsInterspersed(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantJSON   bool
		wantPorts  string
		wantProto  string
		wantPos    []string
		wantSetPts bool
	}{
		{
			name:     "flags before positional still work",
			args:     []string{"--json", "8080"},
			wantJSON: true,
			wantPos:  []string{"8080"},
		},
		{
			name:     "positional before flag",
			args:     []string{"8080", "--json"},
			wantJSON: true,
			wantPos:  []string{"8080"},
		},
		{
			name:     "subcommand words then flag",
			args:     []string{"find", "node", "--json"},
			wantJSON: true,
			wantPos:  []string{"find", "node"},
		},
		{
			name:       "value flag after positional with space",
			args:       []string{"8080", "--ports", "3000,8080"},
			wantPorts:  "3000,8080",
			wantPos:    []string{"8080"},
			wantSetPts: true,
		},
		{
			name:       "value flag after positional with equals",
			args:       []string{"8080", "--ports=3000"},
			wantPorts:  "3000",
			wantPos:    []string{"8080"},
			wantSetPts: true,
		},
		{
			name:      "duration flag after positional",
			args:      []string{"watch", "--interval", "2s"},
			wantPos:   []string{"watch"},
			wantProto: "tcp",
		},
		{
			name:     "double dash keeps later tokens positional",
			args:     []string{"--json", "--", "8080", "--json"},
			wantJSON: true,
			wantPos:  []string{"8080", "--json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, positional, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("parseFlags(%v) error = %v", tt.args, err)
			}
			if options.JSON != tt.wantJSON {
				t.Fatalf("JSON = %v, want %v", options.JSON, tt.wantJSON)
			}
			if options.Ports != tt.wantPorts {
				t.Fatalf("Ports = %q, want %q", options.Ports, tt.wantPorts)
			}
			if options.PortsSet != tt.wantSetPts {
				t.Fatalf("PortsSet = %v, want %v", options.PortsSet, tt.wantSetPts)
			}
			if tt.wantProto != "" && options.Protocol != tt.wantProto {
				t.Fatalf("Protocol = %q, want %q", options.Protocol, tt.wantProto)
			}
			if strings.Join(positional, ",") != strings.Join(tt.wantPos, ",") {
				t.Fatalf("positional = %v, want %v", positional, tt.wantPos)
			}
		})
	}
}

func TestParseFlagsInterspersedIntervalValue(t *testing.T) {
	_, _, err := parseFlags([]string{"watch", "--interval", "2s"})
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}
	options, _, err := parseFlags([]string{"--interval", "3s", "watch"})
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}
	if options.Interval != 3*time.Second {
		t.Fatalf("Interval = %v, want 3s", options.Interval)
	}
}

func TestParseFlagsRejectsUnknownFlagAnywhere(t *testing.T) {
	for _, args := range [][]string{
		{"--json", "--nope"},
		{"8080", "--nope"},
	} {
		if _, _, err := parseFlags(args); err == nil {
			t.Fatalf("parseFlags(%v) error = nil, want unknown-flag error", args)
		}
	}
}

func TestParseRootInterspersed(t *testing.T) {
	command, err := Parse([]string{"8080", "--json"})
	if err != nil {
		t.Fatalf("Parse([8080 --json]) error = %v", err)
	}
	if command.Action != ActionPort || command.Port != 8080 || !command.Flags.JSON {
		t.Fatalf("Parse([8080 --json]) = %+v", command)
	}
	command, err = Parse([]string{"find", "node", "--json"})
	if err != nil {
		t.Fatalf("Parse([find node --json]) error = %v", err)
	}
	if command.Action != ActionFind || command.Query != "node" || !command.Flags.JSON {
		t.Fatalf("Parse([find node --json]) = %+v", command)
	}
}
