// Package config provides functionality for reading and parsing the backup configuration
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	Method     string `yaml:"method"`
	Receiver   string `yaml:"receiver"`
	Passphrase string `yaml:"passphrase,omitempty"`
}

// BackupConfig represents the structure of the backup configuration file
type BackupConfig struct {
	Excludes   []string          `yaml:"excludes"`
	Targets    []BackupTarget    `yaml:"target"`
	Encryption *EncryptionConfig `yaml:"encryption,omitempty"`
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
	yamlData := []byte("# Backup configuration file\n# WARNING: Do not manually edit this file unless you know what you're doing\n")
	yamlData = append(yamlData, []byte("# Created/updated by go-backup on: "+time.Now().Format("2006-01-02 15:04:05")+"\n")...)
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

// EnableEncryption sets up GPG encryption in the config file
func EnableEncryption(config *BackupConfig, receiver string) (string, error) {
	if receiver == "" {
		return "", fmt.Errorf("GPG receiver email must be specified when enabling encryption")
	}
	valid, keyInfo, err := ValidateGPGReceiver(receiver)
	if err != nil {
		return "", fmt.Errorf("error validating GPG key: %w", err)
	}
	if !valid {
		return "", fmt.Errorf("invalid GPG recipient '%s'. Please ensure the key is in your keyring", receiver)
	}
	if config.Encryption == nil {
		config.Encryption = &EncryptionConfig{}
	}
	config.Encryption.Method = "gpg"
	config.Encryption.Receiver = receiver
	return keyInfo, nil
}

// DisableEncryption removes encryption from the config
func DisableEncryption(config *BackupConfig) bool {
	if config.Encryption != nil {
		config.Encryption = nil
		return true
	}
	return false
}

// ValidateGPGReceiver checks if the specified GPG recipient exists in the keyring
func ValidateGPGReceiver(recipient string) (bool, string, error) {
	cmd := exec.Command("gpg", "--list-keys", recipient)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "No public key") {
			return false, "", nil
		}
		return false, "", fmt.Errorf("error checking GPG key: %w", err)
	}
	return true, strings.TrimSpace(string(output)), nil
}

// DeleteTarget removes a backup target by its path. Returns true if deleted, false if not found.
func DeleteTarget(config *BackupConfig, targetPath string) bool {
	idx := -1
	for i, t := range config.Targets {
		if t.Path == targetPath {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	config.Targets = append(config.Targets[:idx], config.Targets[idx+1:]...)
	return true
}

// AddTarget adds a new backup target to the config if it does not already exist.
func AddTarget(config *BackupConfig, target BackupTarget) bool {
	for _, t := range config.Targets {
		if t.Path == target.Path {
			return false // Already exists
		}
	}
	config.Targets = append(config.Targets, target)
	return true
}
