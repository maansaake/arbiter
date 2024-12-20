package cpu

import (
	"tres-bon.se/arbiter/pkg/module"
)

var ValidPID module.Validator[int] = validPID

func validPID(pid int) bool {
	//nolint:gosec
	_, err := getProc(int32(pid))
	return err == nil
}
