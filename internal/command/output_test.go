package command

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MY-Final/portWatch/internal/service"
	"github.com/MY-Final/portWatch/pkg/model"
)

func TestRenderPortsGolden(t *testing.T) {
	ports := []model.PortInfo{
		{Port: 9000, Protocol: "TCP", State: "LISTENING", PID: 22, ProcessName: "second"},
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 17, ProcessName: "first"},
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 9, ProcessName: "earlier"},
	}
	original := append([]model.PortInfo(nil), ports...)
	var got bytes.Buffer
	if err := RenderPorts(&got, ports); err != nil {
		t.Fatalf("RenderPorts() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID PROCESS NAME COMMAND EXECUTABLE PATH\n" +
		"8080 TCP      LISTENING 9   earlier      -       -\n" +
		"8080 TCP      LISTENING 17  first        -       -\n" +
		"9000 TCP      LISTENING 22  second       -       -\n"
	if got.String() != want {
		t.Fatalf("RenderPorts() = %q, want %q", got.String(), want)
	}
	if !equalPorts(ports, original) {
		t.Fatal("RenderPorts() changed input slice")
	}
}

func TestRenderProcessGolden(t *testing.T) {
	var got bytes.Buffer
	process := model.ProcessInfo{PID: 42, Name: "go", Command: "go run main.go", Executable: `C:\Go\bin\go.exe`}
	port := model.PortInfo{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 42}
	if err := RenderProcess(&got, process, port); err != nil {
		t.Fatalf("RenderProcess() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID PROCESS NAME COMMAND        EXECUTABLE PATH\n" +
		"8080 TCP      LISTENING 42  go           go run main.go C:\\Go\\bin\\go.exe\n"
	if got.String() != want {
		t.Fatalf("RenderProcess() = %q, want %q", got.String(), want)
	}
}

func TestRenderPortsEmptyAndMultilineCommand(t *testing.T) {
	var empty bytes.Buffer
	if err := RenderPorts(&empty, nil); err != nil {
		t.Fatalf("RenderPorts(empty) error = %v", err)
	}
	if !strings.HasPrefix(empty.String(), "PORT PROTOCOL STATE") {
		t.Fatalf("RenderPorts(empty) = %q, want header", empty.String())
	}

	var multiline bytes.Buffer
	if err := RenderProcess(&multiline, model.ProcessInfo{Name: "proc", Command: "first\nsecond"}, model.PortInfo{Port: 80, Protocol: "TCP", State: "LISTENING", PID: 1}); err != nil {
		t.Fatalf("RenderProcess(multiline) error = %v", err)
	}
	if !strings.Contains(multiline.String(), "first\nsecond") {
		t.Fatalf("RenderProcess(multiline) = %q, want command preserved", multiline.String())
	}
}

func equalPorts(a, b []model.PortInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRenderPortsTruncatesLongCells(t *testing.T) {
	cases := []struct {
		name   string
		length int
		kept   bool // true when the value must appear verbatim
	}{
		{name: "79 runes stays whole", length: 79, kept: true},
		{name: "80 runes stays whole", length: 80, kept: true},
		{name: "81 runes truncates", length: 81},
		{name: "5000 runes truncates", length: 5000},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, field := range []struct {
				label string
				apply func(info model.ProcessInfo, value string) model.ProcessInfo
			}{
				{label: "COMMAND", apply: func(info model.ProcessInfo, value string) model.ProcessInfo { info.Command = value; return info }},
				{label: "EXECUTABLE PATH", apply: func(info model.ProcessInfo, value string) model.ProcessInfo { info.Executable = value; return info }},
			} {
				t.Run(field.label, func(t *testing.T) {
					value := strings.Repeat("x", testCase.length)
					info := field.apply(model.ProcessInfo{PID: 7, Name: "noisy"}, value)
					var got bytes.Buffer
					if err := RenderProcess(&got, info, model.PortInfo{Port: 80, Protocol: "TCP", State: "LISTENING", PID: 7}); err != nil {
						t.Fatalf("RenderProcess() error = %v", err)
					}
					if testCase.kept {
						if !strings.Contains(got.String(), value) {
							t.Fatalf("RenderProcess() lost a %d-rune %s value", testCase.length, field.label)
						}
						return
					}
					truncated := strings.Repeat("x", 77) + "..."
					if !strings.Contains(got.String(), truncated) {
						t.Fatalf("RenderProcess() %s lacks %q", field.label, truncated)
					}
					if strings.Contains(got.String(), value) {
						t.Fatalf("RenderProcess() wrote the untruncated %s", field.label)
					}
				})
			}
		})
	}
}

func TestRenderPortsWithServicesGolden(t *testing.T) {
	ports := []model.PortInfo{{Port: 5404, Protocol: "TCP", State: "LISTENING", PID: 13220}}
	infos := map[int]model.ProcessInfo{
		13220: {PID: 13220, Name: "QQMusic.exe", Command: `"D:\Apps\Daily\QQMusic\QQMusic.exe" /background`, Executable: `D:\Apps\Daily\QQMusic\QQMusic.exe`},
	}
	var got bytes.Buffer
	if err := RenderPortsWithServices(&got, ports, infos, stubDetector{}); err != nil {
		t.Fatalf("RenderPortsWithServices() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID   PROCESS NAME COMMAND                                         EXECUTABLE PATH                   SERVICE\n" +
		"5404 TCP      LISTENING 13220 QQMusic.exe  \"D:\\Apps\\Daily\\QQMusic\\QQMusic.exe\" /background D:\\Apps\\Daily\\QQMusic\\QQMusic.exe qq-music\n"
	if got.String() != want {
		t.Fatalf("RenderPortsWithServices() = %q, want %q", got.String(), want)
	}
}

func TestRenderPortsWithServicesSortsPortThenPID(t *testing.T) {
	ports := []model.PortInfo{
		{Port: 8080, Protocol: "TCP", State: "LISTENING", PID: 20},
		{Port: 80, Protocol: "TCP", State: "LISTENING", PID: 300},
		{Port: 80, Protocol: "TCP", State: "LISTENING", PID: 100},
		{Port: 443, Protocol: "TCP", State: "LISTENING", PID: 50},
	}
	var got bytes.Buffer
	if err := RenderPortsWithServices(&got, ports, nil, nil); err != nil {
		t.Fatalf("RenderPortsWithServices() error = %v", err)
	}
	var pids []string
	for _, line := range strings.Split(strings.TrimSuffix(got.String(), "\n"), "\n")[1:] {
		fields := strings.Fields(line)
		pids = append(pids, fields[0]+"/"+fields[3])
	}
	want := []string{"80/100", "80/300", "443/50", "8080/20"}
	if !strings.EqualFold(strings.Join(pids, ","), strings.Join(want, ",")) {
		t.Fatalf("row order = %v, want %v", pids, want)
	}
}

func TestRenderPortsWithServicesMissingInfoShowsDashAndUnknown(t *testing.T) {
	ports := []model.PortInfo{{Port: 135, Protocol: "TCP", State: "LISTENING", PID: 992}}
	var got bytes.Buffer
	if err := RenderPortsWithServices(&got, ports, nil, nil); err != nil {
		t.Fatalf("RenderPortsWithServices() error = %v", err)
	}
	want := "PORT PROTOCOL STATE     PID PROCESS NAME COMMAND EXECUTABLE PATH SERVICE\n" +
		"135  TCP      LISTENING 992 -            -       -               Unknown\n"
	if got.String() != want {
		t.Fatalf("RenderPortsWithServices() = %q, want %q", got.String(), want)
	}
}

func TestRenderPortsWithServicesEmptyKeepsHeader(t *testing.T) {
	var got bytes.Buffer
	if err := RenderPortsWithServices(&got, nil, nil, nil); err != nil {
		t.Fatalf("RenderPortsWithServices(empty) error = %v", err)
	}
	want := "PORT PROTOCOL STATE PID PROCESS NAME COMMAND EXECUTABLE PATH SERVICE\n"
	if got.String() != want {
		t.Fatalf("RenderPortsWithServices(empty) = %q, want %q", got.String(), want)
	}
}

type stubDetector struct{}

func (stubDetector) Detect(model.PortInfo, model.ProcessInfo) service.Info {
	return service.Info{Name: "qq-music"}
}
