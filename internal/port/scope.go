package port

import "errors"

// ErrScopeUnsupported indicates that a platform cannot provide a requested
// interactive scan scope.
var ErrScopeUnsupported = errors.New("scan scope is unsupported")
