package processinfo

import (
	"context"
	"fmt"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

// MaxAncestorHops caps the parent chain walk for info output and the TUI.
const MaxAncestorHops = 8

// Ancestors walks the parent chain upward from self, at most max hops.
// The walk stops at the system boundary (PID <= 4), on a self reference or
// an already-seen PID (cycle guard), and on the first Info failure — for
// example a permissions error on an ancestor — leaving the chain truncated
// instead of failing. The returned slice lists the nearest parent first.
func Ancestors(ctx context.Context, manager Manager, self model.ProcessInfo, max int) []model.ProcessInfo {
	if manager == nil || max <= 0 {
		return nil
	}
	ancestors := make([]model.ProcessInfo, 0, max)
	seen := map[int]struct{}{self.PID: {}}
	current := self
	for i := 0; i < max; i++ {
		parent := current.ParentPID
		if parent <= 4 {
			break
		}
		if _, cycle := seen[parent]; cycle {
			break
		}
		info, err := manager.Info(ctx, parent)
		if err != nil {
			break
		}
		info.PID = parent
		seen[parent] = struct{}{}
		ancestors = append(ancestors, info)
		current = info
	}
	return ancestors
}

// FormatAncestors renders "self (pid) ← parent (pid) ← ..." for text
// output; it returns an empty string when there is nothing to show.
func FormatAncestors(selfName string, selfPID int, ancestors []model.ProcessAncestor) string {
	if len(ancestors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ancestors)+1)
	parts = append(parts, fmt.Sprintf("%s (%d)", chainName(selfName), selfPID))
	for _, ancestor := range ancestors {
		parts = append(parts, fmt.Sprintf("%s (%d)", chainName(ancestor.Name), ancestor.PID))
	}
	return strings.Join(parts, " ← ")
}

func chainName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "unknown"
	}
	return name
}
