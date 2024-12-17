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
	yamlReporter := New(&Opts{Buffer: 100, Path: reportPath})
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	yamlReporter.Start(ctx)

	if yamlReporter.report.Start.IsZero() {
		t.Fatal("should have been not zero")
	}
	if !yamlReporter.report.End.IsZero() {
		t.Fatal("should have been zero")
	}

	yamlReporter.Op("mod", "op", &op.Result{Duration: 1 * time.Second}, nil)
	yamlReporter.Op("mod", "op", &op.Result{Duration: 1 * time.Second}, errors.New("operation error"))
	yamlReporter.Op("mod", "op2", &op.Result{Duration: 11 * time.Millisecond}, nil)
	yamlReporter.Op("mod", "op2", &op.Result{Duration: 2 * time.Second}, nil)
	yamlReporter.LogErr("mod", errors.New("some log error"))
	yamlReporter.LogTrigger("mod", &report.TriggerReport[string]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     "some monitored log line",
	})
	yamlReporter.CPU("mod", 12)
	yamlReporter.CPU("mod", 13)
	yamlReporter.CPUErr("mod", errors.New("some CPU error"))
	yamlReporter.CPUTrigger("mod", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12.12,
	})
	yamlReporter.RSS("mod", 12)
	yamlReporter.RSS("mod", 13)
	yamlReporter.RSSErr("mod", errors.New("some RSS error"))
	yamlReporter.RSSTrigger("mod", &report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})
	yamlReporter.VMS("mod", 12)
	yamlReporter.VMS("mod", 11)
	yamlReporter.VMSErr("mod", errors.New("some VMS error"))
	yamlReporter.VMSTrigger("mod", &report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})
	yamlReporter.MetricErr("mod", "metric", errors.New("some metric error"))
	yamlReporter.MetricTrigger("mod", "metric", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     13,
	})

	cancel()

	err := yamlReporter.Finalise()
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
