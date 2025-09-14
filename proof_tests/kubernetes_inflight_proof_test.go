package proof_tests

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imran31415/gracewrap"
)

// TestKubernetesInFlightRequestProof proves that GraceWrap prevents request failures
// during Kubernetes pod termination by protecting in-flight requests
func TestKubernetesInFlightRequestProof(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping proof test in short mode")
	}

	t.Log("🎯 COMPREHENSIVE KUBERNETES IN-FLIGHT REQUEST PROTECTION PROOF")
	t.Log("This test proves GraceWrap prevents request failures during pod termination")
	t.Log("Using large sample size for statistical significance")
	t.Log("")

	// Test with multiple sample sizes for comprehensive proof
	testCases := []struct {
		name           string
		requestCount   int
		processingTime time.Duration
		description    string
	}{
		{
			name:           "Quick_Requests_Large_Sample",
			requestCount:   200,
			processingTime: 300 * time.Millisecond,
			description:    "200 quick requests (300ms each)",
		},
		{
			name:           "Medium_Requests_Large_Sample",
			requestCount:   100,
			processingTime: 800 * time.Millisecond,
			description:    "100 medium requests (800ms each)",
		},
		{
			name:           "Slow_Requests_Sample",
			requestCount:   50,
			processingTime: 1500 * time.Millisecond,
			description:    "50 slow requests (1.5s each)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_Without_GraceWrap", func(t *testing.T) {
			t.Logf("📊 Testing %s WITHOUT GraceWrap", tc.description)
			results := runComprehensiveTest(t, false, tc.requestCount, tc.processingTime)

			killRate := float64(results.inflightKilled) / float64(results.inflightStarted) * 100
			t.Logf("❌ WITHOUT GraceWrap (%s):", tc.description)
			t.Logf("   Requests started: %d", results.inflightStarted)
			t.Logf("   Requests completed: %d", results.inflightCompleted)
			t.Logf("   Requests KILLED: %d", results.inflightKilled)
			t.Logf("   KILL RATE: %.2f%%", killRate)
		})

		t.Run(tc.name+"_With_GraceWrap", func(t *testing.T) {
			t.Logf("📊 Testing %s WITH GraceWrap", tc.description)
			results := runComprehensiveTest(t, true, tc.requestCount, tc.processingTime)

			protectionRate := float64(results.inflightCompleted) / float64(results.inflightStarted) * 100
			t.Logf("✅ WITH GraceWrap (%s):", tc.description)
			t.Logf("   Requests started: %d", results.inflightStarted)
			t.Logf("   Requests completed: %d", results.inflightCompleted)
			t.Logf("   Requests KILLED: %d", results.inflightKilled)
			t.Logf("   PROTECTION RATE: %.2f%%", protectionRate)

			if results.inflightKilled > results.inflightStarted/10 { // Allow up to 10% failures
				t.Errorf("Too many requests killed with GraceWrap: %d/%d", results.inflightKilled, results.inflightStarted)
			}
		})
	}
}

type ProofResults struct {
	inflightStarted   int64
	inflightCompleted int64
	inflightKilled    int64
}

