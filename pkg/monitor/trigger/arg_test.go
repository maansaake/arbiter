package trigger

import (
	"testing"

	"github.com/go-logr/logr"
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/zerologr"
)

func TestValidateTriggerArg(t *testing.T) {
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{Console: true, V: 100}))
	defer func() {
		zerologr.SetLogger(logr.Logger{})
	}()

	tests := []struct {
		Name           string
		Cmdline        string
		ValidationFunc arg.Validator[string]
		Valid          bool
	}{
		{
			Name:           "valid metric",
			Cmdline:        "ABOVE;12,10;metricname",
			ValidationFunc: validMetricTrigger,
			Valid:          true,
		},
		{
			Name:           "invalid metric, signed",
			Cmdline:        "ABOVE;12,-10;metricname",
			ValidationFunc: validMetricTrigger,
			Valid:          false,
		},
		{
			Name:           "invalid metric, no name",
			Cmdline:        "ABOVE;12,10",
			ValidationFunc: validMetricTrigger,
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
