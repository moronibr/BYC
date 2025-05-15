package monitoring

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupManager manages blockchain data backups
type BackupManager struct {
	backupDir string
	logger    *Logger
	metrics   *MetricsCollector
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Timestamp    time.Time `json:"timestamp"`
	BlockHeight  uint64    `json:"blockHeight"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	BackupPath   string    `json:"backupPath"`
	IsCompressed bool      `json:"isCompressed"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager(backupDir string, logger *Logger, metrics *MetricsCollector) (*BackupManager, error) {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	return &BackupManager{
		backupDir: backupDir,
		logger:    logger,
		metrics:   metrics,
	}, nil
}

// CreateBackup creates a backup of the blockchain data
func (bm *BackupManager) CreateBackup(dataDir string, blockHeight uint64) (*BackupInfo, error) {
	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("backup-%s.zip", timestamp))

	// Create zip file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %v", err)
	}
	defer zipFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through data directory
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header: %v", err)
		}

		// Set relative path
		relPath, err := filepath.Rel(dataDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		header.Name = relPath

		// Create zip file
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip file: %v", err)
		}

		// Open source file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %v", err)
		}
		defer file.Close()

		// Copy file contents
		if _, err := io.Copy(writer, file); err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %v", err)
	}

	// Get backup file info
	fileInfo, err := zipFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %v", err)
	}

	// Create backup info
	backupInfo := &BackupInfo{
		Timestamp:    time.Now(),
		BlockHeight:  blockHeight,
		Size:         fileInfo.Size(),
		BackupPath:   backupPath,
		IsCompressed: true,
	}

	// Save backup info
	infoPath := backupPath + ".json"
	infoFile, err := os.Create(infoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup info file: %v", err)
	}
	defer infoFile.Close()

	if err := json.NewEncoder(infoFile).Encode(backupInfo); err != nil {
		return nil, fmt.Errorf("failed to save backup info: %v", err)
	}

	bm.logger.Info("Created backup at %s (height: %d, size: %d bytes)",
		backupPath, blockHeight, fileInfo.Size())

	return backupInfo, nil
}

// RestoreBackup restores a backup
func (bm *BackupManager) RestoreBackup(backupPath string, targetDir string) error {
	// Open backup file
	zipFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer zipFile.Close()

	// Get zip file info
	zipInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get zip file info: %v", err)
	}

	// Open zip reader
	zipReader, err := zip.NewReader(zipFile, zipInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %v", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Extract files
	for _, file := range zipReader.File {
		// Create target file path
		targetPath := filepath.Join(targetDir, file.Name)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directories: %v", err)
		}

		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Open source file
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open source file: %v", err)
		}

		// Create target file
		dstFile, err := os.Create(targetPath)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create target file: %v", err)
		}

		// Copy file contents
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			return fmt.Errorf("failed to copy file contents: %v", err)
		}

		srcFile.Close()
		dstFile.Close()
	}

	bm.logger.Info("Restored backup from %s to %s", backupPath, targetDir)
	return nil
}

// ListBackups lists all available backups
func (bm *BackupManager) ListBackups() ([]*BackupInfo, error) {
	var backups []*BackupInfo

	// Walk through backup directory
	err := filepath.Walk(bm.backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-json files
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Open backup info file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open backup info file: %v", err)
		}
		defer file.Close()

		// Decode backup info
		var backupInfo BackupInfo
		if err := json.NewDecoder(file).Decode(&backupInfo); err != nil {
			return fmt.Errorf("failed to decode backup info: %v", err)
		}

		backups = append(backups, &backupInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}

	return backups, nil
}

// DeleteBackup deletes a backup
func (bm *BackupManager) DeleteBackup(backupPath string) error {
	// Delete backup file
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup file: %v", err)
	}

	// Delete backup info file
	infoPath := backupPath + ".json"
	if err := os.Remove(infoPath); err != nil {
		return fmt.Errorf("failed to delete backup info file: %v", err)
	}

	bm.logger.Info("Deleted backup %s", backupPath)
	return nil
}
