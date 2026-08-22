package watch

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/portwatch/portwatch/internal/port"
	"github.com/portwatch/portwatch/pkg/model"
)

type EventKind string

const (
	Added   EventKind = "added"
	Removed EventKind = "removed"
)

type Event struct {
	Kind       EventKind
	Port       model.PortInfo
	ObservedAt time.Time
}

type Engine struct {
	Scanner  port.Scanner
	Interval time.Duration
	Now      func() time.Time
}

func (e Engine) Run(ctx context.Context, emit func(Event) error) error {
	if e.Scanner == nil || emit == nil {
		return errors.New("watch dependencies are nil")
	}
	if e.Interval <= 0 {
		return errors.New("watch interval must be positive")
	}
	now := e.Now
	if now == nil {
		now = time.Now
	}
	previous := map[portKey]model.PortInfo{}
	first := true
	for {
		currentRecords, err := e.Scanner.List(ctx)
		if err != nil {
			return err
		}
		current := make(map[portKey]model.PortInfo, len(currentRecords))
		for _, record := range currentRecords {
			current[keyOf(record)] = record
		}
		events := diff(previous, current, first, now())
		for _, event := range events {
			if err := emit(event); err != nil {
				return err
			}
		}
		previous = current
		first = false
		timer := time.NewTimer(e.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

type portKey struct {
	Protocol string
	Port     int
	PID      int
}

func keyOf(record model.PortInfo) portKey {
	return portKey{Protocol: record.Protocol, Port: record.Port, PID: record.PID}
}

func diff(previous, current map[portKey]model.PortInfo, first bool, observedAt time.Time) []Event {
	events := make([]Event, 0)
	for key, record := range current {
		if first {
			events = append(events, Event{Kind: Added, Port: record, ObservedAt: observedAt})
			continue
		}
		if _, exists := previous[key]; !exists {
			events = append(events, Event{Kind: Added, Port: record, ObservedAt: observedAt})
		}
	}
	for key, record := range previous {
		if _, exists := current[key]; !exists {
			events = append(events, Event{Kind: Removed, Port: record, ObservedAt: observedAt})
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Port.Port != events[j].Port.Port {
			return events[i].Port.Port < events[j].Port.Port
		}
		if events[i].Kind != events[j].Kind {
			return events[i].Kind == Added
		}
		return events[i].Port.PID < events[j].Port.PID
	})
	return events
}
