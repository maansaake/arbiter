package report

import (
	"context"
	"time"

	"tres-bon.se/arbiter/pkg/module/op"
)

type Reporter interface {
	Start(context.Context)
	Op(module string, result *op.Result, err error)
	LogErr(module string, err error)
	LogTrigger(module string, result string, value string)
	CPU(module string, value float64)
	CPUErr(module string, err error)
	CPUTrigger(module string, result string, value float64)
	RSS(module string, value uint)
	RSSErr(module string, err error)
	RSSTrigger(module string, result string, value uint)
	VMS(module string, value uint)
	VMSErr(module string, err error)
	VMSTrigger(module string, result string, value uint)
	MetricErr(module, metric string, err error)
	MetricTrigger(module, metric string, result string, value float64)
	Finalise() error
}

type report struct {
	Start    time.Time           `yaml:"start"`
	End      time.Time           `yaml:"end"`
	Duration time.Duration       `yaml:"duration"`
	Modules  map[string]*modules `yaml:"modules"`
}

type modules struct {
	CPU        *cpu                   `yaml:"cpu"`
	Mem        *mem                   `yaml:"mem"`
	Operations map[string]*operations `yaml:"operations"`
}

type cpu struct {
	readings uint    `yaml:"-"`
	High     float64 `yaml:"high"`
	Low      float64 `yaml:"low"`
}

type mem struct {
	RSS *rss `yaml:"rss"`
	VMS *vms `yaml:"vms"`
}

type rss struct {
	readings uint    `yaml:"-"`
	High     float64 `yaml:"high"`
	Low      float64 `yaml:"low"`
}

type vms struct {
	readings uint    `yaml:"-"`
	High     float64 `yaml:"high"`
	Low      float64 `yaml:"low"`
}

type operations struct {
	Executions uint             `yaml:"executions"`
	Timing     *operationTiming `yaml:"timing"`
	Errors     []string         `yaml:"errors"`
}

type operationTiming struct {
	Longest  time.Duration `yaml:"longest"`
	Shortest time.Duration `yaml:"shortest"`
	Average  time.Duration `yaml:"average"`
	total    time.Duration `yaml:"-"`
}
