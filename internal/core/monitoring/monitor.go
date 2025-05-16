package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/transaction"
)

// MetricType represents the type of a metric
type MetricType string

const (
	// Blockchain Metrics
	MetricBlockHeight  MetricType = "block_height"
	MetricBlockTime    MetricType = "block_time"
	MetricTxCount      MetricType = "tx_count"
	MetricMempoolSize  MetricType = "mempool_size"
	MetricNetworkPeers MetricType = "network_peers"
	MetricHashRate     MetricType = "hash_rate"
	MetricDifficulty   MetricType = "difficulty"

	// System Metrics
	MetricCPUUsage    MetricType = "cpu_usage"
	MetricMemoryUsage MetricType = "memory_usage"
	MetricDiskUsage   MetricType = "disk_usage"
	MetricGoroutines  MetricType = "goroutines"
	MetricThreadCount MetricType = "thread_count"

	// Network Metrics
	MetricNetworkLatency MetricType = "network_latency"
	MetricNetworkErrors  MetricType = "network_errors"
	MetricNetworkBytes   MetricType = "network_bytes"

	// Application Metrics
	MetricRequestCount   MetricType = "request_count"
	MetricRequestLatency MetricType = "request_latency"
	MetricErrorCount     MetricType = "error_count"
	MetricActiveSessions MetricType = "active_sessions"
	MetricQueueSize      MetricType = "queue_size"
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
	mu          sync.RWMutex
	metrics     map[MetricType][]Metric
	alerts      []*Alert
	blockchain  *block.Blockchain
	txPool      *transaction.TxPool
	alertChan   chan *Alert
	metricChan  chan Metric
	stopChan    chan struct{}
	collectors  map[MetricType]MetricCollector
	aggregators map[MetricType]MetricAggregator
	subscribers []MetricSubscriber
	interval    time.Duration
}

// MetricCollector defines how metrics are collected
type MetricCollector interface {
	Collect() (Metric, error)
}

// MetricAggregator defines how metrics are aggregated
type MetricAggregator interface {
	Aggregate(metrics []Metric) Metric
}

// MetricSubscriber receives metric updates
type MetricSubscriber interface {
	OnMetricUpdate(metric Metric)
}

// NewMonitor creates a new monitor
func NewMonitor(blockchain *block.Blockchain, txPool *transaction.TxPool, interval time.Duration) *Monitor {
	monitor := &Monitor{
		metrics:     make(map[MetricType][]Metric),
		alerts:      make([]*Alert, 0),
		blockchain:  blockchain,
		txPool:      txPool,
		alertChan:   make(chan *Alert, 100),
		metricChan:  make(chan Metric, 1000),
		stopChan:    make(chan struct{}),
		collectors:  make(map[MetricType]MetricCollector),
		aggregators: make(map[MetricType]MetricAggregator),
		subscribers: make([]MetricSubscriber, 0),
		interval:    interval,
	}

	// Initialize collectors
	monitor.initializeCollectors()
	monitor.initializeAggregators()

	return monitor
}

// Start starts the monitor
func (m *Monitor) Start(ctx context.Context) {
	// Start metric collection
	go m.collectMetrics(ctx)

	// Start alert processing
	go m.processAlerts()

	// Start metric processing
	go m.processMetrics()
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// AddMetric adds a new metric
func (m *Monitor) AddMetric(metric Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := m.metrics[metric.Type]
	metrics = append(metrics, metric)
	m.metrics[metric.Type] = metrics

	// Notify subscribers
	for _, subscriber := range m.subscribers {
		go subscriber.OnMetricUpdate(metric)
	}
}

// GetMetrics returns all metrics of a specific type
func (m *Monitor) GetMetrics(metricType MetricType) []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]Metric, len(m.metrics[metricType]))
	copy(metrics, m.metrics[metricType])
	return metrics
}

