//go:build windows

package port

import (
	"encoding/binary"
	"errors"
	"net"
	"testing"
)

func TestParseTCPv4TableFiltersAndSorts(t *testing.T) {
	data := make([]byte, 4+3*v4RowSize)
	binary.LittleEndian.PutUint32(data, 3)
	putV4Row(data[4:], tcpStateListen, 9000, "10.0.0.2", "0.0.0.0", 20)
	putV4Row(data[4+v4RowSize:], tcpStateListen, 8000, "127.0.0.1", "0.0.0.0", 11)
	putV4Row(data[4+2*v4RowSize:], 1, 7000, "0.0.0.0", "0.0.0.0", 12)
	rows, err := parseTCPv4Table(data)
	if err != nil || len(rows) != 2 {
		t.Fatalf("parseTCPv4Table() rows=%d error=%v, want two listeners", len(rows), err)
	}
	if rows[0].Port != 8000 || rows[1].Port != 9000 {
		t.Fatalf("ports = %d, %d, want ascending order", rows[0].Port, rows[1].Port)
	}
	if rows[0].Protocol != "TCP" || rows[0].State != "LISTENING" || rows[0].ProcessName != "" {
		t.Fatalf("metadata = %+v, want TCP LISTENING and empty process name", rows[0])
	}
}

func TestParseTCPv6Table(t *testing.T) {
	data := make([]byte, 4+v6RowSize)
	binary.LittleEndian.PutUint32(data, 1)
	row := data[4:]
	row[15], row[39] = 1, 1
	binary.BigEndian.PutUint16(row[20:22], 6553)
	binary.LittleEndian.PutUint32(row[48:52], tcpStateListen)
	binary.LittleEndian.PutUint32(row[52:56], 42)
	rows, err := parseTCPv6Table(data)
	if err != nil || len(rows) != 1 {
		t.Fatalf("parseTCPv6Table() rows=%d error=%v, want one listener", len(rows), err)
	}
	if rows[0].Port != 6553 || rows[0].LocalAddr != "::1" || rows[0].RemoteAddr != "::1" || rows[0].PID != 42 {
		t.Fatalf("row = %+v, want IPv6 listener", rows[0])
	}
}

func TestParseTCPTableRejectsTruncatedRows(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 1)
	_, err := parseTCPv4Table(data)
	if !errors.Is(err, ErrTCPTable) {
		t.Fatalf("error = %v, want errors.Is(..., ErrTCPTable)", err)
	}
}

func putV4Row(row []byte, state uint32, port uint16, local, remote string, pid uint32) {
	binary.LittleEndian.PutUint32(row[0:4], state)
	copy(row[4:8], net.ParseIP(local).To4())
	binary.BigEndian.PutUint16(row[8:10], port)
	copy(row[12:16], net.ParseIP(remote).To4())
	binary.LittleEndian.PutUint32(row[20:24], pid)
}
