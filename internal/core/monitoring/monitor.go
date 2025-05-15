package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/transaction"
)

// MetricType represents the type of a metric
type MetricType string

const (
	MetricBlockHeight    MetricType = "block_height"
	MetricBlockTime      MetricType = "block_time"
	MetricTxCount        MetricType = "tx_count"
	MetricMempoolSize    MetricType = "mempool_size"
	MetricNetworkPeers   MetricType = "network_peers"
	MetricHashRate       MetricType = "hash_rate"
	MetricDifficulty     MetricType = "difficulty"
	MetricMemoryUsage    MetricType = "memory_usage"
	MetricCPUUsage       MetricType = "cpu_usage"
	MetricDiskUsage      MetricType = "disk_usage"
	MetricNetworkLatency MetricType = "network_latency"
)

// Metric represents a metric
type Metric struct {
	Type      MetricType
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// AlertLevel represents the level of an alert
type AlertLevel string

const (
	AlertInfo     AlertLevel = "info"
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

// Alert represents an alert
type Alert struct {
	Level     AlertLevel
	Message   string
	Timestamp time.Time
	Labels    map[string]string
}

// Monitor represents the monitoring system
type Monitor struct {
	metrics    map[MetricType][]*Metric
	alerts     []*Alert
	blockchain *block.Blockchain
	txPool     *transaction.TxPool
	mu         sync.RWMutex
	alertChan  chan *Alert
	metricChan chan *Metric
	stopChan   chan struct{}
}

// NewMonitor creates a new monitor
func NewMonitor(blockchain *block.Blockchain, txPool *transaction.TxPool) *Monitor {
	return &Monitor{
		metrics:    make(map[MetricType][]*Metric),
		alerts:     make([]*Alert, 0),
		blockchain: blockchain,
		txPool:     txPool,
		alertChan:  make(chan *Alert, 100),
		metricChan: make(chan *Metric, 1000),
		stopChan:   make(chan struct{}),
	}
}

// Start starts the monitor
func (m *Monitor) Start() {
	// Start metric collection
	go m.collectMetrics()

	// Start alert processing
	go m.processAlerts()

	// Start metric processing
	go m.processMetrics()
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// collectMetrics collects metrics
func (m *Monitor) collectMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			// Collect block metrics
			m.collectBlockMetrics()

			// Collect transaction metrics
			m.collectTransactionMetrics()

			// Collect network metrics
			m.collectNetworkMetrics()

			// Collect system metrics
			m.collectSystemMetrics()
		}
	}
}

// collectBlockMetrics collects block metrics
func (m *Monitor) collectBlockMetrics() {
	bestBlock := m.blockchain.GetBestBlock()
	if bestBlock == nil {
		return
	}

	// Block height
	m.metricChan <- &Metric{
		Type:      MetricBlockHeight,
		Value:     float64(bestBlock.Header.Height),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"chain_type": string(bestBlock.Header.Type),
		},
	}

	// Block time
	m.metricChan <- &Metric{
		Type:      MetricBlockTime,
		Value:     float64(time.Since(bestBlock.Header.Timestamp).Seconds()),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"chain_type": string(bestBlock.Header.Type),
		},
	}

	// Difficulty
	m.metricChan <- &Metric{
		Type:      MetricDifficulty,
		Value:     float64(bestBlock.Header.Difficulty),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"chain_type": string(bestBlock.Header.Type),
		},
	}
}

// collectTransactionMetrics collects transaction metrics
func (m *Monitor) collectTransactionMetrics() {
	// Transaction count
	m.metricChan <- &Metric{
		Type:      MetricTxCount,
		Value:     float64(m.blockchain.GetBlockCount()),
		Timestamp: time.Now(),
	}

	// Mempool size
	m.metricChan <- &Metric{
		Type:      MetricMempoolSize,
		Value:     float64(m.txPool.GetSize()),
		Timestamp: time.Now(),
	}
}

// collectNetworkMetrics collects network metrics
func (m *Monitor) collectNetworkMetrics() {
	// Network peers
	m.metricChan <- &Metric{
		Type:      MetricNetworkPeers,
		Value:     0, // TODO: Implement peer count
		Timestamp: time.Now(),
	}

	// Network latency
	m.metricChan <- &Metric{
		Type:      MetricNetworkLatency,
		Value:     0, // TODO: Implement latency measurement
		Timestamp: time.Now(),
	}
}

// collectSystemMetrics collects system metrics
func (m *Monitor) collectSystemMetrics() {
	// Memory usage
	m.metricChan <- &Metric{
		Type:      MetricMemoryUsage,
		Value:     0, // TODO: Implement memory usage measurement
		Timestamp: time.Now(),
	}

	// CPU usage
	m.metricChan <- &Metric{
		Type:      MetricCPUUsage,
		Value:     0, // TODO: Implement CPU usage measurement
		Timestamp: time.Now(),
	}

	// Disk usage
	m.metricChan <- &Metric{
		Type:      MetricDiskUsage,
		Value:     0, // TODO: Implement disk usage measurement
		Timestamp: time.Now(),
	}
}

// processMetrics processes metrics
func (m *Monitor) processMetrics() {
	for {
		select {
		case <-m.stopChan:
			return
		case metric := <-m.metricChan:
			m.mu.Lock()
			m.metrics[metric.Type] = append(m.metrics[metric.Type], metric)
			m.mu.Unlock()

			// Check for alerts
			m.checkAlerts(metric)
		}
	}
}

// processAlerts processes alerts
func (m *Monitor) processAlerts() {
	for {
		select {
		case <-m.stopChan:
			return
		case alert := <-m.alertChan:
			m.mu.Lock()
			m.alerts = append(m.alerts, alert)
			m.mu.Unlock()

			// Log alert
			fmt.Printf("Alert: %s - %s\n", alert.Level, alert.Message)
		}
	}
}

// checkAlerts checks for alerts based on metrics
func (m *Monitor) checkAlerts(metric *Metric) {
	switch metric.Type {
	case MetricBlockTime:
		if metric.Value > 30 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("Block time is high: %.2f seconds", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	case MetricMempoolSize:
		if metric.Value > 10000 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("Mempool size is high: %.0f transactions", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	case MetricNetworkPeers:
		if metric.Value < 5 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("Low number of peers: %.0f", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	case MetricMemoryUsage:
		if metric.Value > 80 {
			m.alertChan <- &Alert{
				Level:     AlertCritical,
				Message:   fmt.Sprintf("High memory usage: %.2f%%", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	case MetricCPUUsage:
		if metric.Value > 90 {
			m.alertChan <- &Alert{
				Level:     AlertCritical,
				Message:   fmt.Sprintf("High CPU usage: %.2f%%", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	case MetricDiskUsage:
		if metric.Value > 90 {
			m.alertChan <- &Alert{
				Level:     AlertCritical,
				Message:   fmt.Sprintf("High disk usage: %.2f%%", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
	}
}

// GetMetrics gets metrics of a specific type
func (m *Monitor) GetMetrics(metricType MetricType) []*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics[metricType]
}

// GetAlerts gets all alerts
func (m *Monitor) GetAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts
}

// GetAlertsByLevel gets alerts of a specific level
func (m *Monitor) GetAlertsByLevel(level AlertLevel) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range m.alerts {
		if alert.Level == level {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}
