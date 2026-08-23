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

// removeBinaryForOS cannot delete the running executable directly, so it
// renames it out of the way first (renames of running binaries are allowed)
// and lets a detached cmd delete the renamed file shortly after this
// process exits.
func removeBinaryForOS(path string) error {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".exe") {
		name = strings.TrimSuffix(name, ext)
	}
	staged := filepath.Join(dir, name+".uninstalling.exe")
	if err := os.Rename(path, staged); err != nil {
		return classifyRemoveError(err)
	}
	delCommand := fmt.Sprintf("timeout /t 1 /nobreak >nul & del /q %s", quoteWindowsPath(staged))
	cmd := exec.Command("cmd", "/c", delCommand)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("schedule deletion of %s: %w", staged, err)
	}
	return nil
}

func quoteWindowsPath(path string) string {
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
