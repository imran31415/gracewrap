package gracewrap

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Config controls graceful behavior.
type Config struct {
	// How long to wait for in-flight requests to finish after we stop accepting new ones.
	DrainTimeout time.Duration
	// Hard stop timeout after drain ends (acts as a final safety deadline).
	HardStopTimeout time.Duration
	// How long to wait for load balancers/service mesh to notice readiness change.
	// This prevents race conditions where new traffic is routed during shutdown.
	LoadBalancerDelay time.Duration
	// Optional logger (fallback to std log)
	Logger *log.Logger
	// Optional Prometheus registry for metrics
	PrometheusRegistry prometheus.Registerer
	// Optional Prometheus gatherer for metrics exposition
	PrometheusGatherer prometheus.Gatherer
	// Enable Prometheus metrics (defaults to false)
	EnableMetrics bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DrainTimeout:       25 * time.Second,
		HardStopTimeout:    5 * time.Second,
		LoadBalancerDelay:  1 * time.Second,
		EnableMetrics:      false,
		PrometheusRegistry: nil,
		PrometheusGatherer: nil,
	}
}

// ConfigFromEnv creates a Config from environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	// Parse DRAIN_TIMEOUT_SECONDS
	if val := os.Getenv("DRAIN_TIMEOUT_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			cfg.DrainTimeout = time.Duration(seconds) * time.Second
		}
	}

	// Parse HARD_STOP_TIMEOUT_SECONDS
	if val := os.Getenv("HARD_STOP_TIMEOUT_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			cfg.HardStopTimeout = time.Duration(seconds) * time.Second
		}
	}

	// Parse LOAD_BALANCER_DELAY_SECONDS
	if val := os.Getenv("LOAD_BALANCER_DELAY_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil && seconds >= 0 {
			cfg.LoadBalancerDelay = time.Duration(seconds) * time.Second
		}
	}

	// Parse ENABLE_METRICS
	if val := os.Getenv("ENABLE_METRICS"); val != "" {
		if enable, err := strconv.ParseBool(val); err == nil {
			cfg.EnableMetrics = enable
		}
	}

	return cfg
}
