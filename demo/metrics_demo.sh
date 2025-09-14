#!/bin/bash

echo "🎯 GraceWrap Prometheus Metrics Demonstration"
echo "=============================================="
echo ""

# Function to check if server is running
check_server() {
    curl -s http://localhost:8080/health/ready > /dev/null 2>&1
    return $?
}

# Function to get specific metrics
get_metric() {
    local metric_name=$1
    curl -s http://localhost:8080/metrics | grep "^$metric_name" | head -1
}

# Function to show key metrics
show_metrics() {
    local phase=$1
    echo "📊 Metrics during $phase:"
    echo "   $(get_metric gracewrap_readiness_status)"
    echo "   $(get_metric gracewrap_inflight_requests)"
    echo "   $(get_metric gracewrap_http_requests_total)"
    echo "   $(get_metric gracewrap_shutdowns_total)"
    echo ""
}

echo "Step 1: Start the graceful demo server"
echo "--------------------------------------"
echo "Run this in another terminal:"
echo "  cd demo && go run prometheus_demo.go graceful"
echo ""
echo "Press Enter when server is running..."
read

# Check if server is up
if ! check_server; then
    echo "❌ Server not running. Please start it first."
    exit 1
fi

echo "✅ Server is running!"
echo ""

echo "Step 2: Baseline metrics (before load)"
echo "-------------------------------------"
show_metrics "startup"

echo "Step 3: Generate some load"
echo "-------------------------"
echo "Generating 20 requests..."

for i in {1..20}; do
    curl -s http://localhost:8080/api/test > /dev/null &
    if [ $((i % 5)) -eq 0 ]; then
        curl -s http://localhost:8080/api/database > /dev/null &
    fi
    sleep 0.2
done

echo "Load generated. Waiting for requests to process..."
sleep 3

echo ""
echo "Step 4: Metrics during normal operation"
echo "--------------------------------------"
show_metrics "normal operation"

echo "Step 5: Start continuous load"
echo "----------------------------"
echo "Starting background load generation..."

# Start background load
(
    for i in {1..100}; do
        curl -s http://localhost:8080/api/test > /dev/null &
        curl -s http://localhost:8080/api/database > /dev/null &
        sleep 0.5
    done
) &

LOAD_PID=$!

sleep 2

echo ""
echo "Step 6: Metrics with active load"
echo "-------------------------------"
show_metrics "active load"

echo "Step 7: Trigger graceful shutdown"
echo "--------------------------------"
echo "Sending SIGTERM to trigger graceful shutdown..."
echo "Watch the metrics change as shutdown progresses:"
echo ""

# Send SIGTERM to the demo server (find the process)
SERVER_PID=$(pgrep -f "prometheus_demo.go graceful")
if [ -n "$SERVER_PID" ]; then
    echo "📡 Sending SIGTERM to server (PID: $SERVER_PID)"
    kill -TERM $SERVER_PID
    
    # Monitor metrics during shutdown
    echo ""
    echo "📊 Monitoring metrics during shutdown:"
    echo "======================================"
    
    for i in {1..10}; do
        if check_server; then
            echo "Time +${i}s:"
            show_metrics "shutdown phase $i"
            sleep 1
        else
            echo "🔚 Server stopped responding"
            break
        fi
    done
    
    echo ""
    echo "✅ Graceful shutdown demonstration completed!"
    echo ""
    echo "🎯 Key observations:"
    echo "   - gracewrap_readiness_status: 1 → 0 (ready to not ready)"
    echo "   - gracewrap_inflight_requests: X → 0 (requests drained)"
    echo "   - gracewrap_shutdowns_total: incremented"
    echo "   - gracewrap_shutdown_duration_seconds: recorded"
    
else
    echo "❌ Could not find server process"
fi

# Clean up background load
kill $LOAD_PID 2>/dev/null

echo ""
echo "🎉 Demo completed! The metrics showed graceful shutdown behavior."
