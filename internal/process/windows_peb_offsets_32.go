//go:build windows && 386

package process

// Offsets into the native-width PEB and RTL_USER_PROCESS_PARAMETERS for
// 32-bit builds. A 32-bit PortWatch cannot read the 64-bit PEB of native
// processes on 64-bit Windows; those calls fail with a mapping error instead
// of returning wrong data.
const (
	offsetPebProcessParameters = 0x10
	offsetParamsCurrentDir     = 0x24
	offsetParamsCommandLine    = 0x40
	offsetUnicodeStringBuffer  = 0x4
	sizePointer                = 0x4
	sizeUnicodeString          = 0x8
)
