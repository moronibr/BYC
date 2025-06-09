package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"byc/internal/logger"

	"go.uber.org/zap"
)

// AlertLevel represents the severity level of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents a system alert
type Alert struct {
	ID         string      `json:"id"`
	Level      AlertLevel  `json:"level"`
	Message    string      `json:"message"`
	Component  string      `json:"component"`
	Timestamp  time.Time   `json:"timestamp"`
	Details    interface{} `json:"details,omitempty"`
	Resolved   bool        `json:"resolved"`
	ResolvedAt *time.Time  `json:"resolved_at,omitempty"`
}

// AlertSystem represents the alert system
type AlertSystem struct {
	alerts     map[string]*Alert
	handlers   []AlertHandler
	mu         sync.RWMutex
	webhookURL string
}

// AlertHandler is a function that handles alerts
type AlertHandler func(alert *Alert)

// NewAlertSystem creates a new alert system
func NewAlertSystem(webhookURL string) *AlertSystem {
	return &AlertSystem{
		alerts:     make(map[string]*Alert),
		handlers:   make([]AlertHandler, 0),
		webhookURL: webhookURL,
	}
}

// RegisterHandler registers a new alert handler
func (a *AlertSystem) RegisterHandler(handler AlertHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handlers = append(a.handlers, handler)
}

// CreateAlert creates a new alert
func (a *AlertSystem) CreateAlert(level AlertLevel, message, component string, details interface{}) {
	alert := &Alert{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Level:     level,
		Message:   message,
		Component: component,
		Timestamp: time.Now(),
		Details:   details,
		Resolved:  false,
	}

	a.mu.Lock()
	a.alerts[alert.ID] = alert
	a.mu.Unlock()

	// Log the alert
	logger.Info("Alert created",
		zap.String("id", alert.ID),
		zap.String("level", string(alert.Level)),
		zap.String("message", alert.Message),
		zap.String("component", alert.Component),
	)

	// Notify handlers
	for _, handler := range a.handlers {
		handler(alert)
	}

	// Send webhook notification if URL is configured
	if a.webhookURL != "" {
		go a.sendWebhookNotification(alert)
	}
}

// ResolveAlert resolves an alert
func (a *AlertSystem) ResolveAlert(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if alert, exists := a.alerts[id]; exists {
		now := time.Now()
		alert.Resolved = true
		alert.ResolvedAt = &now

		logger.Info("Alert resolved",
			zap.String("id", alert.ID),
			zap.String("level", string(alert.Level)),
			zap.String("message", alert.Message),
			zap.String("component", alert.Component),
		)

		// Notify handlers
		for _, handler := range a.handlers {
			handler(alert)
		}

		// Send webhook notification if URL is configured
		if a.webhookURL != "" {
			go a.sendWebhookNotification(alert)
		}
	}
}

// GetAlerts returns all alerts
func (a *AlertSystem) GetAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alerts := make([]*Alert, 0, len(a.alerts))
	for _, alert := range a.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetActiveAlerts returns all unresolved alerts
func (a *AlertSystem) GetActiveAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range a.alerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// sendWebhookNotification sends an alert notification to the configured webhook URL
func (a *AlertSystem) sendWebhookNotification(alert *Alert) {
	payload, err := json.Marshal(alert)
	if err != nil {
		logger.Error("Failed to marshal alert for webhook",
			zap.Error(err),
			zap.String("alert_id", alert.ID),
		)
		return
	}

	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("Failed to send webhook notification",
			zap.Error(err),
			zap.String("alert_id", alert.ID),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Webhook notification failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("alert_id", alert.ID),
		)
	}
}
