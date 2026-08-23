package command

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MY-Final/portWatch/pkg/model"
)

// scriptedScanner returns canned Port results in order, repeating the last
// batch once the script is exhausted.
type scriptedScanner struct {
	scans   int
	records [][]model.PortInfo
	onScan  func(scan int)
}

func (s *scriptedScanner) List(context.Context) ([]model.PortInfo, error) { return nil, nil }

func (s *scriptedScanner) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
	index := s.scans
	if index >= len(s.records) {
		index = len(s.records) - 1
	}
	s.scans++
	if s.onScan != nil {
		s.onScan(s.scans)
	}
	return s.records[index], nil
}

func occupiedRecord(port int) []model.PortInfo {
	return []model.PortInfo{{Port: port, Protocol: "TCP", State: "LISTENING", PID: 42}}
}

func waitTestDeps(scanner *scriptedScanner) Dependencies {
	return Dependencies{Scanner: scanner, Manager: &countingPerPIDManager{calls: map[int]int{}}}
}

func TestWaitReturnsWhenPortFrees(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{occupiedRecord(8080), nil}}
	var out strings.Builder
	err := Wait(context.Background(), scanner, 8080, flagOptions{Expect: "free"}, &out)
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if !strings.Contains(out.String(), "Port 8080 is free.") {
		t.Fatalf("output = %q", out.String())
	}
	if scanner.scans != 2 {
		t.Fatalf("scans = %d, want 2", scanner.scans)
	}
}

func TestWaitReturnsImmediatelyWhenAlreadyFree(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{nil}}
	var out strings.Builder
	if err := Wait(context.Background(), scanner, 8080, flagOptions{Expect: "free"}, &out); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if scanner.scans != 1 {
		t.Fatalf("scans = %d, want a single immediate scan", scanner.scans)
	}
}

func TestWaitExpectOccupied(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{nil, occupiedRecord(8080)}}
	var out strings.Builder
	if err := Wait(context.Background(), scanner, 8080, flagOptions{Expect: "occupied"}, &out); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if !strings.Contains(out.String(), "Port 8080 is occupied.") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestWaitJSONEvent(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{nil}}
	var out strings.Builder
	if err := Wait(context.Background(), scanner, 8080, flagOptions{Expect: "free", JSON: true}, &out); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	line := out.String()
	if !strings.HasPrefix(line, `{"event":"port_free","observed_at":"`) {
		t.Fatalf("json event = %q", line)
	}
	if !strings.HasSuffix(strings.TrimSpace(line), `"port":8080,"protocol":"TCP"}`) {
		t.Fatalf("json event = %q", line)
	}
	const marker = `"observed_at":"`
	start := strings.Index(line, marker) + len(marker)
	end := strings.Index(line[start:], `"`)
	if start <= len(marker) || end < 0 {
		t.Fatalf("json event = %q, missing observed_at", line)
	}
	if _, err := time.Parse(time.RFC3339, line[start:start+end]); err != nil {
		t.Fatalf("observed_at is not RFC3339: %v", err)
	}
}

func TestRunWaitTimeoutExitCode(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{occupiedRecord(8080)}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--timeout", "60ms", "wait", "8080"}, waitTestDeps(scanner), strings.NewReader(""), &stdout, &stderr)
	if code != ExitWaitTimeout {
		t.Fatalf("code = %d, want %d", code, ExitWaitTimeout)
	}
	if !strings.Contains(stderr.String(), "timed out") || !strings.Contains(stderr.String(), "still occupied") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunWaitCustomIntervalAndTimeout(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{occupiedRecord(8080), nil}}
	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"--interval", "5ms", "--timeout", "5s", "wait", "8080"},
		waitTestDeps(scanner), strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
}

func TestRunWaitCancelExitsZero(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	scanner := &scriptedScanner{records: [][]model.PortInfo{occupiedRecord(8080)}}
	scanner.onScan = func(scan int) {
		if scan == 1 {
			cancel()
		}
	}
	var stdout, stderr strings.Builder
	code := run(ctx, []string{"wait", "8080"}, waitTestDeps(scanner), strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("code = %d, want 0 on cancel", code)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want silence on cancellation", stderr.String())
	}
}

func TestRunWaitJSONExitCode(t *testing.T) {
	scanner := &scriptedScanner{records: [][]model.PortInfo{nil}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "wait", "8080"}, waitTestDeps(scanner), strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stdout.String(), `"event":"port_free"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestParseWaitArguments(t *testing.T) {
	if _, err := Parse([]string{"wait"}); err == nil || !strings.Contains(err.Error(), "wait requires a port") {
		t.Fatalf("Parse(wait) error = %v", err)
	}
	if _, err := Parse([]string{"wait", "abc"}); err == nil {
		t.Fatal("wait abc must be rejected")
	}
	if _, err := Parse([]string{"wait", "0"}); err == nil {
		t.Fatal("wait 0 must be rejected")
	}
	if _, err := Parse([]string{"--expect", "maybe", "wait", "8080"}); err == nil || !strings.Contains(err.Error(), "--expect") {
		t.Fatalf("--expect validation error = %v", err)
	}
	if _, err := Parse([]string{"--timeout", "-1s", "wait", "8080"}); err == nil || !strings.Contains(err.Error(), "--timeout") {
		t.Fatalf("--timeout validation error = %v", err)
	}
	command, err := Parse([]string{"--expect", "occupied", "--timeout", "30s", "wait", "8080"})
	if err != nil || command.Action != ActionWait || command.Port != 8080 {
		t.Fatalf("Parse(wait flags) = (%+v, %v)", command, err)
	}
	if command.Flags.IntervalSet {
		t.Fatal("IntervalSet must be false when --interval is absent (wait then uses its own default)")
	}
}
