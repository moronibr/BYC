package types

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultPersistencePath is the default path for persistence
	DefaultPersistencePath = "data/utxo"
	// DefaultPersistenceFile is the default file for persistence
	DefaultPersistenceFile = "utxo.dat"
	// DefaultPersistenceBackup is the default backup file for persistence
	DefaultPersistenceBackup = "utxo.dat.bak"
)

// PersistenceType represents the type of persistence
type PersistenceType byte

const (
	// PersistenceTypeNone indicates no persistence
	PersistenceTypeNone PersistenceType = iota
	// PersistenceTypeFile indicates file-based persistence
	PersistenceTypeFile
	// PersistenceTypeDatabase indicates database-based persistence
	PersistenceTypeDatabase
)

// UTXOPersistence handles persistence of the UTXO set
type UTXOPersistence struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Persistence state
	persistenceType PersistenceType
	path            string
	file            string
	backup          string
	fileHandle      *os.File
}

// NewUTXOPersistence creates a new UTXO persistence handler
func NewUTXOPersistence(utxoSet *UTXOSet) *UTXOPersistence {
	return &UTXOPersistence{
		utxoSet:         utxoSet,
		persistenceType: PersistenceTypeNone,
		path:            DefaultPersistencePath,
		file:            DefaultPersistenceFile,
		backup:          DefaultPersistenceBackup,
	}
}

// Save saves the UTXO set to disk
func (up *UTXOPersistence) Save() error {
	up.mu.Lock()
	defer up.mu.Unlock()

	// Check if persistence is enabled
	if up.persistenceType == PersistenceTypeNone {
		return fmt.Errorf("persistence is not enabled")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(up.path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create backup if file exists
	filePath := filepath.Join(up.path, up.file)
	backupPath := filepath.Join(up.path, up.backup)
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Rename(filePath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create encoder
	encoder := gob.NewEncoder(file)

	// Encode UTXO set
	if err := encoder.Encode(up.utxoSet); err != nil {
		return fmt.Errorf("failed to encode UTXO set: %v", err)
	}

	return nil
}

// Load loads the UTXO set from disk
func (up *UTXOPersistence) Load() error {
	up.mu.Lock()
	defer up.mu.Unlock()

	// Check if persistence is enabled
	if up.persistenceType == PersistenceTypeNone {
		return fmt.Errorf("persistence is not enabled")
	}

	// Check if file exists
	filePath := filepath.Join(up.path, up.file)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create decoder
	decoder := gob.NewDecoder(file)

	// Decode UTXO set
	if err := decoder.Decode(up.utxoSet); err != nil {
		// Try loading from backup
		backupPath := filepath.Join(up.path, up.backup)
		if _, err := os.Stat(backupPath); err == nil {
			backup, err := os.Open(backupPath)
			if err != nil {
				return fmt.Errorf("failed to open backup: %v", err)
			}
			defer backup.Close()

			// Create decoder
			decoder := gob.NewDecoder(backup)

			// Decode UTXO set
			if err := decoder.Decode(up.utxoSet); err != nil {
				return fmt.Errorf("failed to decode UTXO set from backup: %v", err)
			}

			// Restore backup
			if err := os.Rename(backupPath, filePath); err != nil {
				return fmt.Errorf("failed to restore backup: %v", err)
			}

			return nil
		}

		return fmt.Errorf("failed to decode UTXO set: %v", err)
	}

	return nil
}

// GetPersistenceStats returns statistics about the persistence
func (up *UTXOPersistence) GetPersistenceStats() *PersistenceStats {
	up.mu.RLock()
	defer up.mu.RUnlock()

	stats := &PersistenceStats{
		PersistenceType: up.persistenceType,
		Path:            up.path,
		File:            up.file,
		Backup:          up.backup,
	}

	// Get file size
	filePath := filepath.Join(up.path, up.file)
	if info, err := os.Stat(filePath); err == nil {
		stats.FileSize = info.Size()
	}

	// Get backup size
	backupPath := filepath.Join(up.path, up.backup)
	if info, err := os.Stat(backupPath); err == nil {
		stats.BackupSize = info.Size()
	}

	return stats
}

// SetPersistenceType sets the type of persistence
func (up *UTXOPersistence) SetPersistenceType(persistenceType PersistenceType) {
	up.mu.Lock()
	up.persistenceType = persistenceType
	up.mu.Unlock()
}

// SetPath sets the path for persistence
func (up *UTXOPersistence) SetPath(path string) {
	up.mu.Lock()
	up.path = path
	up.mu.Unlock()
}

// SetFile sets the file for persistence
func (up *UTXOPersistence) SetFile(file string) {
	up.mu.Lock()
	up.file = file
	up.mu.Unlock()
}

// SetBackup sets the backup file for persistence
func (up *UTXOPersistence) SetBackup(backup string) {
	up.mu.Lock()
	up.backup = backup
	up.mu.Unlock()
}

// PersistenceStats holds statistics about the persistence
type PersistenceStats struct {
	// PersistenceType is the type of persistence
	PersistenceType PersistenceType
	// Path is the path for persistence
	Path string
	// File is the file for persistence
	File string
	// Backup is the backup file for persistence
	Backup string
	// FileSize is the size of the persistence file in bytes
	FileSize int64
	// BackupSize is the size of the backup file in bytes
	BackupSize int64
}

// String returns a string representation of the persistence statistics
func (ps *PersistenceStats) String() string {
	return fmt.Sprintf(
		"Persistence Type: %d\n"+
			"Path: %s\n"+
			"File: %s (%d bytes)\n"+
			"Backup: %s (%d bytes)",
		ps.PersistenceType,
		ps.Path,
		ps.File, ps.FileSize,
		ps.Backup, ps.BackupSize,
	)
}
