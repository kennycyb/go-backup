package cmd

import (
	"fmt"
	"os"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	configOverwrite bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new backup configuration",
	Long: `Initialize a new backup configuration by creating a .backup.yaml file
in the current directory. This file will define backup targets and settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := ".backup.yaml"

		// Check if config file exists and overwrite flag is not set
		if _, err := os.Stat(configFile); err == nil && !configOverwrite {
			fmt.Printf("Configuration file '%s' already exists. Use --overwrite to replace it.\n", configFile)
			return
		}

		// Create default configuration
		config := configService.BackupConfig{
			Excludes: []string{".git", "node_modules", "bin"},
			Targets: []configService.BackupTarget{
				{
					Path:       ".backups/location1",
					MaxBackups: 7,
				},
			},
			Encryption: []configService.EncryptionConfig{
				{
					Method:     "gpg",
					Receiver:   "user@example.com",
					Passphrase: "", // Passphrase is empty by default for security reasons
				},
			},
		}

		// Write the config to file
		err := configService.WriteBackupConfig(configFile, &config)
		if err != nil {
			fmt.Printf("Error writing configuration file: %v\n", err)
			return
		}

		fmt.Printf("Configuration file '%s' created successfully.\n", configFile)
		fmt.Println("Edit this file to customize your backup targets and settings.")
	},
}

func init() {
	// Local flags for the init command
	initCmd.Flags().BoolVar(&configOverwrite, "overwrite", false, "Overwrite existing configuration file if it exists")

	// Add command to root
	rootCmd.AddCommand(initCmd)
}
