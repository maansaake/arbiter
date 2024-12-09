package monitor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
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
		// Reporter to receive all reports.
		Reporter report.Reporter
		// Disables the internal metric ticker, relying on external calls to
		// PullMetrics to handle metric fetching and triggering threshold checks.
		ExternalPrometheus bool
		// Prometheus metric endpoint address.
		MetricAddr string

		// INTERNAL

		// Metric service instance (if external prometheus is true)
		metricServer *http.Server

		cpuMonitors map[string]cpu.CPU
		cpuTriggers map[string][]trigger.Trigger[float64]

		memMonitors map[string]memory.Memory
		vmsTriggers map[string][]trigger.Trigger[uint]
		rssTriggers map[string][]trigger.Trigger[uint]

		metricMonitors map[string]metric.Metric
		metricTriggers map[string]map[string][]trigger.Trigger[float64]

		logMonitors map[string]log.Log
		logTriggers map[string][]trigger.Trigger[string]
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
		MetricTriggers map[string][]trigger.Trigger[float64]
		LogTriggers    []trigger.Trigger[string]
	}
)

const (
	NO_PERFORMANCE_PID = -1
	NO_LOG_FILE        = "none"
	NO_METRIC_ENDPOINT = "none"
)

var (
	logger          logr.Logger
	monitorInterval = 10 * time.Second

	ErrMetricNotFound         = errors.New("metric not found")
	ErrMetricTypeNotSupported = errors.New("metric type not supported")
)

func DefaultOpt() *Opt {
	return &Opt{
		PID:            NO_PERFORMANCE_PID,
		MetricEndpoint: NO_METRIC_ENDPOINT,
		LogFile:        NO_LOG_FILE,
		CPUTriggers:    make([]trigger.Trigger[float64], 0),
		VMSTriggers:    make([]trigger.Trigger[uint], 0),
		RSSTriggers:    make([]trigger.Trigger[uint], 0),
		MetricTriggers: make(map[string][]trigger.Trigger[float64]),
		LogTriggers:    make([]trigger.Trigger[string], 0),
	}
}

func (o *Opt) String() string {
	return fmt.Sprintf("Opt{PID: %d, Metric: %s, Log: %s, CPUT: %d, VMST: %d, RSST: %d, MetricT: %d, LogT: %d}",
		o.PID, o.MetricEndpoint, o.LogFile, len(o.CPUTriggers), len(o.VMSTriggers), len(o.RSSTriggers), len(o.MetricTriggers), len(o.LogTriggers),
	)
}

func (o *Opt) CPUTriggerFromCmdline(cmdline string) {
	o.CPUTriggers = append(o.CPUTriggers, trigger.From[float64](cmdline))
}

func (o *Opt) VMSTriggerFromCmdline(cmdline string) {
	o.VMSTriggers = append(o.VMSTriggers, trigger.From[uint](cmdline))
}

func (o *Opt) RSSTriggerFromCmdline(cmdline string) {
	o.RSSTriggers = append(o.RSSTriggers, trigger.From[uint](cmdline))
}

func (o *Opt) LogFileTriggerFromCmdline(cmdline string) {
	o.LogTriggers = append(o.LogTriggers, trigger.From[string](cmdline))
}

func (mi *Opt) MetricTriggerFromCmdline(cmdline string) {
	name, t := trigger.NamedFrom[float64](cmdline)
	if _, ok := mi.MetricTriggers[name]; !ok {
		mi.MetricTriggers[name] = make([]trigger.Trigger[float64], 0, 1)
	}
	mi.MetricTriggers[name] = append(mi.MetricTriggers[name], t)
}

func New(opts ...*Opt) *Monitor {
	m := &Monitor{
		cpuMonitors:    make(map[string]cpu.CPU),
		cpuTriggers:    make(map[string][]trigger.Trigger[float64]),
		memMonitors:    make(map[string]memory.Memory),
		vmsTriggers:    make(map[string][]trigger.Trigger[uint]),
		rssTriggers:    make(map[string][]trigger.Trigger[uint]),
		metricMonitors: make(map[string]metric.Metric),
		metricTriggers: make(map[string]map[string][]trigger.Trigger[float64]),
		logMonitors:    make(map[string]log.Log),
		logTriggers:    make(map[string][]trigger.Trigger[string]),
	}
	for _, opt := range opts {
		m.Add(opt)
	}

	return m
}

