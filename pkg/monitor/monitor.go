package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/monitor/trigger"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type (
	Monitor struct {
		cpu.CPU
		memory.Memory
		metric.Metric
		log.Log
		Reporter report.Reporter

		// Disables the internal metric ticker, relying on external calls to
		// PullMetrics to handle metric fetching and triggering threshold checks.
		ExternalPrometheus bool
		// Prometheus metric endpoint address.
		MetricAddr string
	}
	// Opt contains information about a module that the monitor needs to
	// do its work.
	Opt struct {
		Name string

		// Monitoring metadata
		PID            int
		LogFile        string
		MetricEndpoint string

		// Triggers
		CPUTriggers    []trigger.Trigger[float64]
		VMSTriggers    []trigger.Trigger[uint]
		RSSTriggers    []trigger.Trigger[uint]
		LogTriggers    []trigger.Trigger[string]
		MetricTriggers map[string][]trigger.Trigger[uint]
	}
)

const (
	NO_PERFORMANCE_PID = -1
	NO_LOG_FILE        = "none"
	NO_METRIC_ENDPOINT = "none"
)

func (mi *Opt) CPUTriggerFromCmdline(cmdline string) {
	mi.CPUTriggers = append(mi.CPUTriggers, trigger.From[float64](cmdline))
}

func (mi *Opt) VMSTriggerFromCmdline(cmdline string) {
	mi.VMSTriggers = append(mi.VMSTriggers, trigger.From[uint](cmdline))
}

func (mi *Opt) RSSTriggerFromCmdline(cmdline string) {
	mi.RSSTriggers = append(mi.RSSTriggers, trigger.From[uint](cmdline))
}

func (mi *Opt) LogFileTriggerFromCmdline(cmdline string) {
	mi.LogTriggers = append(mi.LogTriggers, trigger.From[string](cmdline))
}

func (mi *Opt) MetricTriggerFromCmdline(cmdline string) {
	name, t := trigger.NamedFrom[uint](cmdline)
	if _, ok := mi.MetricTriggers[name]; !ok {
		mi.MetricTriggers[name] = make([]trigger.Trigger[uint], 0, 1)
	}
	mi.MetricTriggers[name] = append(mi.MetricTriggers[name], t)
}

var (
	logger          logr.Logger
	monitorInterval = 10 * time.Second
)

func (m *Monitor) Start(ctx context.Context, opts []*Opt) error {
	logger = zerologr.New(&zerologr.Opts{Console: true})

	if m.ExternalPrometheus {
		go func() {
			//nolint:gosec
			metricServer := &http.Server{
				Addr:    m.MetricAddr,
				Handler: http.DefaultServeMux, // Use the default handler
			}

			// Arbiter metrics
			http.Handle("/metrics", promhttp.Handler())

			// If metric endpoint(s) registered
			for _, opt := range opts {
				if opt.MetricEndpoint != NO_METRIC_ENDPOINT {
					http.HandleFunc(fmt.Sprintf("/metrics-%s", opt.Name), func(w http.ResponseWriter, r *http.Request) {
						bs, err := m.PullMetrics()
						if err != nil {
							w.WriteHeader(500)
						}
						_, err = w.Write(bs)
						if err != nil {
							w.WriteHeader(500)
						}
					})
				}
			}

			logger.Info("running metrics server on", "address", m.MetricAddr)
			if err := metricServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				logger.Error(err, "unexpected error")
			} else {
				logger.Info("metrics server shut down")
			}
		}()
	}

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
					m.handleCPUUpdate(cpu)
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
					m.handleRSSUpdate(rss)

					vms, err := m.Memory.VMS()
					if err != nil {
						logger.Error(err, "failed to fetch VMS reading")
					} else {
						logger.Info("current VMS", "bytes", vms)
					}
					m.handleVMSUpdate(vms)
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if m.Metric != nil && !m.ExternalPrometheus {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("metric monitor tick")

					rawMetrics, err := m.Metric.Pull()
					if err != nil {
						logger.Error(err, "failed to fetch metrics")
					}
					m.handleMetricUpdate(rawMetrics)
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

func (m *Monitor) handleCPUUpdate(cpu float64)     {}
func (m *Monitor) handleRSSUpdate(rss uint)        {}
func (m *Monitor) handleVMSUpdate(rss uint)        {}
func (m *Monitor) handleMetricUpdate(bytes []byte) {}

func (m *Monitor) logHandler(c string, err error) {
	if err != nil {
		logger.Error(err, "monitor log handler error")
	} else {
		logger.Info("log handler got event: %s", c)
	}
}
