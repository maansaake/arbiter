package trigger

import (
	"testing"

	"github.com/go-logr/logr"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/zerologr"
)

func TestValidateTriggerArg(t *testing.T) {
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{Console: true, V: 100, Caller: true}))
	defer func() {
		zerologr.SetLogger(logr.Logger{})
	}()

	tests := []struct {
		Name           string
		Cmdline        string
		ValidationFunc module.Validator[string]
		Valid          bool
	}{
		{
			Name:           "valid metric",
			Cmdline:        "ABOVE;12,10;metricname",
			ValidationFunc: validMetricTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid metric, signed clear",
			Cmdline:        "ABOVE;12,-10;metricname",
			ValidationFunc: validMetricTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid metric, signed raise",
			Cmdline:        "ABOVE;-12,10;metricname",
			ValidationFunc: validMetricTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid metric, no name",
			Cmdline:        "ABOVE;12,10",
			ValidationFunc: validMetricTrigger,
			Valid:          false,
		},
		{
			Name:           "valid CPU",
			Cmdline:        "ABOVE;12,10",
			ValidationFunc: validCPUTrigger,
			Valid:          true,
		},
		{
			Name:           "valid CPU, no clear",
			Cmdline:        "ABOVE;12",
			ValidationFunc: validCPUTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid CPU, clear above raise",
			Cmdline:        "ABOVE;12,13",
			ValidationFunc: validCPUTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid CPU, raise below clear",
			Cmdline:        "BELOW;13,12",
			ValidationFunc: validCPUTrigger,
			Valid:          false,
		},
		{
			Name:           "valid log",
			Cmdline:        "EQUAL;13,12",
			ValidationFunc: validLogFileTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid log, not EQUAL",
			Cmdline:        "BELOW;13,12",
			ValidationFunc: validLogFileTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid log, no raise value",
			Cmdline:        "BELOW",
			ValidationFunc: validLogFileTrigger,
			Valid:          false,
		},
		{
			Name:           "valid VMS",
			Cmdline:        "ABOVE;1100",
			ValidationFunc: validVMSTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid VMS, clear higher",
			Cmdline:        "ABOVE;1100,1200",
			ValidationFunc: validVMSTrigger,
			Valid:          false,
		},
		{
			Name:           "valid RSS",
			Cmdline:        "ABOVE;1100",
			ValidationFunc: validRSSTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid RSS, has name",
			Cmdline:        "ABOVE;1100;name",
			ValidationFunc: validRSSTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid RSS, too many values",
			Cmdline:        "ABOVE;1100,12,12",
			ValidationFunc: validRSSTrigger,
			Valid:          false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.ValidationFunc(test.Cmdline) != test.Valid {
				t.Fatalf("test '%s' expected to be valid: %t", test.Name, test.Valid)
			}
		})
	}
}
