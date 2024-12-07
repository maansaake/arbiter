package trigger

import (
	"errors"
	"fmt"
	"strings"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/zerologr"
)

/*
This file contains argument validation trigger parsing into trigger instances.
*/

var (
	ValidCPUTrigger     arg.Validator[string] = validCPUTrigger
	ValidVMSTrigger     arg.Validator[string] = validVMSTrigger
	ValidRSSTrigger     arg.Validator[string] = validRSSTrigger
	ValidLogFileTrigger arg.Validator[string] = validLogFileTrigger
	ValidMetricTrigger  arg.Validator[string] = validMetricTrigger

	ErrValueCount = errors.New("unexpected number of raise/clear values")
	ErrTriggerOn  = errors.New("value for TriggerOn was not recognized")
)

func validCPUTrigger(val string) bool {
	split := strings.Split(val, ";")

	// Should have 2 parts
	if len(split) != 2 {
		zerologr.V(100).Info("expected 2 parts")
		return false
	}

	if err := validate[float64](val); err != nil {
		zerologr.V(100).Info("validation failed", "error", err)
		return false
	}

	return true
}

func validVMSTrigger(val string) bool {
	split := strings.Split(val, ";")

	// Should have 3 parts
	if len(split) != 2 {
		zerologr.V(100).Info("expected 2 parts")
		return false
	}

	if err := validate[uint](val); err != nil {
		zerologr.V(100).Info("validation failed", "error", err)
		return false
	}

	return true
}

func validRSSTrigger(val string) bool {
	split := strings.Split(val, ";")

	// Should have 2 parts
	if len(split) != 2 {
		zerologr.V(100).Info("expected 2 parts")
		return false
	}

	if err := validate[uint](val); err != nil {
		zerologr.V(100).Info("validation failed", "error", err)
		return false
	}

	return true
}

func validLogFileTrigger(val string) bool {
	split := strings.Split(val, ";")

	// Should have 2 parts
	if len(split) != 2 {
		zerologr.V(100).Info("expected 2 parts")
		return false
	}

	if err := validate[string](val); err != nil {
		zerologr.V(100).Info("validation failed", "error", err)
		return false
	}

	return true
}

func validMetricTrigger(val string) bool {
	split := strings.Split(val, ";")

	// Should have 3 parts
	if len(split) != 3 {
		zerologr.V(100).Info("metric trigger had unexpected number of parts", "count", len(split))
		return false
	}

	if err := validate[uint](val); err != nil {
		zerologr.V(100).Info("validation failed", "error", err)
		return false
	}

	return true
}

func validate[T TypeConstraint](cmdline string) error {
	split := strings.Split(cmdline, ";")
	ton := parseTriggerOn(split[0])
	if ton == UNKNOWN {
		return ErrTriggerOn
	}
	values := strings.Split(split[1], ",")

	if len(values) != 1 && len(values) != 2 {
		return ErrValueCount
	}

	opts := &Opts[T]{
		TriggerOn: ton,
	}
	raise, err := parseValue[T](values[0])
	if err != nil {
		return fmt.Errorf("%w: %w", arg.ErrParse, err)
	}
	opts.Raise = raise

	if len(values) > 1 {
		opts.SendClear = true
		clear, err := parseValue[T](values[1])
		if err != nil {
			return fmt.Errorf("%w: %w", arg.ErrParse, err)
		}
		opts.Clear = clear
	}

	return Validate(opts)
}
