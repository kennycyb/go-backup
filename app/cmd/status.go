package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show backup status",
	Long: `Show the status of backups, including the last backup time
and the latest backup files for each target.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := ".backup.yaml"
		if cfgFile != "" {
			configFile = cfgFile
		}

		// Check if config file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("%s%sError:%s Configuration file '%s' does not exist.\n", ColorRed, ColorBold, ColorReset, configFile)
			fmt.Printf("Run 'go-backup init' to create a new configuration file first.\n")
			return
		}

		// Read the existing configuration
		config, err := configService.ReadBackupConfig(configFile)
		if err != nil {
			fmt.Printf("%s%sError reading configuration file:%s %v\n", ColorRed, ColorBold, ColorReset, err)
			return
		}

		// No targets found
		if len(config.Targets) == 0 {
			fmt.Printf("%s%s⚠️  No backup targets defined in configuration.%s\n", ColorYellow, ColorBold, ColorReset)
			return
		}

		// Header
		fmt.Printf("%s%s\n==============================\n   📦  Backup Status Report   \n==============================%s\n", ColorCyan, ColorBold, ColorReset)

		// Show encryption information if configured
		if config.Encryption != nil {
			fmt.Printf("\n%s🔒  Encryption: %sEnabled%s\n", ColorYellow, ColorGreen, ColorReset)
			fmt.Printf("%s  • Method:   %s%s\n", ColorDim, ColorReset, config.Encryption.Method)
			fmt.Printf("%s  • Receiver: %s%s\n", ColorDim, ColorReset, config.Encryption.Receiver)
		} else {
			fmt.Printf("\n%s🔓  Encryption: %sDisabled%s\n", ColorYellow, ColorRed, ColorReset)
		}

		hasAnyBackups := false

		for _, target := range config.Targets {
			fmt.Printf("\n%s%s📁 Target:%s %s%s\n", ColorBlue, ColorBold, ColorReset, ColorWhite, target.Path)
			fmt.Printf("%s  • Maximum backups:%s %d\n", ColorDim, ColorReset, target.MaxBackups)

			if len(target.Backups) == 0 {
				fmt.Printf("%s%s  ⚠️  Status: No backups found%s\n", ColorYellow, ColorBold, ColorReset)
				continue
			}

			hasAnyBackups = true

			// The first backup in the list is the most recent one
			latestBackup := target.Backups[0]
			timeSinceBackup := time.Since(latestBackup.CreatedAt)

			fmt.Printf("%s  • Latest backup:%s %s%s\n", ColorDim, ColorReset, ColorGreen, latestBackup.Filename)
			fmt.Printf("%s  • Source:%s %s\n", ColorDim, ColorReset, latestBackup.Source)
			fmt.Printf("%s  • Created:%s %s (%s ago)\n", ColorDim, ColorReset, latestBackup.CreatedAt.Format("2006-01-02 15:04:05"), formatTimeSince(timeSinceBackup))
			fmt.Printf("%s  • Size:%s %s\n", ColorDim, ColorReset, formatFileSize(latestBackup.Size))

			// Check if the backup file exists
			backupFilePath := filepath.Join(target.Path, latestBackup.Filename)
			if _, err := os.Stat(backupFilePath); os.IsNotExist(err) {
				fmt.Printf("%s%s  ❌  Status: WARNING - Backup file not found on disk!%s\n", ColorRed, ColorBold, ColorReset)
			} else {
				fmt.Printf("%s%s  ✅  Status: OK%s\n", ColorGreen, ColorBold, ColorReset)
			}

			// Show the total number of available backups
			fmt.Printf("%s  • Total backups:%s %d/%d\n", ColorDim, ColorReset, len(target.Backups), target.MaxBackups)
		}

		if !hasAnyBackups {
			fmt.Printf("\n%s%sℹ️  No backups have been created yet.%s\n", ColorCyan, ColorBold, ColorReset)
			fmt.Println("Run 'go-backup run' to create your first backup.")
		}
	},
}

// formatTimeSince formats a duration into a human-readable string
func formatTimeSince(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// formatFileSize formats a file size in bytes into a human-readable string
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func init() {
	// Add status command to root
	rootCmd.AddCommand(statusCmd)
}
