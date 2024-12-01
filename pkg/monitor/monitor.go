package monitor

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/monitor/trigger"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type Monitor struct {
	cpu.CPU
	memory.Memory
	metric.Metric
	log.Log
	Reporter report.Reporter

	// Disables the internal metric ticker, relying on external calls to
	// PullMetrics to handle metric fetching and triggering threshold checks.
	DisableMetricTicker bool
}

// ModuleInfo contains information about a module that the monitor needs to
// do its work.
type ModuleInfo struct {
	// Monitoring metadata
	PID            int
	LogFile        string
	MetricEndpoint string

	// Triggers
	CPUTriggers    []trigger.Trigger[float64]
	VMSTrigger     []trigger.Trigger[uint]
	RSSTrigger     []trigger.Trigger[uint]
	LogTriggers    []trigger.Trigger[string]
	MetricTriggers []trigger.Trigger[int]
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

	if m.Metric != nil && !m.DisableMetricTicker {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("metric monitor tick")

					_, err := m.Metric.Pull()
					if err != nil {
						logger.Error(err, "failed to fetch metrics")
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if m.Log != nil {
		err := m.Log.Stream(ctx, m.logHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) PullMetrics() ([]byte, error) {
	if m.Metric == nil {
		panic("monitor has no metric instance set, pull is impossible")
	}
	return m.Metric.Pull()
}

func (m *Monitor) logHandler(c string, err error) {
	if err != nil {
		logger.Error(err, "monitor log handler error")
	} else {
		logger.Info("log handler got event: %s", c)
	}
}
