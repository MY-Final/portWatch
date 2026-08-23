package command

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// errUninstallBlocked marks a self-delete the OS refused because the binary
// is still in use or the install directory denies access.
var errUninstallBlocked = errors.New("binary is in use or access is denied")

// Hooks overridable in tests so uninstall logic can run against a fake
// installation under t.TempDir instead of the real executable and user PATH.
var (
	locateSelf         = os.Executable
	userHomeDir        = os.UserHomeDir
	removeBinary       = removeBinaryForOS
	cleanUserPathEntry = cleanUserPathForOS
)

// Uninstall confirms and deletes the running portwatch binary, then applies
// the conservative PATH policy: only when the binary lived in the default
// per-user install directory and that directory is now empty is the entry
// removed from the user PATH. Any other location is left untouched.
func Uninstall(yes bool, in io.Reader, out io.Writer) error {
	if in == nil || out == nil {
		return errors.New("uninstall dependencies are nil")
	}
	exePath, err := locateSelf()
	if err != nil {
		return fmt.Errorf("locate running binary: %w", err)
	}
	if !yes {
		_, _ = fmt.Fprintf(out, "Executable: %s\n", exePath)
		_, _ = fmt.Fprintf(out, "Uninstall %s? [y/N] ", exePath)
		answer, readErr := bufio.NewReader(in).ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return fmt.Errorf("read confirmation: %w", readErr)
		}
		if !isConfirmation(answer) {
			_, _ = fmt.Fprintln(out, "Cancelled.")
			return ErrUserCancelled
		}
	}
	if err := removeBinary(exePath); err != nil {
		return fmt.Errorf("remove %s: %w", exePath, err)
	}
	_, _ = fmt.Fprintf(out, "Removed %s.\n", exePath)
	return cleanInstallDirectory(filepath.Dir(exePath), out)
}

// cleanInstallDirectory applies the PATH policy for the directory the binary
// was deleted from. The default install directory is %USERPROFILE%\bin on
// Windows and $HOME/.local/bin elsewhere.
func cleanInstallDirectory(dir string, out io.Writer) error {
	home, err := userHomeDir()
	if err != nil {
		home = ""
	}
	if home == "" || !samePath(dir, defaultInstallDir(home)) {
		_, _ = fmt.Fprintf(out, "Note: install directory %s was left in place and the user PATH was not modified.\n", dir)
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		_, _ = fmt.Fprintf(out, "Note: %s is not empty, so the user PATH was not modified.\n", dir)
		return nil
	}
	return cleanUserPathEntry(dir, out)
}

// defaultInstallDir mirrors the directories used by the install scripts.
func defaultInstallDir(home string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "bin")
	}
	return filepath.Join(home, ".local", "bin")
}

// removePathEntry drops every entry of pathValue that names dir and reports
// whether anything was removed. Entries are matched case-insensitively and
// ignoring a trailing separator, which keeps the cleanup safe for the single
// directory this uninstall may touch. Other entries, including empty ones
// from doubled separators, are preserved verbatim.
func removePathEntry(pathValue, dir string) (string, bool) {
	separator := string(os.PathListSeparator)
	parts := strings.Split(pathValue, separator)
	kept := make([]string, 0, len(parts))
	removed := false
	for _, part := range parts {
		if pathEntryEqual(part, dir) {
			removed = true
			continue
		}
		kept = append(kept, part)
	}
	if !removed {
		return pathValue, false
	}
	return strings.Join(kept, separator), true
}

func pathEntryEqual(entry, dir string) bool {
	entry = strings.TrimSpace(strings.TrimRight(entry, `/\`))
	dir = strings.TrimRight(strings.TrimSpace(dir), `/\`)
	return entry != "" && dir != "" && strings.EqualFold(entry, dir)
}

// samePath compares two absolute directory paths, case-insensitively on
// Windows where paths are case-insensitive.
func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
