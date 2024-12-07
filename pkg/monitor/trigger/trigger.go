// Triggers specify raise and clear levels for monitored aspects of an
// application. This could be metrics, performance, or logs.
package trigger

import (
	"errors"
	"strconv"
	"strings"
)

type (
	TypeConstraint interface {
		~int | ~uint | ~float64 | ~string
	}
	// Specifies what the Trigger should trigger on. Above, below, equal, etc. a
	// certain value. Note that strings only allow equality checks.
	TriggerOn int
	Result    int
	// A stateful Trigger, when updated with a value it will return a result for
	// if it's configured boundary for "raise" or "clear" has been
	// crossed/matched. After receiving a "raise" Result, no more "raise" Results
	// are returned until the boundary was cleared and crossed once more. For
	// example, a trigger configured to yield a "raise" Result on values above
	// 9000 will yield a "raise" when updated with 9001. If the next update to
	// the trigger is with the value 10000, a new "raise" Result is not returned,
	// but a "nothing" Result is. Once the trigger has been updated with a value
	// below the raise level of 9000 and once again crosses that boundary, a new
	// "raise" Result is returned. The exception to this rule is strings, a
	// string trigger will always yield a "raise" and "clear" Result when
	// a matching string is provided.
	Trigger[T TypeConstraint] interface {
		Update(val T) Result
	}
	// Options for constructing a new Trigger.
	Opts[T TypeConstraint] struct {
		// Decides when to trigger the input raise/clear values.
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

	// Implementation of the Trigger interface.
	triggerImpl[T TypeConstraint] struct {
		triggerOn TriggerOn
		raise     T
		sendClear bool
		clear     T
		raised    bool
	}
)

const (
	// Yield raise when an input value is comparably above the raise limit.
	// Yield clear when an input value is comparably below or equal to the clear
	// limit.
	ABOVE TriggerOn = iota
	// Yield raise when an input value is comparably above or equal to the raise
	// limit.
	// Yield clear when an input value is comparably below the clear limit.
	ABOVE_OR_EQUAL
	// Yield raise when an input value is comparably below the raise limit.
	// Yield clear when an input value is comparably above or equal to the clear
	// limit.
	BELOW
	// Yield raise when an input value is comparably below or equal to the raise
	// limit.
	// Yield clear when an input value is comparably above the clear limit.
	BELOW_OR_EQUAL
	// Yield raise when an input value is comparably equal to the raise limit.
	// Yield clear when an input value is comparably equal to the clear limit.
	EQUAL

	NOTHING Result = iota
	RAISE
	CLEAR
	UNRECOGNIZED_RESULT = "unrecognized result"
)

func parseTriggerOn(str string) TriggerOn {
	upper := strings.ToUpper(str)
	switch upper {
	case "ABOVE":
		return ABOVE
	case "ABOVE_OR_EQUAL":
		return ABOVE_OR_EQUAL
	case "BELOW":
		return BELOW
	case "BELOW_OR_EQUAL":
		return BELOW_OR_EQUAL
	default:
		return EQUAL
	}
}

func (r Result) String() string {
	switch r {
	case NOTHING:
		return "nothing"
	case RAISE:
		return "raise"
	case CLEAR:
		return "clear"
	}
	return UNRECOGNIZED_RESULT
}

func parseValue[T TypeConstraint](value string) T {
	var zero T
	switch any(zero).(type) {
	case uint:
		uiv, _ := strconv.ParseUint(value, 10, 0)
		return any(uint(uiv)).(T)
	case int:
		iv, _ := strconv.ParseInt(value, 10, 0)
		return any(int(iv)).(T)
	case float64:
		fv, _ := strconv.ParseFloat(value, 64)
		return any(fv).(T)
	case string:
		return any(value).(T)
	default:
		panic("unsupported type")
	}
}

// Parses the command line input into a trigger of a generic type.
// A trigger has the form:
// triggeron;raisevalue,clearvalue
// Clear may be omitted.
// triggeron;raisevalue
// Neither this nor the named form of "NamedFrom" performs ANY error checking
// as arguments will have passed validation prior to this stage.
func From[T TypeConstraint](cmdline string) Trigger[T] {
	t := &triggerImpl[T]{}

	split := strings.Split(cmdline, ";")
	values := strings.Split(split[1], ",")
	t.triggerOn = parseTriggerOn(split[0])
	t.raise = parseValue[T](values[0])

	if len(values) > 1 {
		t.sendClear = true
		t.clear = parseValue[T](values[1])
	}

	return t
}

// Parses the command line input into a named trigger of a generic type. This
// is used only for metrics at the moment, usage may be extended in the future.
// A named trigger has the form:
// triggeron;raisevalue,clearvalue;name
// Clear may be omitted.
// triggeron;raisevalue;name
func NamedFrom[T TypeConstraint](cmdline string) (string, Trigger[T]) {
	t := &triggerImpl[T]{}

	split := strings.Split(cmdline, ";")
	values := strings.Split(split[1], ",")
	t.triggerOn = parseTriggerOn(split[0])
	t.raise = parseValue[T](values[0])

	if len(values) > 1 {
		t.sendClear = true
		t.clear = parseValue[T](values[1])
	}

	return split[2], t
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
// BELOW: Raise must be lower than or equal to clear, as values equal to raise
// won't lead to a raise Result, only values below.
// BELOW_OR_EQUAL: Raise must be lower than clear to avoid collisions where
// both raise and clear could be true, because raise is sent on values equal to
// it.
// EQUAL: String values require EQUAL to be used as there is no other meaningful
// comparison to make. On matching string values to that of raise/clear, a
// corresponding result is returned. A difference between strings and other
// value types is that a repeated call to Update with a matching string that
// should lead to a raise/clear Result will do so. For integers, for example, a
// repeated call to Update with the same number, even if equal to raise, will
// only yield ONE raise Result. Consecutive calls after a call that led to a
// raise Result will not yield a raise Result, same goes for clears. This
// behavior is in line with that of ABOVE, BELOW, etc. marking only the
// crossing and clearing of defined thresholds, not re-confirming that the
// trigger is still residing out-of-bounds.
func Validate[T TypeConstraint](opts *Opts[T]) error {
	if opts.SendClear {
		switch any(opts.Raise).(type) {
		case string:
			if opts.TriggerOn != EQUAL {
				return errors.New("string triggers must use EQUAL")
			}
		default:
			switch opts.TriggerOn {
			case ABOVE:
				if !(opts.Raise >= opts.Clear) {
					return errors.New("For ABOVE, raise can't be lower than clear")
				}
			case ABOVE_OR_EQUAL:
				if !(opts.Raise > opts.Clear) {
					return errors.New("For ABOVE_OR_EQUAL, raise can't be lower or equal to clear")
				}
			case BELOW:
				if !(opts.Raise <= opts.Clear) {
					return errors.New("For BELOW, raise can't be higher than clear")
				}
			case BELOW_OR_EQUAL:
				if !(opts.Raise < opts.Clear) {
					return errors.New("For BELOW_OR_EQUAL, raise can't be higher or equal to clear")
				}
			}
		}
	}

	return nil
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
