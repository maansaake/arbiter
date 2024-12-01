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

	err := Parse(0, module.Modules{mod})
	if err != nil {
		t.Fatal("should not have been an error")
	}
	if *count.Value != 12 {
		t.Fatal("should have been 12")
	}
	if *master.Value {
		t.Fatal("should have been false")
	}
	if do.Disabled {
		t.Fatal("do should be enabled")
	}
	if do.Rate != 100 {
		t.Fatal("do rate should have been 100")
	}
	if !more.Disabled {
		t.Fatal("more should be disabled")
	}
	if more.Rate != 0 {
		t.Fatal("more rate should have been 0")
	}
}
