package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseUninstall(t *testing.T) {
	command, err := Parse([]string{"uninstall"})
	if err != nil || command.Action != ActionUninstall {
		t.Fatalf("Parse(uninstall) = (%+v, %v), want ActionUninstall", command, err)
	}
	command, err = Parse([]string{"--yes", "uninstall"})
	if err != nil || command.Action != ActionUninstall || !command.Flags.Yes {
		t.Fatalf("Parse(--yes uninstall) = (%+v, %v), want ActionUninstall with --yes", command, err)
	}
	command, err = Parse([]string{"uninstall", "--yes"})
	if err != nil || command.Action != ActionUninstall || !command.Flags.Yes {
		t.Fatalf("Parse(uninstall --yes) = (%+v, %v), want interspersed --yes", command, err)
	}
	_, err = Parse([]string{"uninstall", "extra"})
	var parseErr *ParseError
	if !errors.As(err, &parseErr) || parseErr.Message != "uninstall does not take arguments" {
		t.Fatalf("Parse(uninstall extra) error = %v, want argument rejection", err)
	}
}

func TestRemovePathEntry(t *testing.T) {
	sep := string(os.PathListSeparator)
	bin := strings.Join([]string{"C:", "Users", "u", "bin"}, string(filepath.Separator))
	for _, tc := range []struct {
		name    string
		path    string
		want    string
		removed bool
	}{
		{"removes matching entry", strings.Join([]string{"C:\\a", bin, "C:\\b"}, sep), strings.Join([]string{"C:\\a", "C:\\b"}, sep), true},
		{"keeps path when entry absent", strings.Join([]string{"C:\\a", "C:\\b"}, sep), strings.Join([]string{"C:\\a", "C:\\b"}, sep), false},
		{"matches case-insensitively", strings.Join([]string{"C:\\A", strings.ToUpper(bin)}, sep), "C:\\A", true},
		{"matches trailing separator", strings.Join([]string{bin + string(filepath.Separator), "C:\\a"}, sep), "C:\\a", true},
		{"preserves doubled separators elsewhere", strings.Join([]string{"C:\\a", "", bin}, sep), "C:\\a" + sep, true},
		{"preserves trailing separator", strings.Join([]string{"C:\\a", bin, ""}, sep), "C:\\a" + sep, true},
	} {
		got, removed := removePathEntry(tc.path, bin)
		if got != tc.want || removed != tc.removed {
			t.Fatalf("%s: removePathEntry(%q, %q) = (%q, %v), want (%q, %v)", tc.name, tc.path, bin, got, removed, tc.want, tc.removed)
		}
	}
}

// uninstallRecorder collects hook interactions. Tests read it through the
// pointer because the closures append after the installer returns.
type uninstallRecorder struct {
	removed []string
	cleaned []string
}

// installUninstallHooks points the uninstall hooks at a fake installation
// (executable path and user home) and records interactions.
func installUninstallHooks(t *testing.T, executable, home string) *uninstallRecorder {
	t.Helper()
	recorder := &uninstallRecorder{}
	originalSelf, originalHome := locateSelf, userHomeDir
	originalRemove, originalClean := removeBinary, cleanUserPathEntry
	locateSelf = func() (string, error) { return executable, nil }
	userHomeDir = func() (string, error) { return home, nil }
	removeBinary = func(path string) error {
		recorder.removed = append(recorder.removed, path)
		return os.Remove(path)
	}
	cleanUserPathEntry = func(dir string, out io.Writer) error {
		recorder.cleaned = append(recorder.cleaned, dir)
		fmt.Fprintf(out, "cleaned %s\n", dir)
		return nil
	}
	t.Cleanup(func() {
		locateSelf, userHomeDir = originalSelf, originalHome
		removeBinary, cleanUserPathEntry = originalRemove, originalClean
	})
	return recorder
}

