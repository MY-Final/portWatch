//go:build windows

package port

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"unsafe"

	"github.com/portwatch/portwatch/pkg/model"
	"golang.org/x/sys/windows"
)

// ErrTCPTable identifies a failure returned by GetExtendedTcpTable.
var ErrTCPTable = errors.New("GetExtendedTcpTable failed")

const (
	tcpTableOwnerPIDAll     = 5
	udpTableOwnerPIDAll     = 1
	tcpStateListen          = 2
	errorInsufficientBuffer = 122
	afINET                  = 2
	afINET6                 = 23
	v4RowSize               = 24
	v6RowSize               = 56
	udpV4RowSize            = 12
	udpV6RowSize            = 28
)

var (
	iphlpapi            = windows.NewLazySystemDLL("iphlpapi.dll")
	getExtendedTCPTable = iphlpapi.NewProc("GetExtendedTcpTable")
	getExtendedUDPTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

// ErrUDPTable identifies a failure returned by GetExtendedUdpTable.
var ErrUDPTable = errors.New("GetExtendedUdpTable failed")

// WindowsScanner discovers TCP listeners through the Windows IP Helper API.
type WindowsScanner struct{}

// NewWindowsScanner returns a scanner for the current Windows host.
func NewWindowsScanner() *WindowsScanner { return &WindowsScanner{} }
func NewScanner() Scanner                { return NewWindowsScanner() }

var _ Scanner = (*WindowsScanner)(nil)
var _ ProtocolScanner = (*WindowsScanner)(nil)

// List returns all IPv4 and IPv6 TCP listeners.
func (s *WindowsScanner) List(ctx context.Context) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	v4, err := getTCPTable(ctx, afINET)
	if err != nil {
		return nil, err
	}
	v6, err := getTCPTable(ctx, afINET6)
	if err != nil {
		return nil, err
	}
	rows := append(v4, v6...)
	sortPortInfo(rows)
	return rows, nil
}

// Port returns all listeners on number.
func (s *WindowsScanner) Port(ctx context.Context, number int) ([]model.PortInfo, error) {
	if number < 1 || number > 65535 {
		return nil, ErrInvalidPort
	}
	rows, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.PortInfo, 0)
	for _, row := range rows {
		if row.Port == number {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func (s *WindowsScanner) ListProtocol(ctx context.Context, protocol string) ([]model.PortInfo, error) {
	switch strings.ToLower(protocol) {
	case "tcp":
		return s.List(ctx)
	case "udp":
		return s.listUDP(ctx)
	case "all":
		tcpRows, err := s.List(ctx)
		if err != nil {
			return nil, err
		}
		udpRows, err := s.listUDP(ctx)
		if err != nil {
			return nil, err
		}
		rows := append(tcpRows, udpRows...)
		sortPortInfo(rows)
		return rows, nil
	default:
		return nil, fmt.Errorf("unsupported protocol %q", protocol)
	}
}

func (s *WindowsScanner) PortProtocol(ctx context.Context, number int, protocol string) ([]model.PortInfo, error) {
	if number < 1 || number > 65535 {
		return nil, ErrInvalidPort
	}
	rows, err := s.ListProtocol(ctx, protocol)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.PortInfo, 0)
	for _, row := range rows {
		if row.Port == number {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func (s *WindowsScanner) listUDP(ctx context.Context) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	v4, err := getUDPTable(ctx, afINET)
	if err != nil {
		return nil, err
	}
	v6, err := getUDPTable(ctx, afINET6)
	if err != nil {
		return nil, err
	}
	rows := append(v4, v6...)
	sortPortInfo(rows)
	return rows, nil
}

func getTCPTable(ctx context.Context, family uint32) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var size uint32
	status, _, callErr := getExtendedTCPTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, uintptr(family), tcpTableOwnerPIDAll, 0)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if status != errorInsufficientBuffer && status != 0 {
		return nil, tcpTableError(status, callErr)
	}
	if size == 0 {
		return []model.PortInfo{}, nil
	}
	buf := make([]byte, size)
	status, _, callErr = getExtendedTCPTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 1, uintptr(family), tcpTableOwnerPIDAll, 0)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, tcpTableError(status, callErr)
	}
	if family == afINET {
		return parseTCPv4Table(buf)
	}
	return parseTCPv6Table(buf)
}

func getUDPTable(ctx context.Context, family uint32) ([]model.PortInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var size uint32
	status, _, callErr := getExtendedUDPTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, uintptr(family), udpTableOwnerPIDAll, 0)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if status != errorInsufficientBuffer && status != 0 {
		return nil, udpTableError(status, callErr)
	}
	if size == 0 {
		return []model.PortInfo{}, nil
	}
	buf := make([]byte, size)
	status, _, callErr = getExtendedUDPTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 1, uintptr(family), udpTableOwnerPIDAll, 0)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, udpTableError(status, callErr)
	}
	if family == afINET {
		return parseUDPv4Table(buf)
	}
	return parseUDPv6Table(buf)
}

