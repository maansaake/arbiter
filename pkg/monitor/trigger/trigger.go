// Triggers specify raise and clear levels for monitored aspects of an
// application. This could be metrics, performance, or logs. Numerical
// triggers apply to integers (signed and unsigned) and floats.
package trigger

import (
	"cmp"
)

type TriggerOn int

const (
	ABOVE TriggerOn = iota
	ABOVE_OR_EQUAL
	BELOW_OR_EQUAL
	BELOW
	EQUAL
)

type Result int

const (
	NOTHING Result = iota
	RAISE
	CLEAR
)

func (r Result) String() string {
	switch r {
	case NOTHING:
		return "nothing"
	case RAISE:
		return "raise"
	case CLEAR:
		return "clear"
	}
	return "unrecognized result"
}

type Trigger[T cmp.Ordered] interface {
	Update(val T) Result
}

type TriggerOpts[T cmp.Ordered] struct {
	TriggerOn
	// The value to send Raise Results for, depending on the TriggerOn value.
	Raise T
	// Clear Results won't be returned unless this is set.
	SendClear bool
	// The level to send Clear Results for.
	// Clear can be used by string matchers. The updated string value matching
	// the clear field leads to a clear Result.
	Clear T
}

type triggerImpl[T cmp.Ordered] struct {
	triggerOn TriggerOn
	raise     T
	sendClear bool
	clear     T
	raised    bool
}

// Create a new Trigger with the input options. Some rules apply to raise and
// clear levels, or a panic is raised:
//
// For strings: no raise/clear levels are inspected, the strings are matched.
// If any other TriggerOn value than EQUAL is input, a panic is raised.
//
// ABOVE: Raise must be higher or equal to clear, as raise is only triggered
// by higher values than a potential matching clear.
// ABOVE_OR_EQUAL: Raise must be higher than clear, otherwise they can be
// triggered by the same value.
// BELOW: .
func New[T cmp.Ordered](opts *TriggerOpts[T]) Trigger[T] {
	if opts.SendClear {
		switch any(opts.Raise).(type) {
		case string:
			if opts.TriggerOn != EQUAL {
				panic("string triggers must use EQUAL")
			}
		default:
			switch opts.TriggerOn {
			case ABOVE:
				if !(opts.Raise >= opts.Clear) {
					panic("For ABOVE, raise can't be lower than clear")
				}
			case ABOVE_OR_EQUAL:
				if !(opts.Raise > opts.Clear) {
					panic("For ABOVE_OR_EQUAL, raise can't be lower or equal to clear")
				}
			case BELOW:
				if !(opts.Raise <= opts.Clear) {
					panic("For BELOW, raise can't be higher than clear")
				}
			case BELOW_OR_EQUAL:
				if !(opts.Raise < opts.Clear) {
					panic("For BELOW_OR_EQUAL, raise can't be higher or equal to clear")
				}
			}
		}
	}

	t := &triggerImpl[T]{
		triggerOn: opts.TriggerOn,
		// lastValue is set to zero-value of the input type parameter
		// But what if it's zero? I guess that's fine? If new value is above/below
		// or matching the raise value exactly, then raise.
		raise:     opts.Raise,
		sendClear: opts.SendClear,
		clear:     opts.Clear,
		raised:    false,
	}

	return t
}

func (t *triggerImpl[T]) Update(val T) Result {
	clearPossible := t.sendClear && t.raised
	result := NOTHING

	switch t.triggerOn {
	case ABOVE:
		// Values above 'raise' return a RAISE Result if not already raised
		// Values less than or equal to clear return a CLEAR Result if previously raised.
		if val > t.raise && !t.raised {
			result = RAISE
		} else if clearPossible && val <= t.clear {
			result = CLEAR
		}
	case ABOVE_OR_EQUAL:
		// Values above or equal to 'raise' return a RAISE Result if not already
		// raised.
		// Values below or equal to clear return a CLEAR Result if previously raised.
		if val >= t.raise && !t.raised {
			result = RAISE
		} else if clearPossible && val <= t.clear {
			result = CLEAR
		}
	case BELOW:
		// Values below 'raise' return a RAISE Result if not already raised
		// Values above or equal to clear return a CLEAR Result if previously raised.
		if val < t.raise && !t.raised {
			result = RAISE
		} else if clearPossible && val >= t.clear {
			result = CLEAR
		}
	case BELOW_OR_EQUAL:
		// Values below or equal to 'raise' return a RAISE Result if not already
		// raised
		if val <= t.raise && !t.raised {
			result = RAISE
		} else if clearPossible && val >= t.clear {
			result = CLEAR
		}
	case EQUAL:
		// Values equal to 'raise' return a RAISE Result if not already raised, or
		// if T is a string
		switch any(val).(type) {
		case string:
			if val == t.raise {
				result = RAISE
			} else if val == t.clear {
				result = CLEAR
			}
		default:
			if val == t.raise && !t.raised {
				result = RAISE
			} else if clearPossible && val == t.clear {
				result = CLEAR
			}
		}
	default:
		panic("unrecognized value for trigger")
	}
	t.raised = result == RAISE || (t.raised && result == NOTHING)

	return result
}
