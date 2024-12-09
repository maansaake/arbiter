package cli

import (
	"os"
	"testing"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
	testmodule "tres-bon.se/arbiter/pkg/module/test"
	"tres-bon.se/arbiter/pkg/monitor"
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
	os.Args = []string{
		"subcommand",
		"-mod.count=12",
		"-mod.op.do.rate=100",
		"-mod.op.more.disable",
		"-mod.monitor.cpu.trigger=ABOVE;12",
		"-mod.monitor.cpu.trigger=ABOVE;15",
		"-mod.monitor.metric.trigger=ABOVE_OR_EQUAL;12,10;metricname",
		"-mod.monitor.vms.trigger=BELOW;12,14",
		"-mod.monitor.rss.trigger=BELOW;12,14",
		"-mod.monitor.log.trigger=EQUAL;somestring",
	}

	meta, err := Parse(0, module.Modules{mod})
	if err != nil {
		t.Fatal("should not have been an error")
	}

	m := meta[0]
	if m.MonitorOpt.MetricEndpoint != monitor.NO_METRIC_ENDPOINT {
		t.Fatal("should have been NO_METRIC_ENDPOINT")
	}
	if m.MonitorOpt.LogFile != monitor.NO_LOG_FILE {
		t.Fatal("should have been NO_LOG_FILE")
	}
	if m.MonitorOpt.PID != monitor.NO_PERFORMANCE_PID {
		t.Fatal("should have been NO_PERFORMANCE_PID")
	}
	if len(m.MonitorOpt.CPUTriggers) != 2 {
		t.Fatal("should have been 2 CPU triggers")
	}
	if len(m.MonitorOpt.VMSTriggers) != 1 {
		t.Fatal("should have been 1 VMS trigger")
	}
	if len(m.MonitorOpt.RSSTriggers) != 1 {
		t.Fatal("should have been 1 RSS trigger")
	}
	if len(m.MonitorOpt.MetricTriggers) != 1 {
		t.Fatal("should have been 1 metric trigger")
	}
	if len(m.MonitorOpt.LogTriggers) != 1 {
		t.Fatal("should have been 1 log trigger")
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
