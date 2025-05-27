// Package config provides functionality for reading and parsing the backup configuration
package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// BackupRecord represents an individual backup entry
type BackupRecord struct {
	Filename  string    `yaml:"filename"`
	Source    string    `yaml:"source"`
	CreatedAt time.Time `yaml:"createdAt"`
	Size      int64     `yaml:"size"`
}

// BackupTarget represents a target destination for backups
type BackupTarget struct {
	Path       string         `yaml:"path"`
	MaxBackups int            `yaml:"maxBackups"`
	Backups    []BackupRecord `yaml:"backups,omitempty"`
}

// EncryptionConfig represents the encryption configuration
type EncryptionConfig struct {
	Method   string `yaml:"method"`
	Receiver string `yaml:"receiver"`
}

// BackupConfig represents the structure of the backup configuration file
type BackupConfig struct {
	Excludes   []string           `yaml:"excludes"`
	Targets    []BackupTarget     `yaml:"target"`
	Encryption []EncryptionConfig `yaml:"encryption,omitempty"`
}

// ReadBackupConfig reads the backup configuration from the specified file
func ReadBackupConfig(filePath string) (*BackupConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config BackupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set default values for any targets with unspecified/zero maxBackups
	for i := range config.Targets {
		if config.Targets[i].MaxBackups <= 0 {
			config.Targets[i].MaxBackups = 7 // Default value
		}
	}

	return &config, nil
}

// WriteBackupConfig writes the backup configuration to the specified file
func WriteBackupConfig(filePath string, config *BackupConfig) error {
	// Create the directory for the output path if it doesn't exist
	outputDir := filepath.Dir(filePath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}

	// Marshal the config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Add comment at the top of the YAML file
	yamlData := []byte("# Backup configuration file\n")
	yamlData = append(yamlData, data...)

	// Write the config to file
	return os.WriteFile(filePath, yamlData, 0644)
}

// AddBackupRecord adds a new backup record to the specified target in the config
func AddBackupRecord(config *BackupConfig, targetPath string, record BackupRecord) {
	// Find the target index
	targetIndex := -1
	for i, target := range config.Targets {
		if target.Path == targetPath {
			targetIndex = i
			break
		}
	}

	// If target found, add the backup record
	if targetIndex >= 0 {
		// Add the new backup to the beginning of the list for the target
		config.Targets[targetIndex].Backups = append(
			[]BackupRecord{record},
			config.Targets[targetIndex].Backups...,
		)

		// Ensure we have a valid maxBackups value
		maxBackups := config.Targets[targetIndex].MaxBackups
		if maxBackups <= 0 {
			maxBackups = 7 // Default value
			config.Targets[targetIndex].MaxBackups = maxBackups
		}

		// Trim the list to match the maxBackups value if needed
		if len(config.Targets[targetIndex].Backups) > maxBackups {
			config.Targets[targetIndex].Backups = config.Targets[targetIndex].Backups[:maxBackups]
		}
	}
}
