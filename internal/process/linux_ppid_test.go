//go:build linux

package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStatusParentPID(t *testing.T) {
	dir := t.TempDir()
	status := filepath.Join(dir, "status")
	content := "Name:\tnode\nUmask:\t0022\nState:\tS (sleeping)\nPPid:\t1200\nPid:\t1234\n"
	if err := os.WriteFile(status, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readStatusParentPID(status); got != 1200 {
		t.Fatalf("readStatusParentPID() = %d, want 1200", got)
	}
	if got := readStatusParentPID(filepath.Join(dir, "missing")); got != 0 {
		t.Fatalf("readStatusParentPID(missing) = %d, want 0", got)
	}
	if err := os.WriteFile(status, []byte("Name:\tnode\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readStatusParentPID(status); got != 0 {
		t.Fatalf("readStatusParentPID(no PPid) = %d, want 0", got)
	}
	if err := os.WriteFile(status, []byte("PPid:\tnot-a-number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readStatusParentPID(status); got != 0 {
		t.Fatalf("readStatusParentPID(garbage) = %d, want 0", got)
	}
}
