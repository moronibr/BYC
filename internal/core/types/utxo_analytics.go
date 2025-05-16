package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultAnalyticsInterval is the default interval for analytics in seconds
	DefaultAnalyticsInterval = 300 // 5 minutes
	// DefaultRetentionPeriod is the default retention period for analytics data
	DefaultRetentionPeriod = 7 * 24 * time.Hour // 7 days
	// DefaultMaxDataPoints is the default maximum number of data points to keep
	DefaultMaxDataPoints = 2016 // 7 days at 5-minute intervals
)

// UTXOAnalytics handles analytics of the UTXO set
type UTXOAnalytics struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Analytics state
	lastUpdate    time.Time
	interval      time.Duration
	retention     time.Duration
	maxDataPoints int
	dataPoints    []*DataPoint
	metrics       *AnalyticsMetrics
}

// NewUTXOAnalytics creates a new UTXO analytics handler
func NewUTXOAnalytics(utxoSet *UTXOSet) *UTXOAnalytics {
	return &UTXOAnalytics{
		utxoSet:       utxoSet,
		interval:      DefaultAnalyticsInterval * time.Second,
		retention:     DefaultRetentionPeriod,
		maxDataPoints: DefaultMaxDataPoints,
		dataPoints:    make([]*DataPoint, 0),
		metrics:       NewAnalyticsMetrics(),
	}
}

// UpdateAnalytics updates the analytics data
func (ua *UTXOAnalytics) UpdateAnalytics() error {
	ua.mu.Lock()
	defer ua.mu.Unlock()

	// Check if enough time has passed since last update
	if time.Since(ua.lastUpdate) < ua.interval {
		return nil
	}

	// Create new data point
	dp := &DataPoint{
		Timestamp: time.Now(),
		Metrics:   ua.collectMetrics(),
	}

	// Add data point
	ua.dataPoints = append(ua.dataPoints, dp)

	// Remove old data points
	ua.pruneDataPoints()

	// Update metrics
	ua.updateMetrics()

	// Update last update time
	ua.lastUpdate = dp.Timestamp

	return nil
}

// collectMetrics collects metrics from the UTXO set
func (ua *UTXOAnalytics) collectMetrics() *Metrics {
	// Get UTXO set stats
	stats := ua.utxoSet.GetStats()

	return &Metrics{
		TotalUTXOs:        stats.TotalUTXOs,
		SpentUTXOs:        stats.SpentUTXOs,
		UnspentUTXOs:      stats.UnspentUTXOs,
		TotalValue:        stats.TotalValue,
		AverageValue:      stats.TotalValue / float64(stats.TotalUTXOs),
		MaxValue:          stats.MaxValue,
		MinValue:          stats.MinValue,
		MedianValue:       stats.MedianValue,
		ValueDistribution: stats.ValueDistribution,
		AgeDistribution:   stats.AgeDistribution,
		SizeDistribution:  stats.SizeDistribution,
	}
}

// pruneDataPoints removes old data points
func (ua *UTXOAnalytics) pruneDataPoints() {
	// Remove data points older than retention period
	cutoff := time.Now().Add(-ua.retention)
	for i := 0; i < len(ua.dataPoints); i++ {
		if ua.dataPoints[i].Timestamp.Before(cutoff) {
			ua.dataPoints = ua.dataPoints[i+1:]
			break
		}
	}

	// Remove excess data points
	if len(ua.dataPoints) > ua.maxDataPoints {
		ua.dataPoints = ua.dataPoints[len(ua.dataPoints)-ua.maxDataPoints:]
	}
}

// updateMetrics updates the analytics metrics
func (ua *UTXOAnalytics) updateMetrics() {
	// Reset metrics
	ua.metrics = NewAnalyticsMetrics()

	// Update metrics from data points
	for _, dp := range ua.dataPoints {
		ua.metrics.Update(dp.Metrics)
	}
}

// GetAnalytics returns the analytics data
func (ua *UTXOAnalytics) GetAnalytics() *Analytics {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	return &Analytics{
		LastUpdate:    ua.lastUpdate,
		Interval:      ua.interval,
		Retention:     ua.retention,
		MaxDataPoints: ua.maxDataPoints,
		DataPoints:    ua.dataPoints,
		Metrics:       ua.metrics,
	}
}

// SetAnalyticsInterval sets the analytics update interval
func (ua *UTXOAnalytics) SetAnalyticsInterval(interval time.Duration) {
	ua.mu.Lock()
	ua.interval = interval
	ua.mu.Unlock()
}

// SetRetentionPeriod sets the retention period for analytics data
func (ua *UTXOAnalytics) SetRetentionPeriod(retention time.Duration) {
	ua.mu.Lock()
	ua.retention = retention
	ua.mu.Unlock()
}

// SetMaxDataPoints sets the maximum number of data points to keep
func (ua *UTXOAnalytics) SetMaxDataPoints(maxDataPoints int) {
	ua.mu.Lock()
	ua.maxDataPoints = maxDataPoints
	ua.mu.Unlock()
}

