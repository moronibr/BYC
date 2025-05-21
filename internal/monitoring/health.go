package monitoring

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/logger"
	"github.com/moroni/BYC/internal/network"
	"go.uber.org/zap"
)

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Details   struct {
		Blockchain struct {
			GoldenBlocks int  `json:"golden_blocks"`
			SilverBlocks int  `json:"silver_blocks"`
			IsSynced     bool `json:"is_synced"`
		} `json:"blockchain"`
		Network struct {
			Peers        int    `json:"peers"`
			IsConnected  bool   `json:"is_connected"`
			LastSyncTime string `json:"last_sync_time,omitempty"`
		} `json:"network"`
		System struct {
			MemoryUsage int64   `json:"memory_usage_bytes"`
			CPUUsage    float64 `json:"cpu_usage_percent"`
			DiskUsage   int64   `json:"disk_usage_bytes"`
		} `json:"system"`
	} `json:"details"`
}

// HealthCheck represents the health check system
type HealthCheck struct {
	blockchain *blockchain.Blockchain
	node       *network.Node
	lastSync   time.Time
}

// NewHealthCheck creates a new health check system
func NewHealthCheck(bc *blockchain.Blockchain, node *network.Node) *HealthCheck {
	return &HealthCheck{
		blockchain: bc,
		node:       node,
		lastSync:   time.Now(),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.checkHealth()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// checkHealth performs a health check of the system
func (h *HealthCheck) checkHealth() HealthStatus {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	// Check blockchain health
	status.Details.Blockchain.GoldenBlocks = len(h.blockchain.GoldenBlocks)
	status.Details.Blockchain.SilverBlocks = len(h.blockchain.SilverBlocks)
	status.Details.Blockchain.IsSynced = h.checkBlockchainSync()

	// Check network health
	status.Details.Network.Peers = len(h.node.Peers)
	status.Details.Network.IsConnected = status.Details.Network.Peers > 0
	if !h.lastSync.IsZero() {
		status.Details.Network.LastSyncTime = h.lastSync.Format(time.RFC3339)
	}

	// Check system health
	systemHealth := h.checkSystemHealth()
	status.Details.System = systemHealth

	// Update overall status
	if !status.Details.Blockchain.IsSynced || !status.Details.Network.IsConnected {
		status.Status = "degraded"
	}

	if systemHealth.MemoryUsage > 90 || systemHealth.CPUUsage > 90 || systemHealth.DiskUsage > 90 {
		status.Status = "critical"
	}

	return status
}

// checkBlockchainSync checks if the blockchain is in sync
func (h *HealthCheck) checkBlockchainSync() bool {
	// TODO: Implement actual blockchain sync check
	// This would typically compare our blockchain with peers
	return true
}

// checkSystemHealth checks the health of the system
func (h *HealthCheck) checkSystemHealth() struct {
	MemoryUsage int64   `json:"memory_usage_bytes"`
	CPUUsage    float64 `json:"cpu_usage_percent"`
	DiskUsage   int64   `json:"disk_usage_bytes"`
} {
	// TODO: Implement actual system health checks
	// This would typically use system-specific APIs or commands
	return struct {
		MemoryUsage int64   `json:"memory_usage_bytes"`
		CPUUsage    float64 `json:"cpu_usage_percent"`
		DiskUsage   int64   `json:"disk_usage_bytes"`
	}{
		MemoryUsage: 0,
		CPUUsage:    0,
		DiskUsage:   0,
	}
}

// UpdateLastSync updates the last sync time
func (h *HealthCheck) UpdateLastSync() {
	h.lastSync = time.Now()
	logger.Info("Blockchain sync completed", zap.Time("timestamp", h.lastSync))
}

// GetStatus returns the current health status
func (h *HealthCheck) GetStatus() map[string]interface{} {
	status := h.checkHealth()

	// Convert HealthStatus to map
	result := map[string]interface{}{
		"status":    status.Status,
		"timestamp": status.Timestamp,
		"details": map[string]interface{}{
			"blockchain": map[string]interface{}{
				"golden_blocks": status.Details.Blockchain.GoldenBlocks,
				"silver_blocks": status.Details.Blockchain.SilverBlocks,
				"is_synced":     status.Details.Blockchain.IsSynced,
			},
			"network": map[string]interface{}{
				"peers":          status.Details.Network.Peers,
				"is_connected":   status.Details.Network.IsConnected,
				"last_sync_time": status.Details.Network.LastSyncTime,
			},
			"system": map[string]interface{}{
				"memory_usage_bytes": status.Details.System.MemoryUsage,
				"cpu_usage_percent":  status.Details.System.CPUUsage,
				"disk_usage_bytes":   status.Details.System.DiskUsage,
			},
		},
	}

	return result
}
