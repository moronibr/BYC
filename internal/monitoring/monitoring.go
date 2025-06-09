package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"byc/internal/blockchain"
	"byc/internal/logger"
	"byc/internal/network"

	"go.uber.org/zap"
)

// Monitor represents the monitoring system
// It composes Metrics, HealthCheck, and AlertSystem
type Monitor struct {
	metrics     *Metrics
	blockchain  *blockchain.Blockchain
	node        *network.Node
	healthCheck *HealthCheck
	alerts      *AlertSystem
	mu          sync.RWMutex
}

// NewMonitor creates a new monitoring system
func NewMonitor(bc *blockchain.Blockchain, node *network.Node, alertWebhook string) *Monitor {
	metrics := NewMetrics(bc, node)
	return &Monitor{
		metrics:     metrics,
		blockchain:  bc,
		node:        node,
		healthCheck: NewHealthCheck(bc, node),
		alerts:      NewAlertSystem(alertWebhook),
	}
}

// Start starts the monitoring system
func (m *Monitor) Start() error {
	// Start metrics collection
	m.metrics.Start()

	// Start health check and metrics server
	http.Handle("/metrics", m.metrics)
	http.Handle("/health", m.healthCheck)
	http.Handle("/alerts", http.HandlerFunc(m.handleAlerts))

	go func() {
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Error("Failed to start monitoring server", zap.Error(err))
		}
	}()

	return nil
}

// handleAlerts serves alerts as JSON
func (m *Monitor) handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := m.alerts.GetAlerts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// RecordError records an error in the monitoring system
func (m *Monitor) RecordError(err error) {
	m.metrics.RecordNetworkError()
	logger.Error("System error", zap.Error(err))
	m.alerts.CreateAlert(AlertLevelCritical, fmt.Sprintf("System error: %v", err), "system", nil)
}

// RecordLatency records network message latency
func (m *Monitor) RecordLatency(duration time.Duration) {
	m.metrics.RecordNetworkLatency(duration)
}

// RecordMessage records a network message
func (m *Monitor) RecordMessage() {
	// You may want to add a custom metric for messages if needed
}

// GetHealth returns the current health status
func (m *Monitor) GetHealth() map[string]interface{} {
	// Get health check status
	health := m.healthCheck.GetStatus()

	// Add monitoring status
	health["monitoring"] = map[string]interface{}{
		"uptime":     time.Since(m.metrics.startTime).String(),
		"errors":     m.metrics.errorCount,
		"peer_count": m.metrics.peerCount,
	}

	// Add resource usage
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	health["resources"] = map[string]interface{}{
		"memory_alloc": mem.Alloc,
		"memory_total": mem.TotalAlloc,
		"goroutines":   runtime.NumGoroutine(),
		"cpu_count":    runtime.NumCPU(),
		"last_update":  m.metrics.resourceStats.LastUpdate,
	}

	return health
}
