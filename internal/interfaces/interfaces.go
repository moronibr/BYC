package interfaces

import (
	"time"
)

// BackupManager defines the interface for backup operations
type BackupManager interface {
	CreateBackup() (*BackupInfo, error)
	RestoreBackup(name string) error
	ListBackups() []string
	DeleteBackup(name string) error
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID        string
	Timestamp time.Time
	Size      int64
	Checksum  string
}

// MaintenanceManager defines the interface for maintenance operations
type MaintenanceManager interface {
	GetHealth() *SystemHealth
	Start() error
	GetLogs() []MaintenanceLog
	SetSchedule(schedule string) error
	GetTasks() []MaintenanceTask
	SetAlert(email string) error
}

// VersionManager defines the interface for version operations
type VersionManager interface {
	GetCurrentVersion() string
	GetVersionHistory() []VersionInfo
	Upgrade(targetVersion string) error
}

// Wallet defines the interface for wallet operations
type Wallet interface {
	CreateEphraimCoin(bc interface{}) error
	CreateManassehCoin(bc interface{}) error
	CreateJosephCoin(bc interface{}) error
	GetSpecialCoins(bc interface{}) []SpecialCoin
}

// SpecialCoin represents a special coin type
type SpecialCoin struct {
	Type   string
	Amount int64
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

// MaintenanceLog represents a maintenance log entry
type MaintenanceLog struct {
	Timestamp time.Time
	Message   string
}

// MaintenanceTask represents a maintenance task
type MaintenanceTask struct {
	Name        string
	Description string
}

// VersionInfo represents version information
type VersionInfo struct {
	Number string
	Date   time.Time
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	BackupDir string
	Encrypt   bool
	Compress  bool
}

// NewBackupManager creates a new backup manager
func NewBackupManager(config *BackupConfig) (BackupManager, error) {
	return nil, nil // TODO: Implement
}

// NewMaintenanceManager creates a new maintenance manager
func NewMaintenanceManager() MaintenanceManager {
	return nil // TODO: Implement
}

// NewWallet creates a new wallet
func NewWallet() Wallet {
	return nil // TODO: Implement
}

// NewVersionManager creates a new version manager
func NewVersionManager() VersionManager {
	return nil // TODO: Implement
}
