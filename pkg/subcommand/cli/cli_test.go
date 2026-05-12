package cli

import (
	"testing"

	"github.com/maansaake/arbiter/pkg/module"
	modulemock "github.com/maansaake/arbiter/pkg/module/mock"
	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	count := &module.Arg[int]{Name: "count", Required: true, Value: new(int)}
	master := &module.Arg[bool]{Name: "master", Value: new(bool)}

	do := &module.Op{Name: "do"}
	more := &module.Op{Name: "more"}

	mod := &modulemock.Module{
		SetName: "mod",
		SetArgs: module.Args{count, master},
		SetOps:  module.Ops{do, more},
	}

	cmd, err := NewCommand(module.Modules{mod}, func(_ module.Metadata) error { return nil })
	if err != nil {
		t.Fatal("NewCommand should not have returned an error:", err)
	}

	root := &cobra.Command{Use: "root", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{
		FlagsetName,
		"--mod.count=12",
		"--mod.op.do.rate=100",
		"--mod.op.more.disable",
	})

	if err = root.Execute(); err != nil {
		t.Fatal("should not have been an error:", err)
	}

	if !more.Disabled {
		t.Fatal("more should have been disabled")
	}

	if do.Rate != 100 {
		t.Fatal("do rate should have been 100")
	}

	if *count.Value != 12 {
		t.Fatal("module arg count should have been 12")
	}
}