func writeFakeBinary(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallCancelsWithoutConfirmation(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	recorder := installUninstallHooks(t, executable, home)

	var out strings.Builder
	err := Uninstall(false, strings.NewReader("\n"), &out)
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("Uninstall() error = %v, want ErrUserCancelled", err)
	}
	if !strings.Contains(out.String(), "Cancelled.") {
		t.Fatalf("output = %q, want cancellation notice", out.String())
	}
	if _, statErr := os.Stat(executable); statErr != nil {
		t.Fatal("binary was deleted despite cancellation")
	}
	if len(recorder.removed) != 0 || len(recorder.cleaned) != 0 {
		t.Fatalf("removed=%v cleaned=%v, want no interactions after cancel", recorder.removed, recorder.cleaned)
	}
}

func TestUninstallConfirmedRemovesBinaryAndCleansPath(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	recorder := installUninstallHooks(t, executable, home)

	var out strings.Builder
	err := Uninstall(false, strings.NewReader("y\n"), &out)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(recorder.removed) != 1 || recorder.removed[0] != executable {
		t.Fatalf("removed = %v, want [%s]", recorder.removed, executable)
	}
	if _, statErr := os.Stat(executable); !os.IsNotExist(statErr) {
		t.Fatalf("binary still present: %v", statErr)
	}
	if len(recorder.cleaned) != 1 || recorder.cleaned[0] != binDir {
		t.Fatalf("cleaned = %v, want [%s]", recorder.cleaned, binDir)
	}
}

func TestUninstallYesSkipsPrompt(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	recorder := installUninstallHooks(t, executable, home)

	var out strings.Builder
	if err := Uninstall(true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Uninstall(--yes) error = %v", err)
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("output = %q, --yes must not prompt", out.String())
	}
	if len(recorder.cleaned) != 1 {
		t.Fatalf("cleaned = %v, want the empty default bin dir", recorder.cleaned)
	}
}

func TestUninstallLeavesNonEmptyDirectoryPathAlone(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	writeFakeBinary(t, filepath.Join(binDir, "other.exe"))
	recorder := installUninstallHooks(t, executable, home)

	var out strings.Builder
	if err := Uninstall(true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(recorder.cleaned) != 0 {
		t.Fatalf("cleaned = %v, want PATH untouched for a non-empty directory", recorder.cleaned)
	}
	if !strings.Contains(out.String(), "not empty") {
		t.Fatalf("output = %q, want leftover-directory notice", out.String())
	}
}

func TestUninstallLeavesNonDefaultDirectoryPathAlone(t *testing.T) {
	home := t.TempDir()
	otherDir := filepath.Join(home, "tools")
	executable := filepath.Join(otherDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	recorder := installUninstallHooks(t, executable, home)

	var out strings.Builder
	if err := Uninstall(true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(recorder.cleaned) != 0 {
		t.Fatalf("cleaned = %v, want PATH untouched outside the default directory", recorder.cleaned)
	}
	if !strings.Contains(out.String(), "left in place") {
		t.Fatalf("output = %q, want leftover-directory notice", out.String())
	}
}

func TestRunUninstallExitCodes(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	executable := filepath.Join(binDir, "portwatch.exe")
	writeFakeBinary(t, executable)
	_ = installUninstallHooks(t, executable, home)

	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"uninstall"}, Dependencies{}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess || !strings.Contains(stdout.String(), "Cancelled.") {
		t.Fatalf("cancel: code=%d stdout=%q stderr=%q, want 0 with cancellation notice", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(), []string{"uninstall", "extra"}, Dependencies{}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitArguments {
		t.Fatalf("arguments: code=%d, want %d", code, ExitArguments)
	}

	originalRemove := removeBinary
	removeBinary = func(string) error {
		return fmt.Errorf("remove: %w", errUninstallBlocked)
	}
	t.Cleanup(func() { removeBinary = originalRemove })
	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(), []string{"--yes", "uninstall"}, Dependencies{}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitPermission {
		t.Fatalf("blocked: code=%d stderr=%q, want %d", code, stderr.String(), ExitPermission)
	}
	if !strings.Contains(stderr.String(), "in use") {
		t.Fatalf("blocked: stderr = %q, want file-in-use message", stderr.String())
	}
}
