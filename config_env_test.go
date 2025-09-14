package gracewrap

import (
	"os"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("DRAIN_TIMEOUT_SECONDS", "2")
	os.Setenv("HARD_STOP_TIMEOUT_SECONDS", "1")
	os.Setenv("LOAD_BALANCER_DELAY_SECONDS", "3")
	os.Setenv("ENABLE_METRICS", "true")
	t.Cleanup(func() {
		os.Unsetenv("DRAIN_TIMEOUT_SECONDS")
		os.Unsetenv("HARD_STOP_TIMEOUT_SECONDS")
		os.Unsetenv("LOAD_BALANCER_DELAY_SECONDS")
		os.Unsetenv("ENABLE_METRICS")
	})

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
