package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type BackupConfig struct {
	Excludes []string `yaml:"excludes"`
	Targets  []struct {
		Path       string `yaml:"path"`
		MaxBackups int    `yaml:"maxBackups"`
	} `yaml:"target"`
}

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
		config := BackupConfig{
			Targets: []struct {
				Path       string `yaml:"path"`
				MaxBackups int    `yaml:"maxBackups"`
			}{
				{
					Path:       ".backups/location1",
					MaxBackups: 7,
				},
			},
		}

		// Create the directory for the output path if it doesn't exist
		outputDir := filepath.Dir(configFile)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", outputDir, err)
				return
			}
		}

		// Marshal the config to YAML
		data, err := yaml.Marshal(&config)
		if err != nil {
			fmt.Printf("Error marshaling configuration: %v\n", err)
			return
		}

		// Add comment at the top of the YAML file
		yamlData := []byte("# Backup targets\n")
		yamlData = append(yamlData, data...)

		// Write the config to file
		err = os.WriteFile(configFile, yamlData, 0644)
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
}
