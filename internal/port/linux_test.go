//go:build linux

package port

import (
	"fmt"
	"github.com/MY-Final/portWatch/pkg/model"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const procTableHeader = "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode"

func writeProcTable(t *testing.T, path string, rows ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	lines := append([]string{procTableHeader}, rows...)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// procTCPRow renders one /proc/net/tcp line; field 9 carries the socket inode.
func procTCPRow(sequence int, address string, port int, state, inode string) string {
	return fmt.Sprintf("  %2d: %s:%04X 00000000:0000 %s 00000000:00000000 00:00000000 00000000     0        0 %s 1 0000000000000000 100 0 0 10 0",
		sequence, address, port, state, inode)
}

func fakeProcessFDs(t *testing.T, root, pid string, links map[string]string) {
	t.Helper()
	fdDir := filepath.Join(root, pid, "fd")
	if err := os.MkdirAll(fdDir, 0o755); err != nil {
		t.Fatalf("create %s: %v", fdDir, err)
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(fdDir, name)); err != nil {
			t.Fatalf("symlink %s/%s: %v", fdDir, name, err)
		}
	}
}

func TestListProcNetResolvesPIDsFromInodeMap(t *testing.T) {
	root := t.TempDir()
	writeProcTable(t, filepath.Join(root, "net", "tcp"),
		procTCPRow(0, "0100007F", 8080, "0A", "1001"),
		procTCPRow(1, "0100007F", 9090, "0A", "1002"),
		procTCPRow(2, "0100007F", 3000, "0A", "1003"),
		procTCPRow(3, "0100007F", 1500, "01", "1004"), // ESTABLISHED, not a listener
	)
	writeProcTable(t, filepath.Join(root, "net", "tcp6"),
		procTCPRow(0, "00000000000000000000000000000000", 5000, "0A", "1005"),
	)
	fakeProcessFDs(t, root, "100", map[string]string{"0": "socket:[1001]", "1": "pipe:[12]"})
	fakeProcessFDs(t, root, "200", map[string]string{"3": "socket:[1002]", "5": "socket:[1005]"})
	fakeProcessFDs(t, root, "300", map[string]string{"0": "socket:[1003]"})
	fakeProcessFDs(t, root, "self", map[string]string{"0": "socket:[9999]"}) // non-numeric dir, ignored
	if err := os.MkdirAll(filepath.Join(root, "400"), 0o755); err != nil {   // process without fd dir, ignored
		t.Fatal(err)
	}

	rows, err := listProcNet(root, os.ReadDir, "tcp")
	if err != nil {
		t.Fatalf("listProcNet() error = %v", err)
	}
	want := []struct {
		port  int
		pid   int
		local string
	}{
		{3000, 300, "127.0.0.1"},
		{5000, 200, "::"},
		{8080, 100, "127.0.0.1"},
		{9090, 200, "127.0.0.1"},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %+v, want %d rows", rows, len(want))
	}
	for i, expected := range want {
		if rows[i].Port != expected.port || rows[i].PID != expected.pid || rows[i].LocalAddr != expected.local {
			t.Fatalf("rows[%d] = %+v, want port=%d pid=%d local=%s", i, rows[i], expected.port, expected.pid, expected.local)
		}
		if rows[i].State != "LISTENING" || rows[i].Protocol != "TCP" {
			t.Fatalf("rows[%d] state/protocol = %s/%s, want LISTENING/TCP", i, rows[i].State, rows[i].Protocol)
		}
	}
}

func TestListProcNetScansProcRootOnceForManyRows(t *testing.T) {
	root := t.TempDir()
	rows := make([]string, 0, 32)
	for i := 0; i < 32; i++ {
		rows = append(rows, procTCPRow(i, "0100007F", 20000+i, "0A", fmt.Sprintf("%d", 2000+i)))
	}
	writeProcTable(t, filepath.Join(root, "net", "tcp"), rows...)
	fakeProcessFDs(t, root, "100", map[string]string{"0": "socket:[2000]"})

	var rootScans int
	got, err := listProcNet(root, func(path string) ([]os.DirEntry, error) {
		if path == root {
			rootScans++
		}
		return os.ReadDir(path)
	}, "tcp")
	if err != nil {
		t.Fatalf("listProcNet() error = %v", err)
	}
	if rootScans != 1 {
		t.Fatalf("/proc root traversals = %d, want exactly 1 for 32 listening rows", rootScans)
	}
	if len(got) != 32 {
		t.Fatalf("rows = %d, want 32", len(got))
	}
	if got[0].PID != 100 {
		t.Fatalf("rows[0].PID = %d, want 100", got[0].PID)
	}
}

func TestListProcNetToleratesMissingTables(t *testing.T) {
	rows, err := listProcNet(t.TempDir(), os.ReadDir, "tcp")
	if err != nil {
		t.Fatalf("listProcNet() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows = %+v, want none", rows)
	}
}

func TestSocketInode(t *testing.T) {
	for _, tc := range []struct {
		link string
		want string
		ok   bool
	}{
		{"socket:[12345]", "12345", true},
		{"socket:[]", "", false},
		{"pipe:[12]", "", false},
		{"socket", "", false},
		{"socket:[12345] extra", "", false},
	} {
		inode, ok := socketInode(tc.link)
		if inode != tc.want || ok != tc.ok {
			t.Fatalf("socketInode(%q) = (%q, %v), want (%q, %v)", tc.link, inode, ok, tc.want, tc.ok)
		}
	}
}

func TestListProcNetUDP(t *testing.T) {
	root := t.TempDir()
	writeProcTable(t, filepath.Join(root, "net", "udp"),
		procUDPRow(0, "0100007F", 5353, "2001"),
		procUDPRow(1, "00000000", 68, "2002"),
	)
	writeProcTable(t, filepath.Join(root, "net", "udp6"),
		procUDPRow(0, "00000000000000000000000001000000", 546, "2003"),
	)
	fakeProcessFDs(t, root, "500", map[string]string{"0": "socket:[2001]"})
	fakeProcessFDs(t, root, "600", map[string]string{"1": "socket:[2002]", "2": "socket:[2003]"})

	rows, err := listProcNet(root, os.ReadDir, "udp")
	if err != nil {
		t.Fatalf("listProcNet(udp) error = %v", err)
	}
	want := []struct {
		port  int
		pid   int
		local string
	}{
		{68, 600, "0.0.0.0"},
		{5353, 500, "127.0.0.1"},
		{546, 600, "::1"},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %+v, want %d rows", rows, len(want))
	}
	for i, expected := range want {
		if rows[i].Port != expected.port || rows[i].PID != expected.pid || rows[i].LocalAddr != expected.local {
			t.Fatalf("rows[%d] = %+v, want port=%d pid=%d local=%s", i, rows[i], expected.port, expected.pid, expected.local)
		}
		if rows[i].Protocol != "UDP" || rows[i].State != "BOUND" {
			t.Fatalf("rows[%d] protocol/state = %s/%s, want UDP/BOUND", i, rows[i].Protocol, rows[i].State)
		}
	}
}

// procUDPRow renders one /proc/net/udp line; field 9 carries the inode.
func procUDPRow(sequence int, address string, port int, inode string) string {
	return fmt.Sprintf("  %2d: %s:%04X 00000000:0000 07 00000000:00000000 00:00000000 00000000     0        0 %s 1 0000000000000000 100 0 0 10 0",
		sequence, address, port, inode)
}

func TestParseProcTCPAllStates(t *testing.T) {
	root := t.TempDir()
	writeProcTable(t, filepath.Join(root, "net", "tcp"),
		procTCPRow(0, "0100007F", 8080, "0A", "1001"),
		procTCPRow(1, "0100007F", 4444, "01", "1002"),
		procTCPRow(2, "0100007F", 5555, "06", "1003"),
		procTCPRow(3, "0100007F", 6666, "08", "1004"),
		procTCPRow(4, "0100007F", 7777, "99", "1005"),
	)
	inodePIDs := map[string]int{"1001": 1, "1002": 2, "1003": 3, "1004": 4, "1005": 5}

	listeners, err := parseProcTCP(filepath.Join(root, "net", "tcp"), inodePIDs, false)
	if err != nil {
		t.Fatalf("parseProcTCP(listeners) error = %v", err)
	}
	if len(listeners) != 1 || listeners[0].Port != 8080 || listeners[0].State != "LISTENING" {
		t.Fatalf("listeners = %+v, want only the 0A row", listeners)
	}

	all, err := parseProcTCP(filepath.Join(root, "net", "tcp"), inodePIDs, true)
	if err != nil {
		t.Fatalf("parseProcTCP(all) error = %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("rows = %d, want all 5", len(all))
	}
	byPort := map[int]model.PortInfo{}
	for _, row := range all {
		byPort[row.Port] = row
	}
	if byPort[4444].State != "ESTABLISHED" || byPort[4444].RemoteAddr == "" {
		t.Fatalf("4444 = %+v, want ESTABLISHED with a remote address", byPort[4444])
	}
	if byPort[5555].State != "TIME_WAIT" {
		t.Fatalf("5555 state = %s, want TIME_WAIT", byPort[5555].State)
	}
	if byPort[6666].State != "CLOSE_WAIT" {
		t.Fatalf("6666 state = %s, want CLOSE_WAIT", byPort[6666].State)
	}
	if byPort[7777].State != "UNKNOWN" {
		t.Fatalf("7777 state = %s, want UNKNOWN for an unmapped code", byPort[7777].State)
	}
	if byPort[4444].PID != 2 {
		t.Fatalf("4444 pid = %d, want 2 from the inode map", byPort[4444].PID)
	}
}

func TestTCPStateNamesComplete(t *testing.T) {
	for code, name := range tcpStateNames {
		if strings.TrimSpace(name) == "" {
			t.Fatalf("state %q has an empty name", code)
		}
	}
	if tcpStateNames["0A"] != "LISTENING" || tcpStateNames["01"] != "ESTABLISHED" {
		t.Fatalf("core state names changed: %+v", tcpStateNames)
	}
}
