package monitor

import (
	"context"

	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/report"
)

type Monitor struct {
	cpu.CPU
	memory.Memory
	metric.Metric
	log.Log
	Reporter report.Reporter
}

func (m *Monitor) Start(ctx context.Context) error {
	return nil
}
