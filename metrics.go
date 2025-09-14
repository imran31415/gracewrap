package gracewrap

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// metrics holds Prometheus metrics
type metrics struct {
	inflightRequests  prometheus.Gauge
	httpRequestsTotal prometheus.Counter
	grpcRequestsTotal prometheus.Counter
	shutdownDuration  prometheus.Histogram
	readinessStatus   prometheus.Gauge
	shutdownsTotal    prometheus.Counter
	registerer        prometheus.Registerer
	gatherer          prometheus.Gatherer
}

// newMetrics creates and registers Prometheus metrics
func newMetrics(registry prometheus.Registerer) *metrics {
	// If no registry provided, create a fresh one so we don't depend on globals
	var reg prometheus.Registerer
	var gath prometheus.Gatherer
	if registry == nil {
		r := prometheus.NewRegistry()
		reg = r
		gath = r
	} else {
		reg = registry
		// Best effort: if registry is also a Gatherer, use it; otherwise fall back to DefaultGatherer
		if gr, ok := registry.(prometheus.Gatherer); ok {
			gath = gr
		} else {
			gath = prometheus.DefaultGatherer
		}
	}

	m := &metrics{
		inflightRequests: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gracewrap_inflight_requests",
			Help: "Current number of in-flight requests",
		}),
		httpRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gracewrap_http_requests_total",
			Help: "Total number of HTTP requests processed",
		}),
		grpcRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gracewrap_grpc_requests_total",
			Help: "Total number of gRPC requests processed",
		}),
		shutdownDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "gracewrap_shutdown_duration_seconds",
			Help:    "Time taken to complete graceful shutdown",
			Buckets: prometheus.DefBuckets,
		}),
		readinessStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gracewrap_readiness_status",
			Help: "Readiness status (1=ready, 0=not ready)",
		}),
		shutdownsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gracewrap_shutdowns_total",
			Help: "Total number of shutdowns initiated",
		}),
		registerer: reg,
		gatherer:   gath,
	}

	// Register metrics
	reg.MustRegister(
		m.inflightRequests,
		m.httpRequestsTotal,
		m.grpcRequestsTotal,
		m.shutdownDuration,
		m.readinessStatus,
		m.shutdownsTotal,
	)

	return m
}

// updateInflight updates the in-flight requests gauge
func (m *metrics) updateInflight(count int64) {
	m.inflightRequests.Set(float64(count))
}

// incHTTP increments the HTTP requests counter
func (m *metrics) incHTTP() {
	m.httpRequestsTotal.Inc()
}

// incGRPC increments the gRPC requests counter
func (m *metrics) incGRPC() {
	m.grpcRequestsTotal.Inc()
}

// updateReadiness updates the readiness status gauge
func (m *metrics) updateReadiness(ready bool) {
	if ready {
		m.readinessStatus.Set(1)
	} else {
		m.readinessStatus.Set(0)
	}
}

// incShutdowns increments the shutdowns counter
func (m *metrics) incShutdowns() {
	m.shutdownsTotal.Inc()
}

// observeShutdownDuration records the shutdown duration
func (m *metrics) observeShutdownDuration(duration time.Duration) {
	m.shutdownDuration.Observe(duration.Seconds())
}
