package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultBackupInterval is the default interval for backups in seconds
	DefaultBackupInterval = 3600 // 1 hour
	// DefaultMaxBackups is the default maximum number of backups to keep
	DefaultMaxBackups = 24 // 24 hours of backups
	// DefaultBackupDir is the default directory for backups
	DefaultBackupDir = "backups"
	// DefaultBackupPrefix is the default prefix for backup files
	DefaultBackupPrefix = "utxo_backup_"
)

// UTXOBackup handles backup of the UTXO set
type UTXOBackup struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Backup state
	lastBackup   time.Time
	interval     time.Duration
	maxBackups   int
	backupDir    string
	backupPrefix string
	backupFiles  []string
}

// NewUTXOBackup creates a new UTXO backup handler
func NewUTXOBackup(utxoSet *UTXOSet) *UTXOBackup {
	return &UTXOBackup{
		utxoSet:      utxoSet,
		interval:     DefaultBackupInterval * time.Second,
		maxBackups:   DefaultMaxBackups,
		backupDir:    DefaultBackupDir,
		backupPrefix: DefaultBackupPrefix,
		backupFiles:  make([]string, 0),
	}
}

// CreateBackup creates a new backup of the UTXO set
func (ub *UTXOBackup) CreateBackup() error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	// Check if enough time has passed since last backup
	if time.Since(ub.lastBackup) < ub.interval {
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(ub.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Generate backup filename
	filename := filepath.Join(ub.backupDir, fmt.Sprintf("%s%d.bak", ub.backupPrefix, time.Now().Unix()))

	// Get UTXO set data
	data := ub.utxoSet.Serialize()

	// Calculate hash
	hash := sha256.Sum256(data)

	// Create backup file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	// Write backup metadata
	metadata := make([]byte, 0, 37)
	metadata = append(metadata, byte(1)) // Version
	binary.LittleEndian.PutUint32(metadata[1:5], uint32(len(data)))
	metadata = append(metadata, hash[:]...)
	if _, err := file.Write(metadata); err != nil {
		return fmt.Errorf("failed to write backup metadata: %v", err)
	}

	// Write UTXO set data
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write backup data: %v", err)
	}

	// Add backup file to list
	ub.backupFiles = append(ub.backupFiles, filename)

	// Remove old backups
	ub.pruneBackups()

	// Update last backup time
	ub.lastBackup = time.Now()

	return nil
}

// RestoreBackup restores the UTXO set from a backup
func (ub *UTXOBackup) RestoreBackup(filename string) error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	// Open backup file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer file.Close()

	// Read backup metadata
	metadata := make([]byte, 37)
	if _, err := io.ReadFull(file, metadata); err != nil {
		return fmt.Errorf("failed to read backup metadata: %v", err)
	}

	// Check version
	version := metadata[0]
	if version != 1 {
		return fmt.Errorf("unsupported version: %d", version)
	}

	// Get data size
	size := binary.LittleEndian.Uint32(metadata[1:5])
	expectedHash := metadata[5:37]

	// Read UTXO set data
	data := make([]byte, size)
	if _, err := io.ReadFull(file, data); err != nil {
		return fmt.Errorf("failed to read backup data: %v", err)
	}

	// Verify hash
	hash := sha256.Sum256(data)
	if !bytes.Equal(hash[:], expectedHash) {
		return fmt.Errorf("backup verification failed")
	}

	// Deserialize UTXO set
	utxoSet, err := DeserializeUTXOSet(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize UTXO set: %v", err)
	}

	// Update UTXO set
	ub.utxoSet = utxoSet

	return nil
}

// pruneBackups removes old backups
func (ub *UTXOBackup) pruneBackups() {
	// Remove excess backups
	if len(ub.backupFiles) > ub.maxBackups {
		// Remove oldest backups
		for _, filename := range ub.backupFiles[:len(ub.backupFiles)-ub.maxBackups] {
			os.Remove(filename)
		}
		ub.backupFiles = ub.backupFiles[len(ub.backupFiles)-ub.maxBackups:]
	}
}

// GetBackupStats returns statistics about the backups
func (ub *UTXOBackup) GetBackupStats() *BackupStats {
	ub.mu.RLock()
	defer ub.mu.RUnlock()

	stats := &BackupStats{
		LastBackup:   ub.lastBackup,
		Interval:     ub.interval,
		MaxBackups:   ub.maxBackups,
		BackupCount:  len(ub.backupFiles),
		BackupDir:    ub.backupDir,
		BackupPrefix: ub.backupPrefix,
	}

	// Calculate total size
	for _, filename := range ub.backupFiles {
		if info, err := os.Stat(filename); err == nil {
			stats.TotalSize += info.Size()
		}
	}

	return stats
}

// SetBackupInterval sets the backup interval
func (ub *UTXOBackup) SetBackupInterval(interval time.Duration) {
	ub.mu.Lock()
	ub.interval = interval
	ub.mu.Unlock()
}

// SetMaxBackups sets the maximum number of backups to keep
func (ub *UTXOBackup) SetMaxBackups(maxBackups int) {
	ub.mu.Lock()
	ub.maxBackups = maxBackups
	ub.mu.Unlock()
}

// SetBackupDir sets the backup directory
func (ub *UTXOBackup) SetBackupDir(dir string) {
	ub.mu.Lock()
	ub.backupDir = dir
	ub.mu.Unlock()
}

// SetBackupPrefix sets the backup file prefix
func (ub *UTXOBackup) SetBackupPrefix(prefix string) {
	ub.mu.Lock()
	ub.backupPrefix = prefix
	ub.mu.Unlock()
}

// BackupStats holds statistics about the backups
type BackupStats struct {
	// LastBackup is the time of the last backup
	LastBackup time.Time
	// Interval is the backup interval
	Interval time.Duration
	// MaxBackups is the maximum number of backups
	MaxBackups int
	// BackupCount is the current number of backups
	BackupCount int
	// TotalSize is the total size of all backups in bytes
	TotalSize int64
	// BackupDir is the backup directory
	BackupDir string
	// BackupPrefix is the backup file prefix
	BackupPrefix string
}

// String returns a string representation of the backup statistics
func (bs *BackupStats) String() string {
	return fmt.Sprintf(
		"Last Backup: %v, Interval: %v\n"+
			"Max Backups: %d, Current: %d\n"+
			"Total Size: %d bytes\n"+
			"Directory: %s, Prefix: %s",
		bs.LastBackup.Format("2006-01-02 15:04:05"),
		bs.Interval, bs.MaxBackups, bs.BackupCount,
		bs.TotalSize, bs.BackupDir, bs.BackupPrefix,
	)
}
