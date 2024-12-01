package cli

import (
	"os"
	"testing"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
	testmodule "tres-bon.se/arbiter/pkg/module/test"
)

func TestHandleModule(t *testing.T) {
	count := &arg.Arg[int]{Name: "count", Required: true, Value: new(int)}
	master := &arg.Arg[bool]{Name: "master", Value: new(bool)}

	do := &op.Op{Name: "do"}
	more := &op.Op{Name: "more"}

	mod := &testmodule.TestModule{
		SetName: "mod",
		SetArgs: arg.Args{count, master},
		SetOps:  op.Ops{do, more},
	}
	os.Args = []string{"subcommand", "-mod.count=12", "-mod.do.rate=100", "-mod.more.disable"}

	meta, err := Register(0, module.Modules{mod})
	if err != nil {
		t.Fatal("should not have been an error")
	}

	for _, m := range meta {
		if m.MetricEndpoint != NO_METRIC_ENDPOINT {
			t.Fatal("should have been NO_METRIC_ENDPOINT")
		}
		if m.LogFile != NO_LOG_FILE {
			t.Fatal("should have been NO_LOG_FILE")
		}
		if m.PID != NO_PERFORMANCE_PID {
			t.Fatal("should have been NO_PERFORMANCE_PID")
		}
	}
}
