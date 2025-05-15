package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupManager handles backup and restore operations
type BackupManager struct {
	db        Database
	backupDir string
}

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	BlockHeight uint64    `json:"block_height"`
	BlockHash   [32]byte  `json:"block_hash"`
	Checksum    string    `json:"checksum"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager(db Database, backupDir string) (*BackupManager, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &BackupManager{
		db:        db,
		backupDir: backupDir,
	}, nil
}

// CreateBackup creates a new backup of the blockchain data
func (bm *BackupManager) CreateBackup(ctx context.Context) error {
	// Get current chain state
	state, err := bm.db.GetChainState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("backup_%s", timestamp))
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create metadata
	metadata := BackupMetadata{
		Version:     "1.0",
		Timestamp:   time.Now(),
		BlockHeight: state.CurrentHeight,
		BlockHash:   state.CurrentHash,
	}

	// Backup blocks
	if err := bm.backupBlocks(ctx, backupPath, state.CurrentHeight); err != nil {
		return fmt.Errorf("failed to backup blocks: %w", err)
	}

	// Backup UTXO set
	if err := bm.backupUTXOs(ctx, backupPath); err != nil {
		return fmt.Errorf("failed to backup UTXO set: %w", err)
	}

	// Backup chain state
	if err := bm.backupChainState(ctx, backupPath, state); err != nil {
		return fmt.Errorf("failed to backup chain state: %w", err)
	}

	// Save metadata
	metadataFile := filepath.Join(backupPath, "metadata.json")
	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, metadataData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// RestoreBackup restores blockchain data from a backup
func (bm *BackupManager) RestoreBackup(ctx context.Context, backupPath string) error {
	// Read metadata
	metadataFile := filepath.Join(backupPath, "metadata.json")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Begin database transaction
	tx, err := bm.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer bm.db.Rollback(ctx, tx)

	// Restore blocks
	if err := bm.restoreBlocks(ctx, tx, backupPath); err != nil {
		return fmt.Errorf("failed to restore blocks: %w", err)
	}

	// Restore UTXO set
	if err := bm.restoreUTXOs(ctx, tx, backupPath); err != nil {
		return fmt.Errorf("failed to restore UTXO set: %w", err)
	}

	// Restore chain state
	if err := bm.restoreChainState(ctx, tx, backupPath); err != nil {
		return fmt.Errorf("failed to restore chain state: %w", err)
	}

	// Commit transaction
	if err := bm.db.Commit(ctx, tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListBackups returns a list of available backups
func (bm *BackupManager) ListBackups() ([]BackupMetadata, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metadataFile := filepath.Join(bm.backupDir, entry.Name(), "metadata.json")
		metadataData, err := os.ReadFile(metadataFile)
		if err != nil {
			continue // Skip backups with invalid metadata
		}

		var metadata BackupMetadata
		if err := json.Unmarshal(metadataData, &metadata); err != nil {
			continue // Skip backups with invalid metadata
		}

		backups = append(backups, metadata)
	}

	return backups, nil
}

// Helper functions

func (bm *BackupManager) backupBlocks(ctx context.Context, backupPath string, height uint64) error {
	blocksDir := filepath.Join(backupPath, "blocks")
	if err := os.MkdirAll(blocksDir, 0755); err != nil {
		return fmt.Errorf("failed to create blocks directory: %w", err)
	}

	// Backup blocks in batches
	batchSize := uint64(1000)
	for i := uint64(0); i < height; i += batchSize {
		end := i + batchSize
		if end > height {
			end = height
		}

		for j := i; j < end; j++ {
			block, err := bm.db.GetBlockByHeight(ctx, j)
			if err != nil {
				return fmt.Errorf("failed to get block at height %d: %w", j, err)
			}

			blockFile := filepath.Join(blocksDir, fmt.Sprintf("block_%d.json", j))
			blockData, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal block: %w", err)
			}

			if err := os.WriteFile(blockFile, blockData, 0644); err != nil {
				return fmt.Errorf("failed to write block file: %w", err)
			}
		}
	}

	return nil
}

func (bm *BackupManager) backupUTXOs(ctx context.Context, backupPath string) error {
	utxosDir := filepath.Join(backupPath, "utxos")
	if err := os.MkdirAll(utxosDir, 0755); err != nil {
		return fmt.Errorf("failed to create UTXOs directory: %w", err)
	}

	// Get all UTXOs
	utxos, err := bm.db.GetAllUTXOs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get UTXOs: %w", err)
	}

	// Write UTXOs to file
	utxosFile := filepath.Join(utxosDir, "utxos.json")
	utxosData, err := json.MarshalIndent(utxos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal UTXOs: %w", err)
	}

	if err := os.WriteFile(utxosFile, utxosData, 0644); err != nil {
		return fmt.Errorf("failed to write UTXOs file: %w", err)
	}

	return nil
}

func (bm *BackupManager) backupChainState(ctx context.Context, backupPath string, state *ChainState) error {
	stateFile := filepath.Join(backupPath, "chain_state.json")
	stateData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chain state: %w", err)
	}

	if err := os.WriteFile(stateFile, stateData, 0644); err != nil {
		return fmt.Errorf("failed to write chain state file: %w", err)
	}

	return nil
}

func (bm *BackupManager) restoreBlocks(ctx context.Context, tx DBTx, backupPath string) error {
	blocksDir := filepath.Join(backupPath, "blocks")
	entries, err := os.ReadDir(blocksDir)
	if err != nil {
		return fmt.Errorf("failed to read blocks directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		blockFile := filepath.Join(blocksDir, entry.Name())
		blockData, err := os.ReadFile(blockFile)
		if err != nil {
			return fmt.Errorf("failed to read block file: %w", err)
		}

		var block Block
		if err := json.Unmarshal(blockData, &block); err != nil {
			return fmt.Errorf("failed to unmarshal block: %w", err)
		}

		if err := bm.db.StoreBlock(ctx, &block); err != nil {
			return fmt.Errorf("failed to store block: %w", err)
		}
	}

	return nil
}

func (bm *BackupManager) restoreUTXOs(ctx context.Context, tx DBTx, backupPath string) error {
	utxosFile := filepath.Join(backupPath, "utxos", "utxos.json")
	utxosData, err := os.ReadFile(utxosFile)
	if err != nil {
		return fmt.Errorf("failed to read UTXOs file: %w", err)
	}

	var utxos []*DBUTXO
	if err := json.Unmarshal(utxosData, &utxos); err != nil {
		return fmt.Errorf("failed to unmarshal UTXOs: %w", err)
	}

	for _, utxo := range utxos {
		if err := bm.db.StoreUTXO(ctx, utxo.TxHash, utxo.OutputIndex, utxo); err != nil {
			return fmt.Errorf("failed to store UTXO: %w", err)
		}
	}

	return nil
}

func (bm *BackupManager) restoreChainState(ctx context.Context, tx DBTx, backupPath string) error {
	stateFile := filepath.Join(backupPath, "chain_state.json")
	stateData, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read chain state file: %w", err)
	}

	var state ChainState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return fmt.Errorf("failed to unmarshal chain state: %w", err)
	}

	if err := bm.db.StoreChainState(ctx, &state); err != nil {
		return fmt.Errorf("failed to store chain state: %w", err)
	}

	return nil
}
