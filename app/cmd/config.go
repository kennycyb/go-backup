package cmd

import (
	"fmt"
	"os"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	enableEncryption  bool
	disableEncryption bool
	gpgReceiver       string
	deleteTarget      string
	addTarget         string
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

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("Error: Configuration file '%s' does not exist.\n", configFile)
			fmt.Printf("Run 'go-backup init' to create a new configuration file first.\n")
			return
		}

		config, err := configService.ReadBackupConfig(configFile)
		if err != nil {
			fmt.Printf("Error reading configuration file: %v\n", err)
			return
		}

		configChanged := false

		if addTarget != "" {
			target := configService.BackupTarget{Path: addTarget}
			if configService.AddTarget(config, target) {
				fmt.Printf("Target '%s' added to configuration.\n", addTarget)
				configChanged = true
			} else {
				fmt.Printf("Target '%s' already exists in configuration.\n", addTarget)
			}
		}

		if deleteTarget != "" {
			if configService.DeleteTarget(config, deleteTarget) {
				fmt.Printf("Target '%s' deleted from configuration.\n", deleteTarget)
				configChanged = true
			} else {
				fmt.Printf("Target '%s' not found in configuration.\n", deleteTarget)
			}
		}

		if enableEncryption && disableEncryption {
			fmt.Println("Error: Cannot both enable and disable encryption at the same time.")
			return
		}

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

		if disableEncryption {
			if configService.DisableEncryption(config) {
				fmt.Println("Encryption disabled.")
				configChanged = true
			} else {
				fmt.Println("Encryption was not enabled, no change made.")
			}
		}

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

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)

	// Encryption flags
	configCmd.Flags().BoolVar(&enableEncryption, "enable-encryption", false, "Enable encryption for backups")
	configCmd.Flags().BoolVar(&disableEncryption, "disable-encryption", false, "Disable encryption for backups")
	configCmd.Flags().StringVar(&gpgReceiver, "gpg-receiver", "", "GPG recipient email for encryption")
	configCmd.Flags().StringVar(&deleteTarget, "delete-target", "", "Delete a target from the configuration")
	configCmd.Flags().StringVar(&addTarget, "add-target", "", "Add a new backup target to the configuration")
}
