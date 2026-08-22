package watch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/portwatch/portwatch/pkg/model"
)

type sequenceScanner struct {
	sequences [][]model.PortInfo
	index     int
}

func (s *sequenceScanner) List(context.Context) ([]model.PortInfo, error) {
	if s.index >= len(s.sequences) {
		return s.sequences[len(s.sequences)-1], nil
	}
	value := s.sequences[s.index]
	s.index++
	return value, nil
}
func (s *sequenceScanner) Port(context.Context, int) ([]model.PortInfo, error) { return nil, nil }

func TestEngineEmitsInitialAddThenChanges(t *testing.T) {
	first := model.PortInfo{Port: 8080, Protocol: "TCP", PID: 1}
	second := model.PortInfo{Port: 3000, Protocol: "TCP", PID: 2}
	scanner := &sequenceScanner{sequences: [][]model.PortInfo{{first}, {second}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var events []Event
	err := (Engine{Scanner: scanner, Interval: time.Millisecond}).Run(ctx, func(event Event) error {
		events = append(events, event)
		if len(events) == 3 {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if len(events) != 3 || events[0].Kind != Added || events[1].Kind != Added || events[2].Kind != Removed {
		t.Fatalf("events = %+v", events)
	}
}

func TestDiffInitialEventsAreSorted(t *testing.T) {
	events := diff(nil, map[portKey]model.PortInfo{
		{Port: 9000, PID: 2}: {Port: 9000, PID: 2},
		{Port: 3000, PID: 1}: {Port: 3000, PID: 1},
	}, true, time.Unix(0, 0))
	if len(events) != 2 || events[0].Port.Port != 3000 || events[1].Port.Port != 9000 {
		t.Fatalf("events = %+v", events)
	}
}
