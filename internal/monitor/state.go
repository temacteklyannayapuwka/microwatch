package monitor

import "time"

type State struct {
	Up    bool
	Since time.Time
}

type Transition struct {
	From State
	To   State
}

type Tracker struct {
	states map[string]State
}

func (t *Tracker) Observe(name string, up bool, at time.Time) (Transition, bool) {
	if t.states == nil {
		t.states = make(map[string]State)
	}
	previous, exists := t.states[name]
	if exists && previous.Up == up {
		return Transition{}, false
	}
	current := State{Up: up, Since: at}
	t.states[name] = current
	if !exists {
		return Transition{}, false
	}
	return Transition{From: previous, To: current}, true
}
