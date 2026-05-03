package yaml

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"gopkg.in/yaml.v3"
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

	yamlReporter.Op("mod", "op", &module.Result{Duration: 1 * time.Second}, nil)
	yamlReporter.Op("mod", "op", &module.Result{Duration: 1 * time.Second}, errors.New("operation error"))
	yamlReporter.Op("mod", "op2", &module.Result{Duration: 11 * time.Millisecond}, nil)
	yamlReporter.Op("mod", "op2", &module.Result{Duration: 2 * time.Second}, nil)

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