// DataPoint represents a single analytics data point
type DataPoint struct {
	// Timestamp is the time when the data point was collected
	Timestamp time.Time
	// Metrics contains the metrics for this data point
	Metrics *Metrics
}

// Metrics holds the metrics for a data point
type Metrics struct {
	// TotalUTXOs is the total number of UTXOs
	TotalUTXOs int64
	// SpentUTXOs is the number of spent UTXOs
	SpentUTXOs int64
	// UnspentUTXOs is the number of unspent UTXOs
	UnspentUTXOs int64
	// TotalValue is the total value of all UTXOs
	TotalValue float64
	// AverageValue is the average value of UTXOs
	AverageValue float64
	// MaxValue is the maximum value of UTXOs
	MaxValue float64
	// MinValue is the minimum value of UTXOs
	MinValue float64
	// MedianValue is the median value of UTXOs
	MedianValue float64
	// ValueDistribution is the distribution of UTXO values
	ValueDistribution map[string]int64
	// AgeDistribution is the distribution of UTXO ages
	AgeDistribution map[string]int64
	// SizeDistribution is the distribution of UTXO sizes
	SizeDistribution map[string]int64
}

// AnalyticsMetrics holds aggregated analytics metrics
type AnalyticsMetrics struct {
	// TotalUTXOs is the total number of UTXOs
	TotalUTXOs int64
	// SpentUTXOs is the number of spent UTXOs
	SpentUTXOs int64
	// UnspentUTXOs is the number of unspent UTXOs
	UnspentUTXOs int64
	// TotalValue is the total value of all UTXOs
	TotalValue float64
	// AverageValue is the average value of UTXOs
	AverageValue float64
	// MaxValue is the maximum value of UTXOs
	MaxValue float64
	// MinValue is the minimum value of UTXOs
	MinValue float64
	// MedianValue is the median value of UTXOs
	MedianValue float64
	// ValueDistribution is the distribution of UTXO values
	ValueDistribution map[string]int64
	// AgeDistribution is the distribution of UTXO ages
	AgeDistribution map[string]int64
	// SizeDistribution is the distribution of UTXO sizes
	SizeDistribution map[string]int64
}

// NewAnalyticsMetrics creates new analytics metrics
func NewAnalyticsMetrics() *AnalyticsMetrics {
	return &AnalyticsMetrics{
		ValueDistribution: make(map[string]int64),
		AgeDistribution:   make(map[string]int64),
		SizeDistribution:  make(map[string]int64),
	}
}

// Update updates the analytics metrics with new metrics
func (am *AnalyticsMetrics) Update(metrics *Metrics) {
	am.TotalUTXOs = metrics.TotalUTXOs
	am.SpentUTXOs = metrics.SpentUTXOs
	am.UnspentUTXOs = metrics.UnspentUTXOs
	am.TotalValue = metrics.TotalValue
	am.AverageValue = metrics.AverageValue
	am.MaxValue = metrics.MaxValue
	am.MinValue = metrics.MinValue
	am.MedianValue = metrics.MedianValue

	// Update distributions
	for k, v := range metrics.ValueDistribution {
		am.ValueDistribution[k] = v
	}
	for k, v := range metrics.AgeDistribution {
		am.AgeDistribution[k] = v
	}
	for k, v := range metrics.SizeDistribution {
		am.SizeDistribution[k] = v
	}
}

// Analytics holds the complete analytics data
type Analytics struct {
	// LastUpdate is the time of the last update
	LastUpdate time.Time
	// Interval is the update interval
	Interval time.Duration
	// Retention is the retention period
	Retention time.Duration
	// MaxDataPoints is the maximum number of data points
	MaxDataPoints int
	// DataPoints contains all data points
	DataPoints []*DataPoint
	// Metrics contains the aggregated metrics
	Metrics *AnalyticsMetrics
}

// String returns a string representation of the analytics
func (a *Analytics) String() string {
	return fmt.Sprintf(
		"Last Update: %v, Interval: %v\n"+
			"Retention: %v, Max Data Points: %d\n"+
			"Data Points: %d\n"+
			"Metrics:\n"+
			"  Total UTXOs: %d\n"+
			"  Spent UTXOs: %d\n"+
			"  Unspent UTXOs: %d\n"+
			"  Total Value: %.2f\n"+
			"  Average Value: %.2f\n"+
			"  Max Value: %.2f\n"+
			"  Min Value: %.2f\n"+
			"  Median Value: %.2f",
		a.LastUpdate.Format("2006-01-02 15:04:05"),
		a.Interval, a.Retention, a.MaxDataPoints,
		len(a.DataPoints),
		a.Metrics.TotalUTXOs,
		a.Metrics.SpentUTXOs,
		a.Metrics.UnspentUTXOs,
		a.Metrics.TotalValue,
		a.Metrics.AverageValue,
		a.Metrics.MaxValue,
		a.Metrics.MinValue,
		a.Metrics.MedianValue,
	)
}
