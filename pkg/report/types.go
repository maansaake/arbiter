package report

import (
	"context"
	"time"

	"tres-bon.se/arbiter/pkg/module/op"
)

type (
	Reporter interface {
		Start(context.Context)
		Op(module string, result *op.Result, err error)
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
		Timestamp time.Time
		Type      string
		Value     T
	}
)
