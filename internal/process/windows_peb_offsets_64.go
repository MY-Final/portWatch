//go:build windows && (amd64 || arm64)

package process

// Offsets into the native-width PEB and RTL_USER_PROCESS_PARAMETERS. These
// are stable ABI values (see Windows Internals, "Process Parameters"); a
// 64-bit PortWatch reads both 64-bit and WOW64 targets through the 64-bit
// PEB, whose parameters mirror the 32-bit ones.
const (
	offsetPebProcessParameters = 0x20
	offsetParamsCurrentDir     = 0x38
	offsetParamsCommandLine    = 0x70
	offsetUnicodeStringBuffer  = 0x8
	sizePointer                = 0x8
	sizeUnicodeString          = 0x10
)
