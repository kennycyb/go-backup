package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configOverwrite is a flag that determines whether to overwrite existing configuration files
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
		// Prevents accidental overwriting of existing configurations
		if _, err := os.Stat(configFile); err == nil && !configOverwrite {
			fmt.Printf("⚠️ Warning: Configuration file '%s' already exists.\n", configFile)
			fmt.Printf("To create a new config file and overwrite the existing one, use the --overwrite flag.\n")
			fmt.Printf("Example: go-backup init --overwrite\n")
			fmt.Printf("Your existing backup configuration has been preserved.\n")
			return
		}

		// Try to load encryption and target defaults from ~/.backup.yaml
		// This allows users to define global defaults for new configurations
		var encryptionDefault *configService.EncryptionConfig
		var autoTargets []configService.BackupTarget
		usr, err := user.Current()
		if err == nil {
			homeConfig := filepath.Join(usr.HomeDir, ".backup.yaml")
			if f, err := os.Open(homeConfig); err == nil {
				defer f.Close()
				var raw map[string]interface{}
				if err := yaml.NewDecoder(f).Decode(&raw); err == nil {
					// Look for default configuration section in home config
					if def, ok := raw["default"].(map[string]interface{}); ok {
						// Parse encryption default configuration
						if enc, ok := def["encryption"].(map[string]interface{}); ok {
							method, mok := enc["method"].(string)
							receiver, rok := enc["receiver"].(string)
							if mok && rok {
								encryptionDefault = &configService.EncryptionConfig{
									Method:   method,
									Receiver: receiver,
								}
							}
						}
						// Parse target mapping default configuration
						// This creates auto-targets based on directory structure relative to defined base paths
						if tgt, ok := def["target"].([]interface{}); ok {
							cwd, _ := os.Getwd()
							parentDir := filepath.Dir(cwd)
							for _, baseEntry := range tgt {
								baseMap, ok := baseEntry.(map[string]interface{})
								if !ok {
									continue
								}
								base, ok := baseMap["base"].(string)
								targets, ok2 := baseMap["targets"].([]interface{})
								if !ok || !ok2 {
									continue
								}
								// Check if current directory is within the base path
								if rel, err := filepath.Rel(base, parentDir); err == nil && (rel == "." || !strings.HasPrefix(rel, "..")) {
									// Create backup targets based on the relative path
									for _, t := range targets {
										tgtBase, ok := t.(string)
										if !ok {
											continue
										}
										tgtPath := tgtBase
										if rel != "." {
											tgtPath = filepath.Join(tgtBase, rel)
										}
										autoTargets = append(autoTargets, configService.BackupTarget{Path: tgtPath, MaxBackups: 7})
									}
								}
							}
						}
					}
				}
			}
		}

		// Create default configuration with common excludes
		config := configService.BackupConfig{
			Excludes: []string{"node_modules", "bin"},
			Targets:  []configService.BackupTarget{},
		}

		// Use auto-detected targets if available, otherwise provide a default target
		if len(autoTargets) > 0 {
			config.Targets = append(config.Targets, autoTargets...)
		} else {
			config.Targets = append(config.Targets, configService.BackupTarget{
				Path:       ".backups/location1",
				MaxBackups: 7,
			})
		}

		// Set encryption configuration from defaults or provide example values
		if encryptionDefault != nil {
			config.Encryption = encryptionDefault
		} else {
			// Provide example encryption configuration
			// Note: Users should update the receiver email address
			config.Encryption = &configService.EncryptionConfig{
				Method:     "gpg",
				Receiver:   "user@example.com",
				Passphrase: "",
			}
		}

		// Write the config to file
		err = configService.WriteBackupConfig(configFile, &config)
		if err != nil {
			fmt.Printf("Error writing configuration file: %v\n", err)
			return
		}

		// Success message with guidance for next steps
		fmt.Printf("Configuration file '%s' created successfully.\n", configFile)
		fmt.Println("Edit this file to customize your backup targets and settings.")
	},
}

func init() {
	// Register command line flags for the init command
	initCmd.Flags().BoolVar(&configOverwrite, "overwrite", false, "Overwrite existing configuration file if it exists")

	// Register the init command with the root command
	rootCmd.AddCommand(initCmd)
}
