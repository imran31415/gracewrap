package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run load_generator.go [light|heavy|database]")
		fmt.Println("  light:    Generate light load (quick requests)")
		fmt.Println("  heavy:    Generate heavy load (many concurrent requests)")
		fmt.Println("  database: Generate database simulation load (slow requests)")
		fmt.Println("")
		fmt.Println("Make sure the demo server is running first!")
		os.Exit(1)
	}

	mode := os.Args[1]

	fmt.Printf("🔄 Starting load generation: %s mode\n", mode)
	fmt.Println("📊 Watch metrics at: http://localhost:8080/metrics")
	fmt.Println("🛑 Press Ctrl+C to stop load generation")
	fmt.Println("")

	switch mode {
	case "light":
		generateLightLoad()
	case "heavy":
		generateHeavyLoad()
	case "database":
		generateDatabaseLoad()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func generateLightLoad() {
	fmt.Println("💡 Generating light load (1 request every 2 seconds)")

	client := &http.Client{Timeout: 5 * time.Second}
	requestID := 0

	for {
		requestID++
		go func(id int) {
			resp, err := client.Get("http://localhost:8080/api/test")
			if err != nil {
				fmt.Printf("❌ Request %d failed: %v\n", id, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("✅ Request %d: %s\n", id, string(body))
		}(requestID)

		time.Sleep(2 * time.Second)
	}
}

func generateHeavyLoad() {
	fmt.Println("🔥 Generating heavy load (10 concurrent requests every second)")

	client := &http.Client{Timeout: 10 * time.Second}
	requestID := 0

	for {
		// Start 10 concurrent requests
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			requestID++
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				endpoint := "/api/test"
				if id%3 == 0 {
					endpoint = "/api/database" // Mix in some slow requests
				}

				resp, err := client.Get("http://localhost:8080" + endpoint)
				if err != nil {
					fmt.Printf("❌ Request %d failed: %v\n", id, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					fmt.Printf("✅ Request %d succeeded (%s)\n", id, endpoint)
				} else {
					fmt.Printf("⚠️  Request %d status %d (%s)\n", id, resp.StatusCode, endpoint)
				}
			}(requestID)
		}

		time.Sleep(1 * time.Second)

		// Don't wait for requests to complete - keep generating load
		go func() { wg.Wait() }()
	}
}

func generateDatabaseLoad() {
	fmt.Println("💾 Generating database simulation load (slow requests)")

	client := &http.Client{Timeout: 15 * time.Second}
	requestID := 0

	for {
		requestID++
		go func(id int) {
			start := time.Now()
			resp, err := client.Get("http://localhost:8080/api/database")
			duration := time.Since(start)

			if err != nil {
				fmt.Printf("❌ DB Request %d failed after %v: %v\n", id, duration, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("✅ DB Request %d completed in %v\n", id, duration)
			} else {
				fmt.Printf("⚠️  DB Request %d status %d after %v\n", id, resp.StatusCode, duration)
			}
		}(requestID)

		time.Sleep(800 * time.Millisecond) // Slower than heavy load
	}
}