// GetAggregatedMetric returns an aggregated metric
func (m *Monitor) GetAggregatedMetric(metricType MetricType) (Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.metrics[metricType]
	if len(metrics) == 0 {
		return Metric{}, fmt.Errorf("no metrics available for type %s", metricType)
	}

	aggregator, exists := m.aggregators[metricType]
	if !exists {
		return Metric{}, fmt.Errorf("no aggregator available for type %s", metricType)
	}

	return aggregator.Aggregate(metrics), nil
}

// AddSubscriber adds a new metric subscriber
func (m *Monitor) AddSubscriber(subscriber MetricSubscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, subscriber)
}

// initializeCollectors sets up default metric collectors
func (m *Monitor) initializeCollectors() {
	// System metrics collectors
	m.collectors[MetricCPUUsage] = &CPUUsageCollector{}
	m.collectors[MetricMemoryUsage] = &MemoryUsageCollector{}
	m.collectors[MetricGoroutines] = &GoroutineCollector{}
	m.collectors[MetricThreadCount] = &ThreadCountCollector{}

	// Blockchain metrics collectors
	m.collectors[MetricBlockHeight] = &BlockHeightCollector{blockchain: m.blockchain}
	m.collectors[MetricTxCount] = &TxCountCollector{blockchain: m.blockchain}
	m.collectors[MetricMempoolSize] = &MempoolSizeCollector{txPool: m.txPool}
}

// initializeAggregators sets up default metric aggregators
func (m *Monitor) initializeAggregators() {
	// Default aggregator for most metrics
	defaultAggregator := &AverageAggregator{}

	for _, metricType := range []MetricType{
		MetricCPUUsage,
		MetricMemoryUsage,
		MetricNetworkLatency,
		MetricRequestLatency,
	} {
		m.aggregators[metricType] = defaultAggregator
	}

	// Sum aggregator for counters
	sumAggregator := &SumAggregator{}
	for _, metricType := range []MetricType{
		MetricRequestCount,
		MetricErrorCount,
		MetricNetworkBytes,
		MetricTxCount,
	} {
		m.aggregators[metricType] = sumAggregator
	}
}

// collectMetrics periodically collects metrics
func (m *Monitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.collectAllMetrics()
		}
	}
}

// collectAllMetrics collects metrics from all collectors
func (m *Monitor) collectAllMetrics() {
	for _, collector := range m.collectors {
		metric, err := collector.Collect()
		if err != nil {
			continue
		}
		m.AddMetric(metric)
	}
}

// processMetrics processes incoming metrics
func (m *Monitor) processMetrics() {
	for {
		select {
		case <-m.stopChan:
			return
		case metric := <-m.metricChan:
			m.AddMetric(metric)
			m.checkAlerts(metric)
		}
	}
}

// processAlerts processes incoming alerts
func (m *Monitor) processAlerts() {
	for {
		select {
		case <-m.stopChan:
			return
		case alert := <-m.alertChan:
			m.mu.Lock()
			m.alerts = append(m.alerts, alert)
			m.mu.Unlock()
		}
	}
}

// checkAlerts checks if any alerts should be triggered based on a metric
func (m *Monitor) checkAlerts(metric Metric) {
	switch metric.Type {
	case MetricCPUUsage:
		if metric.Value > 90 {
			m.alertChan <- &Alert{
				Level:     AlertCritical,
				Message:   fmt.Sprintf("High CPU usage: %.2f%%", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		} else if metric.Value > 75 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("CPU usage warning: %.2f%%", metric.Value),
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
		} else if metric.Value > 60 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("Memory usage warning: %.2f%%", metric.Value),
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
		} else if metric.Value > 75 {
			m.alertChan <- &Alert{
				Level:     AlertWarning,
				Message:   fmt.Sprintf("Disk usage warning: %.2f%%", metric.Value),
				Timestamp: time.Now(),
				Labels:    metric.Labels,
			}
		}
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
	}
}

// GetAlerts returns all alerts
func (m *Monitor) GetAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]*Alert, len(m.alerts))
	copy(alerts, m.alerts)
	return alerts
}

