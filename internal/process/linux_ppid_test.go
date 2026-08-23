//go:build linux

package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStatusParentPID(t *testing.T) {	dir := t.TempDir()
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

func TestReadCwd(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "project")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "cwd")); err != nil {
		t.Fatal(err)
	}
	if got := readCwd(dir); got != target {
		t.Fatalf("readCwd() = %q, want %q", got, target)
	}
	if got := readCwd(filepath.Join(dir, "missing")); got != "" {
		t.Fatalf("readCwd(missing) = %q, want empty", got)
	}
}
