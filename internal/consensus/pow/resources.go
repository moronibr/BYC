package pow

import (
	"runtime"
	"sync"
	"time"
)

// ResourceManager manages system resources for mining
type ResourceManager struct {
	mu sync.RWMutex

	// System metrics
	cpuCount     int
	memStats     runtime.MemStats
	lastUpdate   time.Time
	updatePeriod time.Duration

	// Resource limits
	maxCPUPercent    float64
	maxMemoryPercent float64
	minWorkers       int
	maxWorkers       int

	// Current state
	currentWorkers int
	workerLoads    map[int]float64 // Worker count -> CPU load mapping
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	rm := &ResourceManager{
		cpuCount:         runtime.NumCPU(),
		updatePeriod:     time.Second * 5,
		maxCPUPercent:    0.8, // 80% CPU usage limit
		maxMemoryPercent: 0.8, // 80% memory usage limit
		minWorkers:       1,
		maxWorkers:       runtime.NumCPU() * 2,
		workerLoads:      make(map[int]float64),
	}

	// Start resource monitoring
	go rm.monitorResources()

	return rm
}

// GetOptimalWorkerCount returns the optimal number of workers based on system resources
func (rm *ResourceManager) GetOptimalWorkerCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Get current system metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate memory usage percentage
	memUsagePercent := float64(memStats.Alloc) / float64(memStats.Sys)

	// If memory usage is too high, reduce workers
	if memUsagePercent > rm.maxMemoryPercent {
		return max(rm.minWorkers, rm.currentWorkers-1)
	}

	// Find the worker count with the best performance/load ratio
	bestWorkerCount := rm.minWorkers
	bestRatio := 0.0

	for workers, load := range rm.workerLoads {
		if workers < rm.minWorkers || workers > rm.maxWorkers {
			continue
		}

		// Calculate performance/load ratio
		// Higher ratio means better efficiency
		ratio := float64(workers) / load
		if ratio > bestRatio {
			bestRatio = ratio
			bestWorkerCount = workers
		}
	}

	return bestWorkerCount
}

// UpdateWorkerLoad updates the CPU load for a specific worker count
func (rm *ResourceManager) UpdateWorkerLoad(workerCount int, load float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.workerLoads[workerCount] = load
	rm.currentWorkers = workerCount
}

// monitorResources continuously monitors system resources
func (rm *ResourceManager) monitorResources() {
	ticker := time.NewTicker(rm.updatePeriod)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.Lock()

		// Update memory stats
		runtime.ReadMemStats(&rm.memStats)
		rm.lastUpdate = time.Now()

		// Clean up old worker load data
		for workers := range rm.workerLoads {
			if workers < rm.minWorkers || workers > rm.maxWorkers {
				delete(rm.workerLoads, workers)
			}
		}

		rm.mu.Unlock()
	}
}

// GetSystemMetrics returns current system metrics
func (rm *ResourceManager) GetSystemMetrics() (cpuCount int, memUsage float64) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.cpuCount, float64(rm.memStats.Alloc) / float64(rm.memStats.Sys)
}

// SetResourceLimits sets the resource usage limits
func (rm *ResourceManager) SetResourceLimits(maxCPU, maxMemory float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.maxCPUPercent = maxCPU
	rm.maxMemoryPercent = maxMemory
}

// SetWorkerLimits sets the minimum and maximum number of workers
func (rm *ResourceManager) SetWorkerLimits(min, max int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.minWorkers = min
	rm.maxWorkers = max
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
