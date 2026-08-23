package command

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MY-Final/portWatch/pkg/model"
)

func TestParseRootArguments(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		action Action
		port   int
	}{
		{name: "empty lists all", action: ActionList},
		{name: "numeric port", args: []string{"8080"}, action: ActionPort, port: 8080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Action != tt.action || got.Port != tt.port {
				t.Fatalf("Parse() = %#v, want action %d port %d", got, tt.action, tt.port)
			}
		})
	}
}

func TestParseRootErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		kind ParseErrorKind
	}{
		{name: "non numeric", args: []string{"abc"}, kind: ParseErrorInvalidPort},
		{name: "out of range", args: []string{"65536"}, kind: ParseErrorInvalidPort},
		{name: "unknown command", args: []string{"wat", "8080"}, kind: ParseErrorUnknownCommand},
		{name: "free without port", args: []string{"free"}, kind: ParseErrorFree},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)
			if err == nil {
				t.Fatal("Parse() error = nil")
			}
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if parseErr.Kind != tt.kind {
				t.Fatalf("error kind = %d, want %d", parseErr.Kind, tt.kind)
			}
		})
	}
}

func TestParseHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--version"}} {
		got, err := Parse(args)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", args, err)
		}
		if got.Action != ActionHelp && got.Action != ActionVersion {
			t.Fatalf("Parse(%v) action = %v", args, got.Action)
		}
	}
}

func TestRunReturnsExitCodeWithoutExit(t *testing.T) {
	var stderr strings.Builder
	if code := Run(context.Background(), []string{"not-a-port"}, Dependencies{}, nil, &stderr); code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: portwatch [flags] [port]") || !strings.Contains(stderr.String(), "portwatch tui [port]") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
	deps := Dependencies{
		Scanner: fakeRootScanner{},
		Manager: fakeRootManager{},
	}
	if code := Run(context.Background(), nil, deps, nil, &stderr); code != 0 {
		t.Fatalf("Run() empty args code = %d, want 0", code)
	}
}

func TestParseFreeCommand(t *testing.T) {
	got, err := Parse([]string{"free", "8080"})
	if err != nil {
		t.Fatalf("Parse(free) error = %v", err)
	}
	if got.Action != ActionFree || got.Port != 8080 {
		t.Fatalf("Parse(free) = %#v, want ActionFree port 8080", got)
	}
}

func TestParseTUICommand(t *testing.T) {
	got, err := Parse([]string{"tui"})
	if err != nil {
		t.Fatalf("Parse(tui) error = %v", err)
	}
	if got.Action != ActionTUI {
		t.Fatalf("Parse(tui) action = %v, want ActionTUI", got.Action)
	}
	focused, err := Parse([]string{"tui", "8080"})
	if err != nil {
		t.Fatalf("Parse(tui 8080) error = %v", err)
	}
	if focused.Action != ActionTUI || focused.Port != 8080 {
		t.Fatalf("Parse(tui 8080) = %#v, want ActionTUI port 8080", focused)
	}
	for _, value := range []string{"0", "65536", "http"} {
		if _, err := Parse([]string{"tui", value}); err == nil {
			t.Errorf("Parse(tui %s) error = nil", value)
		}
	}
}

