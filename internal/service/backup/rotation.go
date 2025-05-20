// Package backup provides functionality for managing backups
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CleanupOldBackups removes older backups, keeping only the specified number of most recent ones
// It deletes older backups that match the prefix and extension pattern.
func CleanupOldBackups(backupDir string, prefix string, maxBackups int) error {
	// Read all files in the backup directory
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("error reading backup directory: %w", err)
	}

	// Filter for backup files with matching prefix and .tar.gz extension
	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() &&
			strings.HasPrefix(file.Name(), prefix) &&
			strings.HasSuffix(file.Name(), ".tar.gz") {
			backupFiles = append(backupFiles, file)
		}
	}

	// If we don't have more backups than the limit, no need to delete any
	if len(backupFiles) <= maxBackups {
		return nil
	}

	// Sort files by modification time (oldest first)
	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, err := backupFiles[i].Info()
		if err != nil {
			return false
		}
		infoJ, err := backupFiles[j].Info()
		if err != nil {
			return true
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Delete older backups
	filesToDelete := backupFiles[:len(backupFiles)-maxBackups]
	for _, file := range filesToDelete {
		filePath := filepath.Join(backupDir, file.Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("  Warning: Failed to delete old backup %s: %v\n", filePath, err)
		} else {
			fmt.Printf("  Deleted old backup: %s\n", filePath)
		}
	}

	return nil
}
