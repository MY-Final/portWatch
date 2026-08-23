//go:build windows

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTransitionalArtifactForOS(t *testing.T) {
	for _, tc := range []struct {
		name string
		want bool
	}{
		{"portwatch.uninstalling.exe", true},
		{"pw.uninstalling.exe", true},
		{"portwatch-uninstall-2851373902.cmd", true},
		{"portwatch-uninstall-x.cmd", true},
		{"portwatch-uninstall-x.cmd.bak", false},
		{"portwatch-uninstall-2851373902.bat", false},
		{"uninstalling.exe", false},
		{"portwatch.exe", false},
		{"portwatch-uninstall.exe", false},
	} {
		if got := isTransitionalArtifactForOS(tc.name); got != tc.want {
			t.Fatalf("isTransitionalArtifactForOS(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// stagedRemoval deletes the target but leaves the two transitional files the
// real delayed self-delete still has in the directory when removeBinary
// returns, reproducing the conditions of the PATH-cleanup regression.
func stagedRemoval(t *testing.T, binDir string) func(string) error {
	t.Helper()
	return func(path string) error {
		if err := os.Remove(path); err != nil {
			return err
		}
		writeFakeBinary(t, filepath.Join(binDir, "portwatch.uninstalling.exe"))
		writeFakeBinary(t, filepath.Join(binDir, "portwatch-uninstall-x.cmd"))
		return nil
	}
}

func TestUninstallCleansPathDespiteTransitionalArtifacts(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	recorder := installUninstallHooks(t, executable, home)
	removeBinary = stagedRemoval(t, binDir)

	var out strings.Builder
	if err := Uninstall(true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(recorder.cleaned) != 1 || recorder.cleaned[0] != binDir {
		t.Fatalf("cleaned = %v, want [%s] despite transitional artifacts", recorder.cleaned, binDir)
	}
	if strings.Contains(out.String(), "not empty") {
		t.Fatalf("output = %q, transitional artifacts must not be reported as leftovers", out.String())
	}
}

func TestUninstallKeepsPathWhenRealFilesRemain(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	writeFakeBinary(t, filepath.Join(binDir, "other.txt"))
	recorder := installUninstallHooks(t, executable, home)
	removeBinary = stagedRemoval(t, binDir)

	var out strings.Builder
	if err := Uninstall(true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(recorder.cleaned) != 0 {
		t.Fatalf("cleaned = %v, want PATH untouched while other.txt remains", recorder.cleaned)
	}
	if !strings.Contains(out.String(), "not empty") {
		t.Fatalf("output = %q, want the leftover-file notice", out.String())
	}
}
