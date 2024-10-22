package monitor

import (
	"context"

	"tres-bon.se/arbiter/pkg/report"
)

type Monitor struct {
	CPU
	Memory
	Metric
	Log
	Reporter report.Reporter
}

type CPU interface {
	Read() float32
}

type Memory interface {
	Read() uint
}

type Metric interface {
	Pull()
}

type Log interface {
}

func (m *Monitor) Start(ctx context.Context) error {
	return nil
}