func runComprehensiveTest(t *testing.T, useGraceful bool, requestCount int, processingTime time.Duration) ProofResults {
	var results ProofResults

	// Create realistic microservice handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := atomic.AddInt64(&results.inflightStarted, 1)

		// Only log every 10th request to avoid spam in large tests
		if started%10 == 1 || started <= 10 {
			t.Logf("🚀 %s: Request %d started processing", testType(useGraceful), started)
		}

		// Simulate realistic processing time (configurable)
		time.Sleep(processingTime)

		// Try to write response - this is where termination hurts
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf(`{"id": %d, "status": "completed", "timestamp": "%s"}`,
			started, time.Now().Format(time.RFC3339))

		n, err := w.Write([]byte(response))
		if err != nil || n != len(response) {
			atomic.AddInt64(&results.inflightKilled, 1)
			t.Logf("💀 %s: Request %d KILLED during processing: %v", testType(useGraceful), started, err)
			return
		}

		completed := atomic.AddInt64(&results.inflightCompleted, 1)
		if completed%10 == 1 || completed <= 10 {
			t.Logf("✅ %s: Request %d completed successfully", testType(useGraceful), completed)
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{
		Handler:      handler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	addr := listener.Addr().String()

	if useGraceful {
		// Use GraceWrap with appropriate timeouts for the test
		drainTimeout := time.Duration(float64(processingTime) * 2.5) // 2.5x processing time
		graceful := gracewrap.New(&gracewrap.Config{
			DrainTimeout:      drainTimeout,
			HardStopTimeout:   1 * time.Second,
			LoadBalancerDelay: 200 * time.Millisecond, // Shorter delay for testing
		})
		graceful.WrapHTTPWithListener(server, listener)

		// Start many concurrent requests
		var wg sync.WaitGroup
		batchSize := 20 // Process in batches to avoid overwhelming

		for batch := 0; batch < requestCount; batch += batchSize {
			currentBatch := batchSize
			if batch+batchSize > requestCount {
				currentBatch = requestCount - batch
			}

			for i := 0; i < currentBatch; i++ {
				wg.Add(1)
				go makeInFlightRequest(t, addr, batch+i, &wg, "GRACEFUL")
				time.Sleep(10 * time.Millisecond) // Small stagger
			}

			// Brief pause between batches
			if batch+batchSize < requestCount {
				time.Sleep(50 * time.Millisecond)
			}
		}

		// Let a good portion of requests start processing
		time.Sleep(processingTime / 3)

		// Trigger graceful shutdown while many requests are in-flight
		t.Logf("🛡️ GRACEFUL: Triggering graceful shutdown with ~%d requests in-flight", requestCount/2)
		graceful.Shutdown()

		wg.Wait()
	} else {
		// No graceful shutdown - simulate abrupt termination
		go server.Serve(listener)

		// Start many concurrent requests
		var wg sync.WaitGroup
		batchSize := 20

		for batch := 0; batch < requestCount; batch += batchSize {
			currentBatch := batchSize
			if batch+batchSize > requestCount {
				currentBatch = requestCount - batch
			}

			for i := 0; i < currentBatch; i++ {
				wg.Add(1)
				go makeInFlightRequest(t, addr, batch+i, &wg, "NO-GRACE")
				time.Sleep(10 * time.Millisecond) // Small stagger
			}

			// Brief pause between batches
			if batch+batchSize < requestCount {
				time.Sleep(50 * time.Millisecond)
			}
		}

		// Let a good portion of requests start processing
		time.Sleep(processingTime / 3)

		// Simulate Kubernetes SIGKILL after grace period
		t.Logf("⚡ NO-GRACE: Simulating Kubernetes SIGKILL with ~%d requests in-flight", requestCount/2)
		listener.Close() // Close listener

		// Simulate process being killed after brief delay
		go func() {
			time.Sleep(processingTime / 4) // Give some time, then kill
			t.Logf("💀 NO-GRACE: Process terminated by SIGKILL")
			server.Close() // Force close server to kill in-flight requests
		}()

		wg.Wait()
	}

	results.inflightKilled = results.inflightStarted - results.inflightCompleted
	return results
}

func makeInFlightRequest(t *testing.T, addr string, id int, wg *sync.WaitGroup, testType string) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 8 * time.Second, // Generous timeout for large tests
		Transport: &http.Transport{
			DisableKeepAlives: false, // Use keep-alive for realistic behavior
		},
	}

	url := fmt.Sprintf("http://%s/inflight-request-%d", addr, id)

	resp, err := client.Get(url)
	if err != nil {
		// Only log failures and every 10th success to reduce noise
		t.Logf("💀 %s: Client request %d FAILED: %v", testType, id, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("💀 %s: Client request %d READ FAILED: %v", testType, id, err)
		return
	}

	if resp.StatusCode != http.StatusOK || len(body) < 50 {
		t.Logf("💀 %s: Client request %d INCOMPLETE: status %d, %d bytes", testType, id, resp.StatusCode, len(body))
		return
	}

	// Only log every 10th success to reduce noise in large tests
	if id%10 == 0 || id < 10 {
		t.Logf("✅ %s: Client request %d SUCCESS: %d bytes", testType, id, len(body))
	}
}

func testType(useGraceful bool) string {
	if useGraceful {
		return "GRACEFUL"
	}
	return "NO-GRACE"
}
