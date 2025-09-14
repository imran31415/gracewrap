package gracewrap

import (
	"context"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// shutdown performs graceful shutdown of all tracked services.
func (g *Graceful) shutdown() {
	g.stopOnce.Do(func() {
		start := time.Now()

		// Update metrics
		if g.metrics != nil {
			g.metrics.incShutdowns()
		}

		// 1. Mark as not ready to stop new traffic
		g.setReady(false)
		g.logger.Printf("Marked as not ready; health checks will now return 503")

		// 2. Wait for load balancers/service mesh to notice readiness change
		if g.config.LoadBalancerDelay > 0 {
			g.logger.Printf("Waiting %v for load balancers to stop routing traffic...", g.config.LoadBalancerDelay)
			time.Sleep(g.config.LoadBalancerDelay)
		}

		// 3. Graceful shutdown with timeout (HTTP servers will close their own listeners)
		drainDeadline := time.Now().Add(g.config.DrainTimeout)
		g.gracefulShutdown(drainDeadline)

		// 4. Wait for in-flight requests to complete
		ok := g.waitForInflight(drainDeadline)
		if !ok {
			g.logger.Printf("In-flight requests did not complete before deadline")
		}

		// 5. Final hard stop if configured
		if g.config.HardStopTimeout > 0 {
			g.logger.Printf("Waiting %v for final cleanup", g.config.HardStopTimeout)
			time.Sleep(g.config.HardStopTimeout)
		}

		// Update metrics
		if g.metrics != nil {
			g.metrics.observeShutdownDuration(time.Since(start))
		}

		g.logger.Printf("Graceful shutdown completed")
	})
}

// gracefulShutdown shuts down all servers gracefully within the deadline.
func (g *Graceful) gracefulShutdown(deadline time.Time) {
	var wg sync.WaitGroup

	// Shutdown HTTP servers
	for _, server := range g.httpServers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()
			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				g.logger.Printf("HTTP server shutdown error: %v", err)
			} else {
				g.logger.Printf("HTTP server shutdown completed")
			}
		}(server)
	}

	// Shutdown gRPC servers
	for _, server := range g.grpcServers {
		wg.Add(1)
		go func(srv *grpc.Server) {
			defer wg.Done()

			// Start graceful stop in background
			done := make(chan struct{})
			go func() {
				srv.GracefulStop()
				close(done)
			}()

			// Force stop if deadline exceeded
			timer := time.NewTimer(time.Until(deadline))
			defer timer.Stop()

			select {
			case <-done:
				g.logger.Printf("gRPC server graceful shutdown completed")
			case <-timer.C:
				g.logger.Printf("gRPC server deadline reached; forcing stop")
				srv.Stop()
			}
		}(server)
	}

	// Wait for all servers to shutdown
	wg.Wait()
}

// waitForInflight waits for all in-flight requests to complete.
func (g *Graceful) waitForInflight(deadline time.Time) bool {
	g.inflight.mu.Lock()
	defer g.inflight.mu.Unlock()

	for g.inflight.n > 0 {
		now := time.Now()
		if now.After(deadline) {
			return false
		}

		// Calculate wait time
		wait := time.Second
		if now.Add(wait).After(deadline) {
			wait = time.Until(deadline)
		}

		// Wait with timeout
		timer := time.NewTimer(wait)
		g.inflight.cv.Wait() // This will be woken up by dec() when count reaches 0
		timer.Stop()
	}

	return true
}

// setReady sets the readiness status.
func (g *Graceful) setReady(ready bool) {
	g.readyMu.Lock()
	g.ready = ready
	g.readyMu.Unlock()

	// Update metrics
	if g.metrics != nil {
		g.metrics.updateReadiness(ready)
	}
}
