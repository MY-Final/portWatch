// Package processinfo resolves process metadata for port records with a
// bounded worker pool. It is shared by the CLI command handlers and the TUI
// so both paths query each process exactly once.
package processinfo

import (
	"context"
	"sync"

	"github.com/MY-Final/portWatch/pkg/model"
)

// Manager is the process capability required to resolve metadata. Any
// process manager exposing Info satisfies it structurally.
type Manager interface {
	Info(ctx context.Context, pid int) (model.ProcessInfo, error)
}

// maxWorkers bounds how many Info calls run concurrently.
const maxWorkers = 8

// Resolve looks up process metadata once per unique PID using a bounded
// worker pool (at most 8 concurrent Info calls). Cancellation stops
// dispatching new lookups; already-started calls are left to their context.
func Resolve(ctx context.Context, manager Manager, records []model.PortInfo) (map[int]model.ProcessInfo, map[int]error) {
	infos := make(map[int]model.ProcessInfo, len(records))
	errorsByPID := make(map[int]error)
	unique := make([]int, 0, len(records))
	seen := make(map[int]struct{}, len(records))
	for _, record := range records {
		if _, known := seen[record.PID]; known {
			continue
		}
		seen[record.PID] = struct{}{}
		unique = append(unique, record.PID)
	}
	if len(unique) == 0 {
		return infos, errorsByPID
	}

	workers := len(unique)
	if workers > maxWorkers {
		workers = maxWorkers
	}
	pids := make(chan int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pid := range pids {
				if ctx.Err() != nil {
					continue
				}
				info, err := manager.Info(ctx, pid)
				mu.Lock()
				if err != nil {
					errorsByPID[pid] = err
				} else {
					infos[pid] = info
				}
				mu.Unlock()
			}
		}()
	}
	for _, pid := range unique {
		select {
		case pids <- pid:
		case <-ctx.Done():
			close(pids)
			wg.Wait()
			return infos, errorsByPID
		}
	}
	close(pids)
	wg.Wait()
	return infos, errorsByPID
}

// ApplyNames copies resolved process names back onto the matching records.
func ApplyNames(records []model.PortInfo, infos map[int]model.ProcessInfo) {
	for i := range records {
		if info, ok := infos[records[i].PID]; ok {
			records[i].ProcessName = info.Name
		}
	}
}
