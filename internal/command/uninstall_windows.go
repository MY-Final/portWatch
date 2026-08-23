//go:build windows

package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Names of the short-lived files the delayed self-delete leaves next to the
// binary: the renamed image and the generated deletion script. Both are gone
// about a second after the uninstall exits; cleanInstallDirectory ignores
// them when deciding whether the install directory is empty.
const (
	stagedBinarySuffix   = ".uninstalling.exe"
	deletionScriptPrefix = "portwatch-uninstall-"
	deletionScriptSuffix = ".cmd"
)

// isTransitionalArtifactForOS reports whether a directory entry name is one
// of those transitional files.
func isTransitionalArtifactForOS(name string) bool {
	return strings.HasSuffix(name, stagedBinarySuffix) ||
		(strings.HasPrefix(name, deletionScriptPrefix) && strings.HasSuffix(name, deletionScriptSuffix))
}

// removeBinaryForOS cannot delete the running executable directly, so it
// renames it out of the way first (renames of running binaries are allowed)
// and lets a detached cmd script delete the renamed file shortly after this
// process exits. The deletion runs from a generated batch file because
// passing a compound command with quoted paths through exec.Command mangles
// the quotes (Go escapes them MSVCRT-style with backslashes, which cmd.exe
// does not understand). The script waits with ping instead of timeout.exe
// because a detached process has no console and timeout.exe fails there
// immediately, which would make del run while this process still maps the
// file.
//
// The script is created in the system temp directory rather than next to
// the binary: cmd /c mangles quoting for script paths containing & ( ) ^ %
// (the audit's adversarial-directory tests), while quoting inside batch
// content keeps those characters literal. The staged path inside the script
// still escapes % as %% because batch files expand %VAR% even in quotes.
func removeBinaryForOS(path string) error {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".exe") {
		name = strings.TrimSuffix(name, ext)
	}
	staged := filepath.Join(dir, name+stagedBinarySuffix)
	if err := os.Rename(path, staged); err != nil {
		return classifyRemoveError(err)
	}
	script, err := os.CreateTemp(os.TempDir(), deletionScriptPrefix+"*"+deletionScriptSuffix)
	if err != nil {
		return fmt.Errorf("create uninstall script for %s: %w", staged, err)
	}
	fmt.Fprintf(script, "@ping -n 2 127.0.0.1 >nul\r\n@del /q %s\r\n@del /q \"%%~f0\"\r\n", quoteWindowsPath(staged))
	closeErr := script.Close()
	if closeErr != nil {
		return fmt.Errorf("write uninstall script: %w", closeErr)
	}
	cmd := exec.Command("cmd", "/c", script.Name())
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("schedule deletion of %s: %w", staged, err)
	}
	return nil
}

func quoteWindowsPath(path string) string {
	// Batch files expand %VAR% inside double quotes; doubling renders a
	// literal percent sign. Quotes themselves are doubled per cmd rules.
	path = strings.ReplaceAll(path, "%", "%%")
	return `"` + strings.ReplaceAll(path, `"`, `""`) + `"`
}

func classifyRemoveError(err error) error {
	var errno syscall.Errno
	if errors.As(err, &errno) &&
		(errno == windows.ERROR_SHARING_VIOLATION ||
			errno == windows.ERROR_LOCK_VIOLATION ||
			errno == windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("%w: %v", errUninstallBlocked, err)
	}
	return err
}

// cleanUserPathForOS removes dir from the user-level PATH in the registry.
// The raw (unexpanded) value and its kind are preserved so existing entries
// such as %JAVA_HOME%\bin keep working. A WM_SETTINGCHANGE broadcast lets
// already-running shells notice the update without re-logging in.
func cleanUserPathForOS(dir string, out io.Writer) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open user environment: %w", err)
	}
	defer key.Close()
	value, valueType, err := key.GetStringValue("Path")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			_, _ = fmt.Fprintf(out, "User PATH did not reference %s.\n", dir)
			return nil
		}
		return fmt.Errorf("read user PATH: %w", err)
	}
	updated, removed := removePathEntry(value, dir)
	if !removed {
		_, _ = fmt.Fprintf(out, "User PATH did not reference %s.\n", dir)
		return nil
	}
	// Rewrite with the original value kind so entries like %JAVA_HOME%\bin
	// keep expanding after the update.
	var setErr error
	if valueType == registry.EXPAND_SZ {
		setErr = key.SetExpandStringValue("Path", updated)
	} else {
		setErr = key.SetStringValue("Path", updated)
	}
	if setErr != nil {
		return fmt.Errorf("update user PATH: %w", setErr)
	}
	broadcastEnvironmentChange()
	_, _ = fmt.Fprintf(out, "Removed %s from the user PATH.\n", dir)
	return nil
}

// broadcastEnvironmentChange is best-effort; failures are ignored because the
// registry value is already updated and new terminals pick it up regardless.
func broadcastEnvironmentChange() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")
	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x1a
		smtoAbortIfHung = 0x0002
	)
	var result uintptr
	environment, err := windows.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	_, _, _ = proc.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(environment)),
		uintptr(smtoAbortIfHung),
		1000,
		uintptr(unsafe.Pointer(&result)),
	)
}