func TestRunRejectsNonTCPTUI(t *testing.T) {
	var stderr strings.Builder
	deps := Dependencies{Scanner: fakeRootScanner{}, Manager: fakeRootManager{}}
	if code := Run(context.Background(), []string{"--protocol", "udp", "tui"}, deps, nil, &stderr); code != 2 {
		t.Fatalf("Run(udp tui) code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "TCP listening ports only") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestParseFindJoinsQueryWords(t *testing.T) {
	got, err := Parse([]string{"find", "spring", "boot"})
	if err != nil {
		t.Fatalf("Parse(find) error = %v", err)
	}
	if got.Action != ActionFind || got.Query != "spring boot" {
		t.Fatalf("Parse(find) = %#v", got)
	}
}

func TestParseWatchPort(t *testing.T) {
	got, err := Parse([]string{"watch", "8080"})
	if err != nil {
		t.Fatalf("Parse(watch) error = %v", err)
	}
	if got.Action != ActionWatch || got.Port != 8080 {
		t.Fatalf("Parse(watch) = %#v", got)
	}
}

func TestParsePortRange(t *testing.T) {
	got, err := Parse([]string{"3000-4000"})
	if err != nil {
		t.Fatalf("Parse(range) error = %v", err)
	}
	if got.Action != ActionPortRange || got.Port != 3000 || got.PortEnd != 4000 {
		t.Fatalf("Parse(range) = %#v", got)
	}
}

func TestParsePortRangeRejectsInvalidBounds(t *testing.T) {
	for _, value := range []string{"0-80", "80-0", "4000-3000", "80-65536", "80-"} {
		if _, err := Parse([]string{value}); err == nil {
			t.Errorf("Parse(%q) error = nil", value)
		}
	}
}

func TestParsePortSetSortsAndDeduplicates(t *testing.T) {
	got, err := Parse([]string{"--ports", "8080,3000,8080"})
	if err != nil {
		t.Fatalf("Parse(--ports) error = %v", err)
	}
	if got.Action != ActionPortSet || !reflect.DeepEqual(got.Ports, []int{3000, 8080}) {
		t.Fatalf("Parse(--ports) = %#v", got)
	}
}

func TestParsePortSetRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", "0,80", "80,65536", "80,wat"} {
		args := []string{"--ports", value}
		if value == "" {
			args = []string{"--ports="}
		}
		if _, err := Parse(args); err == nil {
			t.Errorf("Parse(%v) error = nil", args)
		}
	}
}

func TestParseProtocolNormalizesSupportedValue(t *testing.T) {
	got, err := Parse([]string{"--protocol", "UDP", "8080"})
	if err != nil {
		t.Fatalf("Parse(protocol) error = %v", err)
	}
	if got.Flags.Protocol != "udp" {
		t.Fatalf("protocol = %q, want udp", got.Flags.Protocol)
	}
}

