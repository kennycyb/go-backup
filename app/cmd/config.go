package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	enableEncryption  bool
	disableEncryption bool
	gpgReceiver       string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure backup settings",
	Long: `Configure backup settings in your .backup.yaml file.
This command allows you to modify various settings in your backup configuration
without having to manually edit the YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := ".backup.yaml"
		if cfgFile != "" {
			configFile = cfgFile
		}

		// Check if config file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("Error: Configuration file '%s' does not exist.\n", configFile)
			fmt.Printf("Run 'go-backup init' to create a new configuration file first.\n")
			return
		}

		// Read the existing configuration
		config, err := configService.ReadBackupConfig(configFile)
		if err != nil {
			fmt.Printf("Error reading configuration file: %v\n", err)
			return
		}

		// Track if any changes were made
		configChanged := false

		// Handle encryption settings
		if enableEncryption && disableEncryption {
			fmt.Println("Error: Cannot both enable and disable encryption at the same time.")
			return
		}

		if enableEncryption {
			if gpgReceiver == "" {
				fmt.Println("Error: GPG receiver email must be specified when enabling encryption.")
				fmt.Println("Use --gpg-receiver flag to specify the recipient email for GPG encryption.")
				return
			}

			// Validate the GPG recipient
			valid, keyInfo, err := validateGPGReceiver(gpgReceiver)
			if err != nil {
				fmt.Printf("Error validating GPG key: %v\n", err)
				return
			}
			if !valid {
				fmt.Printf("Error: Invalid GPG recipient '%s'. Please ensure the key is in your keyring.\n", gpgReceiver)
				return
			}

			fmt.Printf("Found GPG key for recipient: %s\n", keyInfo)

			// Create or update encryption configuration
			if config.Encryption == nil {
				config.Encryption = &configService.EncryptionConfig{}
			}
			config.Encryption.Method = "gpg"
			config.Encryption.Receiver = gpgReceiver

			fmt.Printf("Encryption enabled with GPG for recipient: %s\n", gpgReceiver)
			configChanged = true
		}

		if disableEncryption {
			if config.Encryption != nil {
				config.Encryption = nil
				fmt.Println("Encryption disabled.")
				configChanged = true
			} else {
				fmt.Println("Encryption was not enabled, no change made.")
			}
		}

		// Save configuration if changes were made
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

// validateGPGReceiver checks if the specified GPG recipient exists in the keyring.
// It returns a boolean indicating if the key is valid, the key information, and an error if the check failed.
func validateGPGReceiver(recipient string) (bool, string, error) {
	// Run gpg command to list keys matching the recipient
	cmd := exec.Command("gpg", "--list-keys", recipient)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "No public key") {
			return false, "", nil
		}
		return false, "", fmt.Errorf("error checking GPG key: %w", err)
	}

	// If we got output, the key exists
	return true, strings.TrimSpace(string(output)), nil
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)

	// Encryption flags
	configCmd.Flags().BoolVar(&enableEncryption, "enable-encryption", false, "Enable encryption for backups")
	configCmd.Flags().BoolVar(&disableEncryption, "disable-encryption", false, "Disable encryption for backups")
	configCmd.Flags().StringVar(&gpgReceiver, "gpg-receiver", "", "GPG recipient email for encryption")
}
