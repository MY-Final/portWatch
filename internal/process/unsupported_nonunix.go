//go:build !windows && !linux && !darwin

package process

import (
	"context"
	"github.com/portwatch/portwatch/pkg/model"
)

type unsupportedManager struct{}

func NewManager() Manager { return unsupportedManager{} }
func (unsupportedManager) Info(context.Context, int) (model.ProcessInfo, error) {
	return model.ProcessInfo{}, ErrNotSupported
}
func (unsupportedManager) Exists(context.Context, int) (bool, error) { return false, ErrNotSupported }
func (unsupportedManager) Terminate(context.Context, int) error      { return ErrNotSupported }
