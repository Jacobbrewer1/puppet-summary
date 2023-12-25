package entities

type State string

const (
	StateChanged   State = "CHANGED"
	StateUnchanged State = "UNCHANGED"
	StateFailed    State = "FAILED"
	StateSkipped   State = "SKIPPED"
)

var States = []State{
	StateChanged,
	StateUnchanged,
	StateFailed,
	StateSkipped,
}

func (s State) String() string {
	return string(s)
}

func (s State) IsIn(states ...State) bool {
	for _, state := range states {
		if state == s {
			return true
		}
	}
	return false
}

func (s State) Valid() bool {
	return s.IsIn(
		StateChanged,
		StateUnchanged,
		StateFailed,
		StateSkipped,
	)
}

// PuppetState is used to return the number of nodes in a given state,
// and is used for the submission of metrics.
type PuppetState struct {
	State   string
	Count   int
	Percent float64
}
