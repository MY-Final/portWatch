//go:build !windows && !linux && !darwin

package port

import (
	"context"
	"github.com/MY-Final/portWatch/pkg/model"
)

type unsupportedScanner struct{}

func NewScanner() Scanner                                                 { return unsupportedScanner{} }
func (unsupportedScanner) List(context.Context) ([]model.PortInfo, error) { return nil, ErrUnsupported }
func (unsupportedScanner) Port(context.Context, int) ([]model.PortInfo, error) {
	return nil, ErrUnsupported
}
