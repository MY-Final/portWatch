package command

import (
	"context"
	"strings"
	"testing"

	"github.com/portwatch/portwatch/pkg/model"
)

func TestFindFiltersAndAggregatesPorts(t *testing.T) {
	scanner := &freeScanner{initial: []model.PortInfo{
		{Port: 8080, PID: 12},
		{Port: 3000, PID: 12},
		{Port: 5432, PID: 13},
	}}
	manager := &freeManager{infos: map[int]model.ProcessInfo{
		12: {PID: 12, Name: "Node.exe"},
		13: {PID: 13, Name: "postgres.exe"},
	}}
	var out strings.Builder
	if err := Find(context.Background(), scanner, manager, "node", &out); err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "12") || !strings.Contains(text, "3000,8080") || strings.Contains(text, "postgres") {
		t.Fatalf("Find() output = %q", text)
	}
}