func TestRunJSONDoesNotWriteTableOnProcessInfoError(t *testing.T) {
	deps := Dependencies{
		Scanner: onePortScanner{record: model.PortInfo{Port: 8080, Protocol: "TCP", PID: 42}},
		Manager: failingInfoManager{},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code == ExitSuccess {
		t.Fatal("run() code = success, want error")
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") || strings.Contains(stdout.String(), "PORT PROTOCOL") {
		t.Fatalf("stdout = %q, want JSON without table", stdout.String())
	}
	if !strings.Contains(stderr.String(), "operation failed") {
		t.Fatalf("stderr = %q, want error", stderr.String())
	}
}

func TestRunRejectsZeroWatchInterval(t *testing.T) {
	deps := Dependencies{Scanner: fakeRootScanner{}, Manager: fakeRootManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--interval", "0s", "watch"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSystem {
		t.Fatalf("run() code = %d, want %d", code, ExitSystem)
	}
	if !strings.Contains(stderr.String(), "watch interval must be positive") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunFreeJSONKeepsStdoutAsJSON(t *testing.T) {
	deps := Dependencies{Scanner: &freeScanner{}, Manager: &freeManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "free", "8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d, want %d", code, ExitSuccess)
	}
	var response model.FreeResponse
	if err := json.Unmarshal([]byte(stdout.String()), &response); err != nil {
		t.Fatalf("stdout JSON error = %v; stdout=%q; stderr=%q", err, stdout.String(), stderr.String())
	}
	if response.Status != "available" || response.Port != 8080 {
		t.Fatalf("response = %+v", response)
	}
}

type rangeScanner struct {
	ports []model.PortInfo
}

func (s rangeScanner) List(context.Context) ([]model.PortInfo, error)      { return s.ports, nil }
func (s rangeScanner) Port(context.Context, int) ([]model.PortInfo, error) { return nil, nil }

type protocolRootScanner struct {
	ports []model.PortInfo
}

func (s protocolRootScanner) List(context.Context) ([]model.PortInfo, error) { return s.ports, nil }
func (s protocolRootScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return s.ports, nil
}
func (s protocolRootScanner) ListProtocol(context.Context, string) ([]model.PortInfo, error) {
	return s.ports, nil
}
func (s protocolRootScanner) PortProtocol(context.Context, int, string) ([]model.PortInfo, error) {
	return s.ports, nil
}

func TestRunProtocolUsesOptionalScanner(t *testing.T) {
	deps := Dependencies{
		Scanner: protocolRootScanner{ports: []model.PortInfo{{Port: 5353, Protocol: "UDP", State: "BOUND", PID: 7}}},
		Manager: fakeRootManager{},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "--protocol", "udp"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"protocol":"UDP"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunProtocolReportsUnsupportedScanner(t *testing.T) {
	deps := Dependencies{Scanner: fakeRootScanner{}, Manager: fakeRootManager{}}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--protocol", "udp"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSystem || !strings.Contains(stderr.String(), "not supported") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

type countingInfoManager struct{ calls int }

func (m *countingInfoManager) Info(context.Context, int) (model.ProcessInfo, error) {
	m.calls++
	return model.ProcessInfo{Name: "demo.exe"}, nil
}
func (m *countingInfoManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (m *countingInfoManager) Terminate(context.Context, int) error      { return nil }

func TestRunCachesProcessInfoByPID(t *testing.T) {
	manager := &countingInfoManager{}
	deps := Dependencies{
		Scanner: rangeScanner{ports: []model.PortInfo{{Port: 5353, PID: 7}, {Port: 5353, PID: 7}}},
		Manager: manager,
	}
	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"--json", "--ports", "5353"}, deps, strings.NewReader(""), &stdout, &stderr); code != ExitSuccess {
		t.Fatalf("run() code = %d stderr=%q", code, stderr.String())
	}
	if manager.calls != 1 {
		t.Fatalf("Info() calls = %d, want one lookup per PID", manager.calls)
	}
}

func TestRunPortRangeJSONFiltersAndSorts(t *testing.T) {
	deps := Dependencies{
		Scanner: rangeScanner{ports: []model.PortInfo{{Port: 8080, PID: 2}, {Port: 3000, PID: 1}, {Port: 9000, PID: 3}}},
		Manager: fakeRootManager{},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "3000-8080"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d, stderr=%q", code, stderr.String())
	}
	var response model.PortsResponse
	if err := json.Unmarshal([]byte(stdout.String()), &response); err != nil {
		t.Fatalf("stdout JSON error = %v; stdout=%q", err, stdout.String())
	}
	if len(response.Ports) != 2 || response.Ports[0].Port != 3000 || response.Ports[1].Port != 8080 {
		t.Fatalf("ports = %+v", response.Ports)
	}
}

func TestRunPortSetJSONFiltersAndSorts(t *testing.T) {
	deps := Dependencies{
		Scanner: rangeScanner{ports: []model.PortInfo{{Port: 8080, PID: 2}, {Port: 3000, PID: 1}, {Port: 9000, PID: 3}}},
		Manager: fakeRootManager{},
	}
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"--json", "--ports", "8080,3000"}, deps, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("run() code = %d, stderr=%q", code, stderr.String())
	}
	var response model.PortsResponse
	if err := json.Unmarshal([]byte(stdout.String()), &response); err != nil {
		t.Fatalf("stdout JSON error = %v; stdout=%q", err, stdout.String())
	}
	if len(response.Ports) != 2 || response.Ports[0].Port != 3000 || response.Ports[1].Port != 8080 {
		t.Fatalf("ports = %+v", response.Ports)
	}
}

type fakeRootScanner struct{}

func (fakeRootScanner) List(context.Context) ([]model.PortInfo, error) { return nil, nil }
func (fakeRootScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return nil, nil
}

type fakeRootManager struct{}

func (fakeRootManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, nil
}
func (fakeRootManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (fakeRootManager) Terminate(context.Context, int) error      { return nil }

type onePortScanner struct{ record model.PortInfo }

func (s onePortScanner) List(context.Context) ([]model.PortInfo, error) {
	return []model.PortInfo{s.record}, nil
}
func (s onePortScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return []model.PortInfo{s.record}, nil
}

type failingInfoManager struct{}

func (failingInfoManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, errors.New("metadata unavailable")
}
func (failingInfoManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (failingInfoManager) Terminate(context.Context, int) error      { return nil }

func TestReportProcessInfoErrorsIsDeterministic(t *testing.T) {
	errorsByPID := map[int]error{
		5:  errors.New("failure for pid five"),
		2:  errors.New("failure for pid two"),
		9:  errors.New("failure for pid nine"),
		20: errors.New("failure for pid twenty"),
	}
	var out strings.Builder
	for i := 0; i < 100; i++ {
		out.Reset()
		reportProcessInfoErrors(&out, errorsByPID)
		got := out.String()
		if !strings.Contains(got, "failure for pid two") {
			t.Fatalf("reportProcessInfoErrors() = %q, want lowest-PID error", got)
		}
		for _, unexpected := range []string{"pid five", "pid nine", "pid twenty"} {
			if strings.Contains(got, unexpected) {
				t.Fatalf("reportProcessInfoErrors() = %q, must not surface %q", got, unexpected)
			}
		}
	}
}

type countingPerPIDManager struct {
	mu    sync.Mutex
	calls map[int]int
}

func (m *countingPerPIDManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[pid]++
	return model.ProcessInfo{PID: pid, Name: "proc"}, nil
}
func (m *countingPerPIDManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (m *countingPerPIDManager) Terminate(context.Context, int) error      { return nil }

func TestResolveProcessInfosQueriesEachPIDOnce(t *testing.T) {
	manager := &countingPerPIDManager{calls: map[int]int{}}
	records := []model.PortInfo{
		{Port: 80, PID: 11}, {Port: 443, PID: 22}, {Port: 8080, PID: 11},
		{Port: 3000, PID: 33}, {Port: 3001, PID: 22},
	}
	infos, errorsByPID := resolveProcessInfos(context.Background(), manager, records)
	for _, pid := range []int{11, 22, 33} {
		if manager.calls[pid] != 1 {
			t.Fatalf("Info() calls for pid %d = %d, want exactly 1", pid, manager.calls[pid])
		}
		info, ok := infos[pid]
		if !ok || info.PID != pid || info.Name != "proc" {
			t.Fatalf("infos[%d] = %+v, ok=%v", pid, info, ok)
		}
	}
	if len(errorsByPID) != 0 {
		t.Fatalf("errorsByPID = %v, want none", errorsByPID)
	}
}

// barrierInfoManager blocks each Info call until the test observes two
// concurrent calls, proving resolveProcessInfos actually runs in parallel.
type barrierInfoManager struct {
	entered chan struct{}
	release chan struct{}
}

func (m *barrierInfoManager) Info(_ context.Context, pid int) (model.ProcessInfo, error) {
	m.entered <- struct{}{}
	<-m.release
	return model.ProcessInfo{PID: pid, Name: "proc"}, nil
}
func (m *barrierInfoManager) Exists(context.Context, int) (bool, error) { return false, nil }
func (m *barrierInfoManager) Terminate(context.Context, int) error      { return nil }

func TestResolveProcessInfosRunsConcurrently(t *testing.T) {
	manager := &barrierInfoManager{entered: make(chan struct{}), release: make(chan struct{})}
	records := []model.PortInfo{{Port: 1, PID: 101}, {Port: 2, PID: 102}}
	type result struct {
		infos       map[int]model.ProcessInfo
		errorsByPID map[int]error
	}
	done := make(chan result, 1)
	go func() {
		infos, errorsByPID := resolveProcessInfos(context.Background(), manager, records)
		done <- result{infos: infos, errorsByPID: errorsByPID}
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-manager.entered:
		case <-time.After(5 * time.Second):
			t.Fatal("second concurrent Info call never started; resolveProcessInfos appears serial")
		}
	}
	close(manager.release)
	select {
	case got := <-done:
		if len(got.infos) != 2 || len(got.errorsByPID) != 0 {
			t.Fatalf("infos=%v errors=%v", got.infos, got.errorsByPID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("resolveProcessInfos did not finish after release")
	}
}
