package cli

import (
	"os"
	"testing"

	"github.com/maansaake/arbiter/pkg/module"
	modulemock "github.com/maansaake/arbiter/pkg/module/mock"
)

func TestParse(t *testing.T) {
	count := &module.Arg[int]{Name: "count", Required: true, Value: new(int)}
	master := &module.Arg[bool]{Name: "master", Value: new(bool)}

	do := &module.Op{Name: "do"}
	more := &module.Op{Name: "more"}

	mod := &modulemock.Module{
		SetName: "mod",
		SetArgs: module.Args{count, master},
		SetOps:  module.Ops{do, more},
	}
	os.Args = []string{
		"subcommand",
		"-mod.count=12",
		"-mod.op.do.rate=100",
		"-mod.op.more.disable",
	}

	_, err := Parse(os.Args[1:], module.Modules{mod})
	if err != nil {
		t.Fatal("should not have been an error")
	}

	if !more.Disabled {
		t.Fatal("more should have been disabled")
	}
	if !(do.Rate == 100) {
		t.Fatal("do rate should have been 100")
	}
	if *count.Value != 12 {
		t.Fatal("module arg count should have been 12")
	}
}
