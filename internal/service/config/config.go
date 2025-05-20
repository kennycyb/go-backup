// Package config provides functionality for reading and parsing the backup configuration
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BackupConfig represents the structure of the backup configuration file
type BackupConfig struct {
	Excludes []string `yaml:"excludes"`
	Targets  []struct {
		Path       string `yaml:"path"`
		MaxBackups int    `yaml:"maxBackups"`
	} `yaml:"target"`
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
