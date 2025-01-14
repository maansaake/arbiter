package report

import (
	"context"
	"time"

	"tres-bon.se/arbiter/pkg/module"
)

type (
	Reporter interface {
		Start(context.Context)
		Op(module, op string, result *module.Result, err error)
		LogErr(module string, err error)
		LogTrigger(module string, tr *TriggerReport[string])
		CPU(module string, value float64)
		CPUErr(module string, err error)
		CPUTrigger(module string, tr *TriggerReport[float64])
		RSS(module string, value uint)
		RSSErr(module string, err error)
		RSSTrigger(module string, tr *TriggerReport[uint])
		VMS(module string, value uint)
		VMSErr(module string, err error)
		VMSTrigger(module string, tr *TriggerReport[uint])
		MetricErr(module, metric string, err error)
		MetricTrigger(module, metric string, tr *TriggerReport[float64])
		Finalise() error
	}
	TriggerReport[T ~int | ~uint | ~float64 | ~string] struct {
		Timestamp time.Time `yaml:"timestamp"`
		Type      string    `yaml:"type"`
		Value     T         `yaml:"value"`
	}
)
