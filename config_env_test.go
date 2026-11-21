package gracewrap

import (
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("DRAIN_TIMEOUT_SECONDS", "2")
	t.Setenv("HARD_STOP_TIMEOUT_SECONDS", "1")
	t.Setenv("LOAD_BALANCER_DELAY_SECONDS", "3")
	t.Setenv("ENABLE_METRICS", "true")

	cfg := ConfigFromEnv()
	if cfg.DrainTimeout != 2*time.Second {
		t.Fatalf("expected drain 2s, got %v", cfg.DrainTimeout)
	}
	if cfg.HardStopTimeout != 1*time.Second {
		t.Fatalf("expected hard stop 1s, got %v", cfg.HardStopTimeout)
	}
	if cfg.LoadBalancerDelay != 3*time.Second {
		t.Fatalf("expected load balancer delay 3s, got %v", cfg.LoadBalancerDelay)
	}
	if !cfg.EnableMetrics {
		t.Fatalf("expected metrics enabled")
	}
}