func (m *Monitor) Add(opt *Opt) {
	if opt.PID != NO_PERFORMANCE_PID {
		//nolint:gosec
		m.cpuMonitors[opt.Name] = cpu.NewLocalCPUMonitor(int32(opt.PID))
		//nolint:gosec
		m.memMonitors[opt.Name] = memory.NewLocalMemoryMonitor(int32(opt.PID))
	}

	if opt.MetricEndpoint != NO_METRIC_ENDPOINT {
		m.metricMonitors[opt.Name] = metric.NewMetricMonitor(opt.MetricEndpoint)

		if m.ExternalPrometheus {
			http.HandleFunc(fmt.Sprintf("/metrics-%s", opt.Name), func(w http.ResponseWriter, r *http.Request) {
				bs, err := m.PullMetrics(opt.Name)
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

	if opt.LogFile != NO_LOG_FILE {
		m.logMonitors[opt.Name] = log.NewLogFileMonitor(opt.LogFile)
	}

	m.cpuTriggers[opt.Name] = make([]trigger.Trigger[float64], len(opt.CPUTriggers))
	copy(m.cpuTriggers[opt.Name], opt.CPUTriggers)

	m.vmsTriggers[opt.Name] = make([]trigger.Trigger[uint], len(opt.VMSTriggers))
	copy(m.vmsTriggers[opt.Name], opt.VMSTriggers)

	m.rssTriggers[opt.Name] = make([]trigger.Trigger[uint], len(opt.RSSTriggers))
	copy(m.rssTriggers[opt.Name], opt.RSSTriggers)

	m.logTriggers[opt.Name] = make([]trigger.Trigger[string], len(opt.LogTriggers))
	copy(m.logTriggers[opt.Name], opt.LogTriggers)

	m.metricTriggers[opt.Name] = make(map[string][]trigger.Trigger[float64], len(opt.MetricTriggers))
	for metric, triggers := range opt.MetricTriggers {
		m.metricTriggers[opt.Name][metric] = make([]trigger.Trigger[float64], len(triggers))
		copy(m.metricTriggers[opt.Name][metric], triggers)
	}
}

func (m *Monitor) Start(ctx context.Context) error {
	logger = zerologr.New(&zerologr.Opts{Console: true})

	if m.ExternalPrometheus {
		go func() {
			//nolint:gosec
			m.metricServer = &http.Server{
				Addr:    m.MetricAddr,
				Handler: http.DefaultServeMux, // Use the default handler
			}

			// Arbiter metrics
			http.Handle("/metrics", promhttp.Handler())

			logger.Info("running metrics server on", "address", m.MetricAddr)
			if err := m.metricServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				logger.Error(err, "unexpected error")
			} else {
				logger.Info("metrics server shut down")
			}
		}()
	}

	if len(m.cpuMonitors) != 0 {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("cpu monitor tick")

					for name, reader := range m.cpuMonitors {
						cpu, err := reader.Read()
						if err != nil {
							m.Reporter.CPUErr(name, err)
						} else {
							m.handleCPUUpdate(name, cpu)
						}
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if len(m.memMonitors) != 0 {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("memory monitor tick")

					for name, reader := range m.memMonitors {
						rss, err := reader.RSS()
						if err != nil {
							m.Reporter.RSSErr(name, err)
						} else {
							m.handleRSSUpdate(name, rss)
						}

						vms, err := reader.VMS()
						if err != nil {
							m.Reporter.VMSErr(name, err)
						} else {
							m.handleVMSUpdate(name, vms)
						}
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	if len(m.metricMonitors) != 0 && !m.ExternalPrometheus {
		go func() {
			tick := time.NewTicker(monitorInterval)
			for {
				select {
				case <-tick.C:
					logger.Info("metric monitor tick")

					for name, reader := range m.metricMonitors {
						rawMetrics, err := reader.Pull()
						if err != nil {
							m.Reporter.MetricErr(name, "none", err)
						} else {
							m.handleMetricUpdate(name, rawMetrics)
						}
					}
				case <-ctx.Done():
					tick.Stop()
					return
				}
			}
		}()
	}

	for name, logMonitor := range m.logMonitors {
		err := logMonitor.Stream(ctx, m.logHandler(name))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) PullMetrics(name string) ([]byte, error) {
	rawMetrics, err := m.metricMonitors[name].Pull()
	if err != nil {
		m.Reporter.MetricErr(name, "none", err)
	} else {
		m.handleMetricUpdate(name, rawMetrics)
	}
	return rawMetrics, err
}

func (m *Monitor) handleCPUUpdate(name string, cpu float64) {
	for _, trig := range m.cpuTriggers[name] {
		res := trig.Update(cpu)

		if res == trigger.RAISE || res == trigger.CLEAR {
			m.Reporter.CPUTrigger(name, res.String(), cpu)
		}
	}
}

func (m *Monitor) handleRSSUpdate(name string, rss uint) {
	for _, trig := range m.rssTriggers[name] {
		res := trig.Update(rss)

		if res == trigger.RAISE || res == trigger.CLEAR {
			m.Reporter.RSSTrigger(name, res.String(), rss)
		}
	}
}

func (m *Monitor) handleVMSUpdate(name string, vms uint) {
	for _, trig := range m.vmsTriggers[name] {
		res := trig.Update(vms)

		if res == trigger.RAISE || res == trigger.CLEAR {
			m.Reporter.RSSTrigger(name, res.String(), vms)
		}
	}
}

func (m *Monitor) handleMetricUpdate(name string, rawMetrics []byte) {
	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(bytes.NewReader(rawMetrics))
	if err != nil {
		m.Reporter.MetricErr(name, "none", err)
		return
	}

	for metricName, triggers := range m.metricTriggers[name] {
		if family, ok := families[metricName]; !ok {
			m.Reporter.MetricErr(name, metricName, ErrMetricNotFound)
		} else {
			for _, metric := range family.Metric {
				// TODO: add label support for triggers
				// TODO: add support for more metric types?
				var val float64
				switch family.GetType() {
				case io_prometheus_client.MetricType_COUNTER:
					val = metric.Counter.GetValue()
				case io_prometheus_client.MetricType_GAUGE:
					val = metric.Gauge.GetValue()
				default:
					m.Reporter.MetricErr(name, metricName, fmt.Errorf("%w: %s", ErrMetricTypeNotSupported, family.GetType()))
					return
				}

				for _, trig := range triggers {
					res := trig.Update(val)

					if res == trigger.RAISE || res == trigger.CLEAR {
						m.Reporter.MetricTrigger(name, metricName, res.String(), val)
					}
				}
			}
		}
	}
}

func (m *Monitor) logHandler(name string) log.LogHandler {
	return func(log string, err error) {
		if err != nil {
			m.Reporter.LogErr(name, err)
		} else {
			for _, trig := range m.logTriggers[name] {
				res := trig.Update(log)

				if res == trigger.RAISE || res == trigger.CLEAR {
					m.Reporter.LogTrigger(name, res.String(), log)
				}
			}
		}
	}
}
