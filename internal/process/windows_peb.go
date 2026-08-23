//go:build windows

package process

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// queryProcessParameters reads the command line and working directory from
// the target's PEB. Each call is a few ReadProcessMemory syscalls, so a full
// port listing no longer spawns one PowerShell per PID.
func queryProcessParameters(handle windows.Handle, pid int) (remoteProcessParameters, error) {
	var info processBasicInformation
	if err := windows.NtQueryInformationProcess(handle, processBasicInformationClass, unsafe.Pointer(&info), uint32(unsafe.Sizeof(info)), nil); err != nil {
		return remoteProcessParameters{}, mapNTStatus("query process basic information", pid, err)
	}
	if info.PebBaseAddress == 0 {
		return remoteProcessParameters{}, fmt.Errorf("read process parameters for pid %d: %w", pid, ErrProcessNotFound)
	}

	peb := make([]byte, offsetPebProcessParameters+sizePointer)
	if err := readProcessMemory(handle, info.PebBaseAddress, peb); err != nil {
		return remoteProcessParameters{}, fmt.Errorf("read PEB for pid %d: %w", pid, err)
	}
	parametersAddress := readPointer(peb[offsetPebProcessParameters:])
	if parametersAddress == 0 {
		return remoteProcessParameters{}, fmt.Errorf("read process parameters for pid %d: process has no parameters", pid)
	}

	block := make([]byte, offsetParamsCommandLine+sizeUnicodeString)
	if err := readProcessMemory(handle, parametersAddress, block); err != nil {
		return remoteProcessParameters{}, fmt.Errorf("read process parameters for pid %d: %w", pid, err)
	}

	parameters := remoteProcessParameters{}
	var err error
	if parameters.CurrentDirectory, err = readRemoteUnicodeString(handle, block, offsetParamsCurrentDir); err != nil {
		return remoteProcessParameters{}, fmt.Errorf("read current directory for pid %d: %w", pid, err)
	}
	if parameters.CommandLine, err = readRemoteUnicodeString(handle, block, offsetParamsCommandLine); err != nil {
		return remoteProcessParameters{}, fmt.Errorf("read command line for pid %d: %w", pid, err)
	}
	return parameters, nil
}

type remoteProcessParameters struct {
	CurrentDirectory string
	CommandLine      string
}

// processBasicInformation mirrors PROCESS_BASIC_INFORMATION. ExitStatus,
// AffinityMask and BasePriority are kept pointer-sized in this Go mirror so
// PebBaseAddress keeps its native offset on both 32-bit and 64-bit builds;
// only PebBaseAddress is consumed.
type processBasicInformation struct {
	ExitStatus                   uintptr
	PebBaseAddress               uintptr
	AffinityMask                 uintptr
	BasePriority                 uintptr
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessID uintptr
}

const processBasicInformationClass = 0

// NTSTATUS values returned by NtQueryInformationProcess for handles that
// cannot be inspected.
const (
	ntStatusAccessDenied     = 0xC0000022
	ntStatusInvalidHandle    = 0xC0000008
	ntStatusInvalidParameter = 0xC000000D
)

// maxRemoteStringLength caps one UNICODE_STRING read so a stale length read
// from a dying process cannot trigger a huge allocation.
const maxRemoteStringLength = 64 * 1024

func readProcessMemory(handle windows.Handle, address uintptr, buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}
	var read uintptr
	if err := windows.ReadProcessMemory(handle, address, &buffer[0], uintptr(len(buffer)), &read); err != nil {
		return mapProcessError("read process memory", 0, err)
	}
	if read != uintptr(len(buffer)) {
		return errors.New("short read from process memory")
	}
	return nil
}

// readRemoteUnicodeString resolves one UNICODE_STRING whose fixed part was
// already copied into block at offset. Length is in bytes, not code units.
func readRemoteUnicodeString(handle windows.Handle, block []byte, offset int) (string, error) {
	length := int(binary.LittleEndian.Uint16(block[offset : offset+2]))
	buffer := readPointer(block[offset+offsetUnicodeStringBuffer:])
	if length <= 0 || buffer == 0 {
		return "", nil
	}
	if length > maxRemoteStringLength {
		length = maxRemoteStringLength
	}
	units := make([]uint16, length/2)
	var read uintptr
	if err := windows.ReadProcessMemory(handle, buffer, (*byte)(unsafe.Pointer(&units[0])), uintptr(length), &read); err != nil {
		return "", mapProcessError("read process string", 0, err)
	}
	if read != uintptr(length) {
		return "", errors.New("short read from process string")
	}
	return windows.UTF16ToString(units), nil
}

func readPointer(raw []byte) uintptr {
	if sizePointer == 8 {
		return uintptr(binary.LittleEndian.Uint64(raw))
	}
	return uintptr(binary.LittleEndian.Uint32(raw))
}

func mapNTStatus(operation string, pid int, err error) error {
	var status windows.NTStatus
	if errors.As(err, &status) {
		switch uint32(status) {
		case ntStatusAccessDenied:
			return fmt.Errorf("%s for pid %d: %w", operation, pid, ErrAccessDenied)
		case ntStatusInvalidHandle, ntStatusInvalidParameter:
			return fmt.Errorf("%s for pid %d: %w", operation, pid, ErrProcessNotFound)
		}
	}
	return fmt.Errorf("%s for pid %d: %w", operation, pid, err)
}
