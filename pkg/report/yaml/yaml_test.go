package yaml

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

func TestYAMLReporter(t *testing.T) {
	reportPath := "report.yaml"
	reporter := New(&Opts{Buffer: 100, Path: reportPath})
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	reporter.Start(ctx)

	yamlReporter := reporter.(*yamlReporter)
	if yamlReporter.report.Start.IsZero() {
		t.Fatal("should have been not zero")
	}
	if !yamlReporter.report.End.IsZero() {
		t.Fatal("should have been zero")
	}

	reporter.Op("mod", "op", &op.Result{Duration: 1 * time.Second}, nil)
	reporter.Op("mod", "op", &op.Result{Duration: 1 * time.Second}, errors.New("operation error"))
	reporter.Op("mod", "op2", &op.Result{Duration: 11 * time.Millisecond}, nil)
	reporter.Op("mod", "op2", &op.Result{Duration: 2 * time.Second}, nil)
	reporter.LogErr("mod", errors.New("some log error"))
	reporter.LogTrigger("mod", &report.TriggerReport[string]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     "some monitored log line",
	})
	reporter.CPU("mod", 12)
	reporter.CPU("mod", 13)
	reporter.CPUErr("mod", errors.New("some CPU error"))
	reporter.CPUTrigger("mod", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12.12,
	})
	reporter.RSS("mod", 12)
	reporter.RSS("mod", 13)
	reporter.RSSErr("mod", errors.New("some RSS error"))
	reporter.RSSTrigger("mod", &report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})
	reporter.VMS("mod", 12)
	reporter.VMS("mod", 11)
	reporter.VMSErr("mod", errors.New("some VMS error"))
	reporter.VMSTrigger("mod", &report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})
	reporter.MetricErr("mod", "metric", errors.New("some metric error"))
	reporter.MetricTrigger("mod", "metric", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})

	cancel()

	err := reporter.Finalise()
	if err != nil {
		t.Fatal("error on finalise", err)
	}

	start := yamlReporter.report.Start
	end := yamlReporter.report.End

	bs, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal("failed to read file:", reportPath)
	}
	parsedReport := &yamlReport{}
	err = yaml.Unmarshal(bs, parsedReport)
	if err != nil {
		t.Fatal("failed to unmarshal report")
	}

	if !start.Equal(parsedReport.Start) {
		t.Fatal("start should have matched", "old", start, "new", parsedReport.Start)
	}
	if !end.Equal(parsedReport.End) {
		t.Fatal("end should have matched", "old", end, "new", parsedReport.End)
	}
}
