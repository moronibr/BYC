package monitoring

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"byc/internal/blockchain"
	"byc/internal/network"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Create test blockchain and node
	bc := blockchain.NewBlockchain()
	node, err := network.NewNode(&network.Config{})
	assert.NoError(t, err)

	// Create health check system
	health := NewHealthCheck(bc, node)

	// Test health check endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	health.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var status HealthStatus
	err = json.NewDecoder(w.Body).Decode(&status)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, 0, status.Details.Blockchain.GoldenBlocks)
	assert.Equal(t, 0, status.Details.Blockchain.SilverBlocks)
	assert.True(t, status.Details.Blockchain.IsSynced)
	assert.Equal(t, 0, status.Details.Network.Peers)
	assert.False(t, status.Details.Network.IsConnected)
}

func TestAlertSystem(t *testing.T) {
	// Create alert system
	alerts := NewAlertSystem("")

	// Test creating an alert
	alerts.CreateAlert(AlertLevelWarning, "Test alert", "test", nil)

	// Get all alerts
	allAlerts := alerts.GetAlerts()
	assert.Equal(t, 1, len(allAlerts))
	assert.Equal(t, AlertLevelWarning, allAlerts[0].Level)
	assert.Equal(t, "Test alert", allAlerts[0].Message)
	assert.Equal(t, "test", allAlerts[0].Component)
	assert.False(t, allAlerts[0].Resolved)

	// Test resolving an alert
	alerts.ResolveAlert(allAlerts[0].ID)

	// Get active alerts
	activeAlerts := alerts.GetActiveAlerts()
	assert.Equal(t, 0, len(activeAlerts))

	// Get all alerts again
	allAlerts = alerts.GetAlerts()
	assert.Equal(t, 1, len(allAlerts))
	assert.True(t, allAlerts[0].Resolved)
	assert.NotNil(t, allAlerts[0].ResolvedAt)
}

func TestMetrics(t *testing.T) {
	// Create test blockchain and node
	bc := blockchain.NewBlockchain()
	node, err := network.NewNode(&network.Config{})
	assert.NoError(t, err)

	// Create metrics system
	metrics := NewMetrics(bc, node)

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	metrics.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test recording metrics
	metrics.RecordPeerConnection()
	metrics.RecordPeerDisconnection()
	metrics.RecordNetworkLatency(100 * time.Millisecond)
	metrics.RecordNetworkError()
	metrics.RecordTransaction()

	// Start metrics collection
	metrics.Start()

	// Wait for metrics to be collected
	time.Sleep(100 * time.Millisecond)
}