// GetAlertsByLevel returns alerts of a specific level
func (m *Monitor) GetAlertsByLevel(level AlertLevel) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filteredAlerts []*Alert
	for _, alert := range m.alerts {
		if alert.Level == level {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}
	return filteredAlerts
}

// CPUUsageCollector collects CPU usage metrics
type CPUUsageCollector struct{}

func (c *CPUUsageCollector) Collect() (Metric, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return Metric{
		Type:      MetricCPUUsage,
		Value:     float64(m.Sys) / float64(1024*1024), // Convert to MB
		Timestamp: time.Now(),
		Labels:    map[string]string{"unit": "MB"},
	}, nil
}

// MemoryUsageCollector collects memory usage metrics
type MemoryUsageCollector struct{}

func (c *MemoryUsageCollector) Collect() (Metric, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return Metric{
		Type:      MetricMemoryUsage,
		Value:     float64(m.Alloc) / float64(1024*1024), // Convert to MB
		Timestamp: time.Now(),
		Labels:    map[string]string{"unit": "MB"},
	}, nil
}

// GoroutineCollector collects goroutine count metrics
type GoroutineCollector struct{}

func (c *GoroutineCollector) Collect() (Metric, error) {
	return Metric{
		Type:      MetricGoroutines,
		Value:     float64(runtime.NumGoroutine()),
		Timestamp: time.Now(),
	}, nil
}

// ThreadCountCollector collects thread count metrics
type ThreadCountCollector struct{}

func (c *ThreadCountCollector) Collect() (Metric, error) {
	return Metric{
		Type:      MetricThreadCount,
		Value:     float64(runtime.NumCPU()),
		Timestamp: time.Now(),
	}, nil
}

// BlockHeightCollector collects block height metrics
type BlockHeightCollector struct {
	blockchain *block.Blockchain
}

func (c *BlockHeightCollector) Collect() (Metric, error) {
	bestBlock := c.blockchain.GetBestBlock()
	if bestBlock == nil {
		return Metric{}, nil
	}

	return Metric{
		Type:      MetricBlockHeight,
		Value:     float64(bestBlock.Header.Height),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"chain_type": string(bestBlock.Header.Type),
		},
	}, nil
}

// TxCountCollector collects transaction count metrics
type TxCountCollector struct {
	blockchain *block.Blockchain
}

func (c *TxCountCollector) Collect() (Metric, error) {
	return Metric{
		Type:      MetricTxCount,
		Value:     float64(c.blockchain.GetBlockCount()),
		Timestamp: time.Now(),
	}, nil
}

// MempoolSizeCollector collects mempool size metrics
type MempoolSizeCollector struct {
	txPool *transaction.TxPool
}

func (c *MempoolSizeCollector) Collect() (Metric, error) {
	return Metric{
		Type:      MetricMempoolSize,
		Value:     float64(c.txPool.GetSize()),
		Timestamp: time.Now(),
	}, nil
}

// AverageAggregator calculates the average of metrics
type AverageAggregator struct{}

func (a *AverageAggregator) Aggregate(metrics []Metric) Metric {
	if len(metrics) == 0 {
		return Metric{}
	}

	var sum float64
	for _, m := range metrics {
		sum += m.Value
	}

	return Metric{
		Type:      metrics[0].Type,
		Value:     sum / float64(len(metrics)),
		Timestamp: time.Now(),
		Labels:    metrics[0].Labels,
	}
}

// SumAggregator calculates the sum of metrics
type SumAggregator struct{}

func (a *SumAggregator) Aggregate(metrics []Metric) Metric {
	if len(metrics) == 0 {
		return Metric{}
	}

	var sum float64
	for _, m := range metrics {
		sum += m.Value
	}

	return Metric{
		Type:      metrics[0].Type,
		Value:     sum,
		Timestamp: time.Now(),
		Labels:    metrics[0].Labels,
	}
}
