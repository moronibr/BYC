package maintenance

import (
	"fmt"
	"sync"
	"time"
)

// MaintenanceTask represents a maintenance task
type MaintenanceTask struct {
	ID          string
	Name        string
	Description string
	Schedule    time.Duration
	LastRun     time.Time
	NextRun     time.Time
	Handler     func() error
}

// SystemHealth represents system health status
type SystemHealth struct {
	Status     string
	LastCheck  time.Time
	Components map[string]ComponentHealth
	LastError  error
}

// ComponentHealth represents health status of a component
type ComponentHealth struct {
	Status    string
	LastCheck time.Time
	Error     error
}

// MaintenanceManager handles system maintenance tasks
type MaintenanceManager struct {
	mu          sync.RWMutex
	tasks       map[string]*MaintenanceTask
	health      *SystemHealth
	stopChan    chan struct{}
	isRunning   bool
	healthCheck func() error
}

// NewMaintenanceManager creates a new maintenance manager
func NewMaintenanceManager() *MaintenanceManager {
	return &MaintenanceManager{
		tasks:    make(map[string]*MaintenanceTask),
		health:   &SystemHealth{Components: make(map[string]ComponentHealth)},
		stopChan: make(chan struct{}),
	}
}

// RegisterTask registers a new maintenance task
func (mm *MaintenanceManager) RegisterTask(task *MaintenanceTask) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	task.NextRun = time.Now().Add(task.Schedule)
	mm.tasks[task.ID] = task
	return nil
}

// Start starts the maintenance manager
func (mm *MaintenanceManager) Start() error {
	mm.mu.Lock()
	if mm.isRunning {
		mm.mu.Unlock()
		return fmt.Errorf("maintenance manager already running")
	}
	mm.isRunning = true
	mm.mu.Unlock()

	go mm.runMaintenanceLoop()
	return nil
}

// Stop stops the maintenance manager
func (mm *MaintenanceManager) Stop() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if !mm.isRunning {
		return
	}

	close(mm.stopChan)
	mm.isRunning = false
}

// GetHealth returns the current system health status
func (mm *MaintenanceManager) GetHealth() *SystemHealth {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	health := *mm.health
	health.Components = make(map[string]ComponentHealth)
	for k, v := range mm.health.Components {
		health.Components[k] = v
	}
	return &health
}

// runMaintenanceLoop runs the maintenance loop
func (mm *MaintenanceManager) runMaintenanceLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.runScheduledTasks()
			mm.checkSystemHealth()
		case <-mm.stopChan:
			return
		}
	}
}

// runScheduledTasks runs all scheduled maintenance tasks
func (mm *MaintenanceManager) runScheduledTasks() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	for _, task := range mm.tasks {
		if now.After(task.NextRun) {
			go func(t *MaintenanceTask) {
				if err := t.Handler(); err != nil {
					mm.updateComponentHealth(t.ID, "error", err)
				} else {
					mm.updateComponentHealth(t.ID, "healthy", nil)
				}
				t.LastRun = now
				t.NextRun = now.Add(t.Schedule)
			}(task)
		}
	}
}

// checkSystemHealth performs system health check
func (mm *MaintenanceManager) checkSystemHealth() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.healthCheck != nil {
		if err := mm.healthCheck(); err != nil {
			mm.health.Status = "unhealthy"
			mm.health.LastError = err
		} else {
			mm.health.Status = "healthy"
			mm.health.LastError = nil
		}
	}
	mm.health.LastCheck = time.Now()
}

// updateComponentHealth updates the health status of a component
func (mm *MaintenanceManager) updateComponentHealth(componentID, status string, err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.health.Components[componentID] = ComponentHealth{
		Status:    status,
		LastCheck: time.Now(),
		Error:     err,
	}
}

// SetHealthCheck sets the system health check function
func (mm *MaintenanceManager) SetHealthCheck(check func() error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.healthCheck = check
}
