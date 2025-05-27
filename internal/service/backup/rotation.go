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

	// Filter for backup files with matching prefix and .tar.gz extension (possibly with .gpg)
	var backupFiles []os.DirEntry
	for _, file := range files {
		fileName := file.Name()
		if !file.IsDir() &&
			strings.HasPrefix(fileName, prefix) &&
			(strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tar.gz.gpg")) {
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

	// Delete older backups and their associated config files
	filesToDelete := backupFiles[:len(backupFiles)-maxBackups]
	for _, file := range filesToDelete {
		fileName := file.Name()
		backupFilePath := filepath.Join(backupDir, fileName)

		// Delete the backup file
		if err := os.Remove(backupFilePath); err != nil {
			fmt.Printf("  Warning: Failed to delete old backup %s: %v\n", backupFilePath, err)
		} else {
			fmt.Printf("  Deleted old backup: %s\n", backupFilePath)
		}

		// Check for and delete any associated config file
		// Extract the base name for the config file by removing extensions
		configBaseName := fileName
		// Handle .tar.gz.gpg case
		if strings.HasSuffix(configBaseName, ".tar.gz.gpg") {
			configBaseName = strings.TrimSuffix(configBaseName, ".tar.gz.gpg")
		} else if strings.HasSuffix(configBaseName, ".tar.gz") {
			// Handle .tar.gz case
			configBaseName = strings.TrimSuffix(configBaseName, ".tar.gz")
		} else {
			// Handle other cases by removing extensions one by one
			if strings.HasSuffix(configBaseName, ".gpg") {
				configBaseName = strings.TrimSuffix(configBaseName, ".gpg")
			}
			if strings.HasSuffix(configBaseName, ".gz") {
				configBaseName = strings.TrimSuffix(configBaseName, ".gz")
			}
			if strings.HasSuffix(configBaseName, ".tar") {
				configBaseName = strings.TrimSuffix(configBaseName, ".tar")
			}
		}

		// Create the config file path
		configFilePath := filepath.Join(backupDir, configBaseName+".backup.yaml")

		// Check if the config file exists and delete it
		if _, err := os.Stat(configFilePath); err == nil {
			if err := os.Remove(configFilePath); err != nil {
				fmt.Printf("  Warning: Failed to delete associated config file %s: %v\n", configFilePath, err)
			} else {
				fmt.Printf("  Deleted associated config file: %s\n", configFilePath)
			}
		}

		// Also check for other possible config file names (for backward compatibility or different formats)
		possibleConfigNames := []string{
			configBaseName + ".backup.yaml",        // Standard format
			configBaseName + ".tar.gz.backup.yaml", // Possible format with extension
			configBaseName + ".gpg.backup.yaml",    // Possible format with gpg extension
		}

		for _, possibleName := range possibleConfigNames {
			if possibleName == configBaseName+".backup.yaml" {
				continue // Already checked above
			}

			possiblePath := filepath.Join(backupDir, possibleName)
			if _, err := os.Stat(possiblePath); err == nil {
				if err := os.Remove(possiblePath); err != nil {
					fmt.Printf("  Warning: Failed to delete associated config file %s: %v\n", possiblePath, err)
				} else {
					fmt.Printf("  Deleted associated config file: %s\n", possiblePath)
				}
			}
		}
	}

	return nil
}
