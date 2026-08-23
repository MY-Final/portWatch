//go:build !windows

package command

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// removeBinaryForOS deletes the executable directly; unix allows removing a
// running binary, so no transitional files ever exist.
func removeBinaryForOS(path string) error {
	if err := os.Remove(path); err != nil {
		return classifyRemoveError(err)
	}
	return nil
}

func isTransitionalArtifactForOS(string) bool { return false }

func classifyRemoveError(err error) error {
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("%w: %v", errUninstallBlocked, err)
	}
	return err
}

// cleanUserPathForOS does not edit shell rc files. Unix has no single
// user-level PATH store, so the uninstall prints the line to remove instead.
func cleanUserPathForOS(dir string, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "Note: no user PATH store exists on this platform; remove %s from PATH in your shell profile if it was added there.\n", dir)
	return nil
}