func tcpTableError(status uintptr, callErr error) error {
	if callErr == nil {
		callErr = windows.Errno(status)
	}
	return fmt.Errorf("%w: %v", ErrTCPTable, callErr)
}

func udpTableError(status uintptr, callErr error) error {
	if callErr == nil {
		callErr = windows.Errno(status)
	}
	return fmt.Errorf("%w: %v", ErrUDPTable, callErr)
}

func parseUDPv4Table(data []byte) ([]model.PortInfo, error) {
	return parseUDPTable(data, udpV4RowSize, func(row []byte) (model.PortInfo, bool) {
		return model.PortInfo{
			Port:      int(binary.BigEndian.Uint16(row[4:6])),
			Protocol:  "UDP",
			LocalAddr: net.IP(row[0:4]).String(),
			State:     "BOUND",
			PID:       int(binary.LittleEndian.Uint32(row[8:12])),
		}, true
	})
}

func parseUDPv6Table(data []byte) ([]model.PortInfo, error) {
	return parseUDPTable(data, udpV6RowSize, func(row []byte) (model.PortInfo, bool) {
		return model.PortInfo{
			Port:      int(binary.BigEndian.Uint16(row[20:22])),
			Protocol:  "UDP",
			LocalAddr: net.IP(row[0:16]).String(),
			State:     "BOUND",
			PID:       int(binary.LittleEndian.Uint32(row[24:28])),
		}, true
	})
}

func parseUDPTable(data []byte, rowSize int, parse func([]byte) (model.PortInfo, bool)) ([]model.PortInfo, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: table header is truncated", ErrUDPTable)
	}
	count := int(binary.LittleEndian.Uint32(data[:4]))
	if count > (len(data)-4)/rowSize {
		return nil, fmt.Errorf("%w: table rows are truncated", ErrUDPTable)
	}
	rows := make([]model.PortInfo, 0, count)
	for i := 0; i < count; i++ {
		row, ok := parse(data[4+i*rowSize : 4+(i+1)*rowSize])
		if ok {
			rows = append(rows, row)
		}
	}
	sortPortInfo(rows)
	return rows, nil
}

func parseTCPv4Table(data []byte) ([]model.PortInfo, error) {
	return parseTCPTable(data, v4RowSize, func(row []byte) (model.PortInfo, bool) {
		if binary.LittleEndian.Uint32(row[0:4]) != tcpStateListen {
			return model.PortInfo{}, false
		}
		return model.PortInfo{
			Port:       int(binary.BigEndian.Uint16(row[8:10])),
			Protocol:   "TCP",
			LocalAddr:  net.IP(row[4:8]).String(),
			RemoteAddr: net.IP(row[12:16]).String(),
			State:      "LISTENING",
			PID:        int(binary.LittleEndian.Uint32(row[20:24])),
		}, true
	})
}

func parseTCPv6Table(data []byte) ([]model.PortInfo, error) {
	return parseTCPTable(data, v6RowSize, func(row []byte) (model.PortInfo, bool) {
		if binary.LittleEndian.Uint32(row[48:52]) != tcpStateListen {
			return model.PortInfo{}, false
		}
		return model.PortInfo{
			Port:       int(binary.BigEndian.Uint16(row[20:22])),
			Protocol:   "TCP",
			LocalAddr:  net.IP(row[0:16]).String(),
			RemoteAddr: net.IP(row[24:40]).String(),
			State:      "LISTENING",
			PID:        int(binary.LittleEndian.Uint32(row[52:56])),
		}, true
	})
}

func parseTCPTable(data []byte, rowSize int, parse func([]byte) (model.PortInfo, bool)) ([]model.PortInfo, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: table header is truncated", ErrTCPTable)
	}
	count := int(binary.LittleEndian.Uint32(data[:4]))
	if count > (len(data)-4)/rowSize {
		return nil, fmt.Errorf("%w: table rows are truncated", ErrTCPTable)
	}
	rows := make([]model.PortInfo, 0, count)
	for i := 0; i < count; i++ {
		row, ok := parse(data[4+i*rowSize : 4+(i+1)*rowSize])
		if ok {
			rows = append(rows, row)
		}
	}
	sortPortInfo(rows)
	return rows, nil
}

func sortPortInfo(rows []model.PortInfo) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Port != rows[j].Port {
			return rows[i].Port < rows[j].Port
		}
		if rows[i].LocalAddr != rows[j].LocalAddr {
			return rows[i].LocalAddr < rows[j].LocalAddr
		}
		if rows[i].RemoteAddr != rows[j].RemoteAddr {
			return rows[i].RemoteAddr < rows[j].RemoteAddr
		}
		return rows[i].PID < rows[j].PID
	})
}
