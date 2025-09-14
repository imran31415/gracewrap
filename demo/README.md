# GraceWrap Prometheus Metrics Demonstration

This demo shows GraceWrap's Prometheus metrics in action during graceful shutdown scenarios.

## 🎯 **What You'll See**

The demo demonstrates how GraceWrap metrics behave during:
1. **Normal operation** - steady metrics
2. **Load generation** - in-flight request tracking
3. **Graceful shutdown** - clean metric transitions
4. **Request protection** - zero request failures

## 🚀 **Quick Demo (Command Line)**

### Terminal 1: Start the graceful server
```bash
make demo-server-graceful
# Or manually: cd demo/server && go run prometheus_demo.go graceful
```

### Terminal 2: Generate load and observe metrics
```bash
# Light load
make demo-load-light

# Heavy load  
make demo-load-heavy

# Database simulation
cd demo/loadgen && go run load_generator.go database
```

### Terminal 3: Check metrics manually
```bash
# View all metrics
curl http://localhost:8080/metrics

# Check specific metrics
curl -s http://localhost:8080/metrics | grep gracewrap_readiness_status
curl -s http://localhost:8080/metrics | grep gracewrap_inflight_requests
```

### Trigger Shutdown
Press `Ctrl+C` in Terminal 1 and watch the metrics change during graceful shutdown.

## 📊 **Full Prometheus + Grafana Demo**

For a complete monitoring experience with dashboards:

### 1. Start Prometheus and Grafana
```bash
cd demo
docker-compose up -d
```

### 2. Start the demo server
```bash
make demo-server-graceful
```

### 3. Open Grafana Dashboard
- **Grafana**: http://localhost:3000 (admin/admin)
- **Dashboard**: "GraceWrap Graceful Shutdown Metrics"
- **Prometheus**: http://localhost:9090

### 4. Generate Load
```bash
# In another terminal
make demo-load-heavy
```

### 5. Trigger Graceful Shutdown
Press `Ctrl+C` in the server terminal and watch the dashboard show:
- **Readiness status**: 1 → 0
- **In-flight requests**: X → 0 (draining)
- **Shutdown duration**: Recorded
- **Request rate**: Clean drop to zero

## 📈 **Key Metrics to Watch**

| Metric | What to Observe |
|--------|----------------|
| `gracewrap_readiness_status` | 1 → 0 during shutdown |
| `gracewrap_inflight_requests` | Drains to 0 during shutdown |
| `gracewrap_http_requests_total` | Steady increase, then stops |
| `gracewrap_shutdown_duration_seconds` | Records shutdown timing |
| `gracewrap_shutdowns_total` | Increments on each shutdown |

## 🎯 **Expected Behavior**

### **During Normal Operation:**
- `gracewrap_readiness_status = 1`
- `gracewrap_inflight_requests` fluctuates with load
- `gracewrap_http_requests_total` steadily increases

### **During Graceful Shutdown:**
1. `gracewrap_readiness_status` drops to 0
2. `gracewrap_inflight_requests` drains to 0
3. `gracewrap_shutdowns_total` increments
4. `gracewrap_shutdown_duration_seconds` records timing
5. No abrupt metric drops (clean shutdown)

### **What You WON'T See (Thanks to GraceWrap):**
- ❌ Abrupt metric drops
- ❌ Non-zero in-flight requests at shutdown
- ❌ Request failures during shutdown

## 🔧 **Demo Commands**

```bash
# Quick metrics check
make demo-metrics

# Start monitoring stack
docker-compose up -d

# Clean up
docker-compose down
```

This demonstration provides visual proof that GraceWrap ensures clean, observable shutdown behavior with zero request loss.
