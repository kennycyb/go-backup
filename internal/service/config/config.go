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

// BackupStatus represents the status of the last backup run
type BackupStatus struct {
	Timestamp time.Time `yaml:"timestamp"`
	Status    string    `yaml:"status"` // "Success" or "Failure"
	Message   string    `yaml:"message,omitempty"`
}

// BackupTarget represents a target destination for backups
type BackupTarget struct {
	Path       string         `yaml:"path,omitempty"`
	File       string         `yaml:"file,omitempty"`
	MaxBackups int            `yaml:"maxBackups,omitempty"`
	Backups    []BackupRecord `yaml:"backups,omitempty"`
	LastRun    *BackupStatus  `yaml:"lastRun,omitempty"`
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

// GlobalBackupEntry represents a single backup location tracked in the global registry
type GlobalBackupEntry struct {
	Location string    `yaml:"location"` // Full path to the directory containing .backup.yaml
	RunAt    time.Time `yaml:"run_at"`   // Last run timestamp
}

// GlobalBackupRegistry represents the structure of ~/.backup.yaml global config
type GlobalBackupRegistry struct {
	Default struct {
		Encryption *EncryptionConfig `yaml:"encryption,omitempty"`
	} `yaml:"default,omitempty"`
	Backups []GlobalBackupEntry `yaml:"backups,omitempty"`
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

// IsFileTarget returns true if this target is a single file backup (no rotation)
func (t BackupTarget) IsFileTarget() bool {
	return t.File != ""
}

// GetDestination returns the destination path for this target
func (t BackupTarget) GetDestination() string {
	if t.IsFileTarget() {
		return t.File
	}
	return t.Path
}

// AddBackupRecord adds a new backup record to the specified target in the config
func AddBackupRecord(config *BackupConfig, targetPath string, record BackupRecord) {
	// Find the target index
	targetIndex := -1
	for i, target := range config.Targets {
		if target.GetDestination() == targetPath {
			targetIndex = i
			break
		}
	}

	// If target found, add the backup record
	if targetIndex >= 0 {
		// For file targets, only keep the most recent backup record
		if config.Targets[targetIndex].IsFileTarget() {
			config.Targets[targetIndex].Backups = []BackupRecord{record}
		} else {
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
		if t.GetDestination() == targetPath {
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
		if t.GetDestination() == target.GetDestination() {
			return false // Already exists
		}
	}
	config.Targets = append(config.Targets, target)
	return true
}

// UpdateTargetStatus updates the last run status for a specific target
func UpdateTargetStatus(config *BackupConfig, targetPath string, status string, message string) {
	for i, target := range config.Targets {
		if target.GetDestination() == targetPath {
			config.Targets[i].LastRun = &BackupStatus{
				Timestamp: time.Now(),
				Status:    status,
				Message:   message,
			}
			break
		}
	}
}

// UpdateGlobalRegistry updates the global ~/.backup.yaml file to track backup locations
// If the file doesn't exist, this function returns nil without creating it
func UpdateGlobalRegistry(localConfigDir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	globalConfigPath := filepath.Join(homeDir, ".backup.yaml")

	// Check if global config exists
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		// Global config doesn't exist, silently return
		return nil
	}

	// Read existing global config
	data, err := os.ReadFile(globalConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read global config: %w", err)
	}

	var registry GlobalBackupRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to parse global config: %w", err)
	}

	// Get absolute path of the local config directory
	absPath, err := filepath.Abs(localConfigDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Update or add entry for this backup location
	now := time.Now()
	found := false
	for i := range registry.Backups {
		if registry.Backups[i].Location == absPath {
			registry.Backups[i].RunAt = now
			found = true
			break
		}
	}

	if !found {
		// Add new entry
		registry.Backups = append(registry.Backups, GlobalBackupEntry{
			Location: absPath,
			RunAt:    now,
		})
	}

	// Write updated config
	updatedData, err := yaml.Marshal(&registry)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	// Add header comment
	header := "# Global backup registry\n# Tracks all backup locations and their last run times\n"
	finalData := []byte(header)
	finalData = append(finalData, updatedData...)

	if err := os.WriteFile(globalConfigPath, finalData, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	return nil
}

// ReadGlobalRegistry reads the global backup registry from ~/.backup.yaml
func ReadGlobalRegistry() (*GlobalBackupRegistry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	globalConfigPath := filepath.Join(homeDir, ".backup.yaml")

	// Check if global config exists
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("global config file ~/.backup.yaml does not exist")
	}

	// Read global config
	data, err := os.ReadFile(globalConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var registry GlobalBackupRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	return &registry, nil
}
