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

// Uninstall confirms and deletes the running portwatch binary together with
// the sibling alias binaries installed next to it, then applies the
// conservative PATH policy: only when the binaries lived in the default
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
	aliases := siblingAliases(exePath)
	if !yes {
		_, _ = fmt.Fprintf(out, "Executable: %s\n", exePath)
		for _, alias := range aliases {
			_, _ = fmt.Fprintf(out, "Alias: %s\n", alias)
		}
		prompt := fmt.Sprintf("Uninstall %s", exePath)
		if len(aliases) > 0 {
			prompt = fmt.Sprintf("Uninstall %s and %d alias(es) in %s", filepath.Base(exePath), len(aliases), filepath.Dir(exePath))
		}
		_, _ = fmt.Fprintf(out, "%s? [y/N] ", prompt)
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
	// Installers drop portwatch and pw side by side; remove the other names
	// so one uninstall leaves nothing behind. A running alias that refuses
	// removal simply stays and is reported by the leftover-directory note.
	for _, alias := range aliases {
		if err := os.Remove(alias); err == nil {
			_, _ = fmt.Fprintf(out, "Removed %s.\n", alias)
		}
	}
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
	if err != nil {
		_, _ = fmt.Fprintf(out, "Note: could not inspect %s, so the user PATH was not modified.\n", dir)
		return nil
	}
	// The Windows delayed self-delete leaves a renamed image and a deletion
	// script behind for about a second after this process exits; they do not
	// count as leftovers, or the default-directory PATH cleanup could never
	// run. Anything else in the directory still blocks the cleanup.
	leftovers := 0
	for _, entry := range entries {
		if !entry.IsDir() && isTransitionalArtifactForOS(entry.Name()) {
			continue
		}
		leftovers++
	}
	if leftovers > 0 {
		_, _ = fmt.Fprintf(out, "Note: %s is not empty, so the user PATH was not modified.\n", dir)
		return nil
	}
	return cleanUserPathEntry(dir, out)
}

// siblingAliases lists the other known binary names (portwatch / pw) that
// sit next to the running executable, matching its extension style.
func siblingAliases(exePath string) []string {
	dir := filepath.Dir(exePath)
	self := filepath.Base(exePath)
	ext := ""
	if strings.EqualFold(filepath.Ext(self), ".exe") {
		ext = ".exe"
	}
	aliases := make([]string, 0, 1)
	for _, name := range []string{"portwatch" + ext, "pw" + ext} {
		if strings.EqualFold(name, self) {
			continue
		}
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			aliases = append(aliases, candidate)
		}
	}
	return aliases
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
