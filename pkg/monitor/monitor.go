package monitor

import (
	"bytes"
	"context"
	"time"

	"github.com/go-logr/logr"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type Monitor struct {
	cpu.CPU
	memory.Memory
	metric.Metric
	log.Log
	Reporter report.Reporter
}

var (
	logger          logr.Logger
	monitorInterval = 10 * time.Second
)

func (m *Monitor) Start(ctx context.Context) error {
	logger = zerologr.New(&zerologr.Opts{Console: true})

	if m.CPU != nil {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("cpu monitor tick")

					cpu, err := m.CPU.Read()
					if err != nil {
						logger.Error(err, "failed to fetch CPU reading")
					} else {
						logger.Info("current CPU", "percent", cpu)
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if m.Memory != nil {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("memory monitor tick")

					rss, err := m.Memory.RSS()
					if err != nil {
						logger.Error(err, "failed to fetch RSS reading")
					} else {
						logger.Info("current RSS", "bytes", rss)
					}

					vms, err := m.Memory.VMS()
					if err != nil {
						logger.Error(err, "failed to fetch VMS reading")
					} else {
						logger.Info("current VMS", "bytes", vms)
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if m.Metric != nil {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("metric monitor tick")

					metrics, err := m.Metric.Pull()
					if err != nil {
						logger.Error(err, "failed to fetch metrics")
					}
					// TODO: remove this test print
					n := bytes.Index(metrics, []byte("\n"))
					logger.Info(string(metrics[:n]))
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if m.Log != nil {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("log monitor tick")
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	return nil
}

func (m *Monitor) LatestRawMetrics(ns string) []byte {
	return m.Metric.LatestRawMetrics()
}
