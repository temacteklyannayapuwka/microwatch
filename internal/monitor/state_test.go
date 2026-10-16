package monitor

import (
	"testing"
	"time"
)

func TestTrackerTransitions(t *testing.T) {
	start := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	var tracker Tracker
	steps := []struct {
		up      bool
		changed bool
		from    State
	}{
		{up: true},
		{up: true},
		{up: false, changed: true, from: State{Up: true, Since: start}},
		{up: false},
		{up: true, changed: true, from: State{Up: false, Since: start.Add(2 * time.Second)}},
		{up: true},
		{up: false, changed: true, from: State{Up: true, Since: start.Add(4 * time.Second)}},
	}
	for i, step := range steps {
		at := start.Add(time.Duration(i) * time.Second)
		transition, changed := tracker.Observe("catalog", step.up, at)
		if changed != step.changed {
			t.Fatalf("step %d: changed = %v, want %v", i, changed, step.changed)
		}
		if changed && (transition.From != step.from || transition.To != (State{Up: step.up, Since: at})) {
			t.Fatalf("step %d: unexpected transition %+v", i, transition)
		}
	}
}

func TestTrackerStartsDown(t *testing.T) {
	start := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	var tracker Tracker
	if _, changed := tracker.Observe("orders", false, start); changed {
		t.Fatal("first observation must not be a transition")
	}
	tracker.Observe("orders", false, start.Add(time.Second))
	transition, changed := tracker.Observe("orders", true, start.Add(5*time.Second))
	if !changed || transition.From.Up || !transition.To.Up || transition.To.Since.Sub(transition.From.Since) != 5*time.Second {
		t.Fatalf("expected recovery after 5 seconds, got %+v, changed %v", transition, changed)
	}
}

func TestTrackerSeparatesServices(t *testing.T) {
	start := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	var tracker Tracker
	tracker.Observe("catalog", true, start)
	tracker.Observe("orders", false, start.Add(time.Second))
	if _, changed := tracker.Observe("catalog", true, start.Add(2*time.Second)); changed {
		t.Fatal("another service must not affect catalog state")
	}
	transition, changed := tracker.Observe("orders", true, start.Add(3*time.Second))
	if !changed || transition.From.Since != start.Add(time.Second) || transition.From.Up {
		t.Fatalf("unexpected orders transition: %+v, changed %v", transition, changed)
	}
}
