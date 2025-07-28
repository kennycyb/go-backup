package cmd

import (
	"fmt"
	"os"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

// Command-line flags for configuration management
var (
	enableEncryption  bool   // Flag to enable GPG encryption for backups
	disableEncryption bool   // Flag to disable encryption for backups
	gpgReceiver       string // GPG recipient email address for encryption
	deleteTarget      string // Target path to remove from backup configuration
	addTarget         string // Target path to add to backup configuration
)

// configCmd represents the config command for managing backup settings
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure backup settings",
	Long: `Configure backup settings in your .backup.yaml file.
This command allows you to modify various settings in your backup configuration
without having to manually edit the YAML file.

Examples:
  go-backup config --add-target /path/to/directory
  go-backup config --delete-target /path/to/directory
  go-backup config --enable-encryption --gpg-receiver user@example.com
  go-backup config --disable-encryption`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine configuration file path - use custom path if provided, otherwise default
		configFile := ".backup.yaml"
		if cfgFile != "" {
			configFile = cfgFile
		}

		// Verify configuration file exists before attempting to modify it
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("Error: Configuration file '%s' does not exist.\n", configFile)
			fmt.Printf("Run 'go-backup init' to create a new configuration file first.\n")
			return
		}

		// Read existing configuration from file
		config, err := configService.ReadBackupConfig(configFile)
		if err != nil {
			fmt.Printf("Error reading configuration file: %v\n", err)
			return
		}

		// Initialize variable to track if any configuration changes are made
		configChanged := false

		// Handle adding new backup targets
		if addTarget != "" {
			target := configService.BackupTarget{Path: addTarget}
			if configService.AddTarget(config, target) {
				fmt.Printf("Target '%s' added to configuration.\n", addTarget)
				configChanged = true
			} else {
				fmt.Printf("Target '%s' already exists in configuration.\n", addTarget)
			}
		}

		// Handle removing existing backup targets
		if deleteTarget != "" {
			if configService.DeleteTarget(config, deleteTarget) {
				fmt.Printf("Target '%s' deleted from configuration.\n", deleteTarget)
				configChanged = true
			} else {
				fmt.Printf("Target '%s' not found in configuration.\n", deleteTarget)
			}
		}

		// Validate that both encryption flags are not used simultaneously
		if enableEncryption && disableEncryption {
			fmt.Println("Error: Cannot both enable and disable encryption at the same time.")
			return
		}

		// Handle enabling GPG encryption
		if enableEncryption {
			keyInfo, err := configService.EnableEncryption(config, gpgReceiver)
			if err != nil {
				fmt.Printf("Error enabling encryption: %v\n", err)
				return
			}
			fmt.Printf("Found GPG key for recipient: %s\n", keyInfo)
			fmt.Printf("Encryption enabled with GPG for recipient: %s\n", gpgReceiver)
			configChanged = true
		}

		// Handle disabling encryption
		if disableEncryption {
			if configService.DisableEncryption(config) {
				fmt.Println("Encryption disabled.")
				configChanged = true
			} else {
				fmt.Println("Encryption was not enabled, no change made.")
			}
		}

		// Write updated configuration to file only if changes were made
		if configChanged {
			err := configService.WriteBackupConfig(configFile, config)
			if err != nil {
				fmt.Printf("Error writing configuration file: %v\n", err)
				return
			}
			fmt.Printf("Configuration file '%s' updated successfully.\n", configFile)
		} else {
			fmt.Println("No changes were made to the configuration.")
		}
	},
}

// init initializes the config command with its flags and adds it to the root command
func init() {
	// Add config command to root command tree
	rootCmd.AddCommand(configCmd)

	// Define encryption-related flags
	configCmd.Flags().BoolVar(&enableEncryption, "enable-encryption", false, "Enable encryption for backups")
	configCmd.Flags().BoolVar(&disableEncryption, "disable-encryption", false, "Disable encryption for backups")
	configCmd.Flags().StringVar(&gpgReceiver, "gpg-receiver", "", "GPG recipient email for encryption")

	// Define target management flags
	configCmd.Flags().StringVar(&deleteTarget, "delete-target", "", "Delete a target from the configuration")
	configCmd.Flags().StringVar(&addTarget, "add-target", "", "Add a new backup target to the configuration")
}
