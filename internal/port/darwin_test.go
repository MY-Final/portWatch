//go:build darwin

package port

import (
	"testing"
)

func TestParseLsofRowsTCP(t *testing.T) {
	output := []byte("p23184\nconode\nn[::1]:5173\np1200\ncCode Helper\nn127.0.0.1:3000->127.0.0.1:52244\n")
	rows := parseLsofRows(output, "TCP", "LISTENING")
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want 2", rows)
	}
	if rows[0].Port != 3000 || rows[0].PID != 1200 || rows[0].Protocol != "TCP" || rows[0].State != "LISTENING" {
		t.Fatalf("rows[0] = %+v", rows[0])
	}
	if rows[1].Port != 5173 || rows[1].PID != 23184 || rows[1].LocalAddr != "[::1]" || rows[1].ProcessName != "onode" {
		t.Fatalf("rows[1] = %+v", rows[1])
	}
}

func TestParseLsofRowsUDP(t *testing.T) {
	rows := parseLsofRows([]byte("p999\ncmDNSResponder\nn*:5353\n"), "UDP", "BOUND")
	if len(rows) != 1 {
		t.Fatalf("rows = %+v, want 1", rows)
	}
	row := rows[0]
	if row.Port != 5353 || row.Protocol != "UDP" || row.State != "BOUND" || row.PID != 999 || row.LocalAddr != "*" {
		t.Fatalf("row = %+v", row)
	}
}

func TestParseLsofRowsSkipsMalformedLines(t *testing.T) {
	rows := parseLsofRows([]byte("p1\ncx\nnno-colon-here\n\nnnot-a-port:n\n"), "TCP", "LISTENING")
	if len(rows) != 0 {
		t.Fatalf("rows = %+v, want none from malformed input", rows)
	}
}
