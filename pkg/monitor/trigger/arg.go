package trigger

import (
	"tres-bon.se/arbiter/pkg/arg"
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
)

func validCPUTrigger(val string) bool {
	return false
}

func validVMSTrigger(val string) bool {
	return false
}

func validRSSTrigger(val string) bool {
	return false
}

func validLogFileTrigger(val string) bool {
	return false
}

func validMetricTrigger(val string) bool {
	return false
}
