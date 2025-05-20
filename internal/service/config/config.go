// Package config provides functionality for reading and parsing the backup configuration
package config

import (
	"os"

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
