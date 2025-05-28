package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	detailed    bool
	listPath    string
	listAll     bool
	showHistory bool
)

// Backup represents a backup file with metadata
type Backup struct {
	Name      string
	Path      string
	Size      int64
	CreatedAt time.Time
	Source    string
	Timestamp string
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long: `List all available backups with their metadata.
This command will display information about existing backups.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Color and emoji constants (reuse from status.go if available)
		const (
			ColorReset  = "\033[0m"
			ColorRed    = "\033[31m"
			ColorGreen  = "\033[32m"
			ColorYellow = "\033[33m"
			ColorBlue   = "\033[34m"
			ColorCyan   = "\033[36m"
			ColorWhite  = "\033[37m"
			ColorBold   = "\033[1m"
			ColorDim    = "\033[2m"
		)

		fmt.Printf("%s%s\n==============================\n   📦  Backup List           \n==============================%s\n", ColorCyan, ColorBold, ColorReset)

		// Handle history mode separately
		if showHistory {
			listBackupHistory()
			return
		}

		// Get current directory name for filtering
		currentDir := ""
		if !listAll {
			// Get the current directory
			workDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Warning: Could not get current directory: %v\n", err)
				fmt.Println("Using default prefix: go-backup")
				currentDir = "go-backup"
			} else {
				// Extract the base name
				currentDir = filepath.Base(workDir)
				if currentDir == "." || currentDir == "/" {
					currentDir = "go-backup"
				}
			}
			fmt.Printf("%sFiltering backups for source:%s %s\n", ColorDim, ColorReset, currentDir)
		}

		// Determine backup locations to scan
		backupLocations := []string{}

		// If path flag is provided, use it as the only location
		if listPath != "" {
			backupLocations = append(backupLocations, listPath)
		} else {
			// Read from config file
			configPath := ".backup.yaml"
			config, err := configService.ReadBackupConfig(configPath)
			if err != nil {
				fmt.Printf("Warning: Could not read config file: %v\n", err)
				fmt.Println("Using default backup location: .backups/")
				backupLocations = append(backupLocations, ".backups/")
			} else {
				// Add all target paths from config
				for _, target := range config.Targets {
					backupLocations = append(backupLocations, target.Path)
				}

				// If no targets defined, use default
				if len(backupLocations) == 0 {
					fmt.Println("No backup locations found in config. Using default: .backups/")
					backupLocations = append(backupLocations, ".backups/")
				}
			}
		}

		// List backups in all locations
		locationGroups := make(map[string][]Backup)

		fmt.Printf("\n%s%sScanning backup locations:%s\n", ColorCyan, ColorBold, ColorReset)
		for _, location := range backupLocations {
			fmt.Printf("%s→ %s%s\n", ColorBlue, location, ColorReset)
			// Check if location exists
			if _, err := os.Stat(location); os.IsNotExist(err) {
				fmt.Printf("  %s⚠️  Directory does not exist, skipping%s\n", ColorYellow, ColorReset)
				continue
			}

			// Get backups in this location
			backups, err := findBackupsInLocation(location, currentDir)
			if err != nil {
				fmt.Printf("  Error reading backups: %v\n", err)
				continue
			}

			// Store backups by location
			locationGroups[location] = backups
			fmt.Printf("  %sFound %d backups%s\n", ColorDim, len(backups), ColorReset)
		}

		// Check if we found any backups
		totalBackups := 0
		for _, backups := range locationGroups {
			totalBackups += len(backups)
		}

		if totalBackups == 0 {
			if listAll {
				fmt.Printf("\n%s%sNo backups found.%s\n", ColorYellow, ColorBold, ColorReset)
			} else {
				fmt.Printf("\n%s%sNo backups found for source '%s'.%s\n", ColorYellow, ColorBold, currentDir, ColorReset)
				fmt.Printf("%sUse --all flag to list all backups regardless of source.%s\n", ColorDim, ColorReset)
			}
			return
		}

		if listAll {
			fmt.Printf("\n%sFound %d backups across %d locations:%s\n", ColorGreen, totalBackups, len(locationGroups), ColorReset)
		} else {
			fmt.Printf("\n%sFound %d backups for source '%s' across %d locations:%s\n", ColorGreen, totalBackups, currentDir, len(locationGroups), ColorReset)
		}

		// Display backups by location
		for location, backups := range locationGroups {
			fmt.Printf("\n%s📁 Location:%s %s\n", ColorBlue, ColorReset, location)

			// Sort backups by creation time (newest first)
			sort.Slice(backups, func(i, j int) bool {
				return backups[i].CreatedAt.After(backups[j].CreatedAt)
			})

			// Group by source within this location
			sourceGroups := make(map[string][]Backup)
			for _, backup := range backups {
				sourceGroups[backup.Source] = append(sourceGroups[backup.Source], backup)
			}

			// Display each source group
			for source, sourceBackups := range sourceGroups {
				fmt.Printf("  %s📦 Source:%s %s (%d backups)\n", ColorCyan, ColorReset, source, len(sourceBackups))
				for i, backup := range sourceBackups {
					// Only show top 5 backups per source unless detailed is enabled
					if !detailed && i >= 5 {
						fmt.Printf("    %s... and %d more (use --detailed to see all)%s\n", ColorDim, len(sourceBackups)-5, ColorReset)
						break
					}

					// Format file size for human readability
					sizeStr := formatSize(backup.Size)

					if detailed {
						// Detailed view
						fmt.Printf("    %s•%s %s\n", ColorDim, ColorReset, backup.Name)
						fmt.Printf("      %sSize:%s %s\n", ColorDim, ColorReset, sizeStr)
						fmt.Printf("      %sCreated:%s %s\n", ColorDim, ColorReset, backup.CreatedAt.Format("2006-01-02 15:04:05"))
						fmt.Println()
					} else {
						// Simple view
						timeAgo := formatTimeAgo(backup.CreatedAt)
						fmt.Printf("    %s•%s %s %s(%s, %s ago)%s\n", ColorGreen, ColorReset, backup.Name, ColorDim, sizeStr, timeAgo, ColorReset)
					}
				}
			}
		}
	},
}

// findBackupsInLocation scans a directory for backup files
func findBackupsInLocation(dir string, filterPrefix string) ([]Backup, error) {
	backups := []Backup{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".tar.gz") {
			continue // Skip non-backup files
		}

		// If filtering is enabled, skip files that don't match the current directory prefix
		if filterPrefix != "" && !listAll && !strings.HasPrefix(fileName, filterPrefix+"-") {
			continue
		}

		// Get file info
		info, err := file.Info()
		if err != nil {
			fmt.Printf("Warning: Could not get info for %s: %v\n", fileName, err)
			continue
		}

		// Parse file name to extract source and timestamp
		parts := strings.Split(strings.TrimSuffix(fileName, ".tar.gz"), "-")
		if len(parts) < 3 {
			// Not a valid backup file name format, skip
			continue
		}

		// The format is source-date-time.tar.gz
		// Last two parts make up the timestamp
		sourceNameParts := parts[:len(parts)-2]
		sourceName := strings.Join(sourceNameParts, "-")
		timestampStr := fmt.Sprintf("%s-%s", parts[len(parts)-2], parts[len(parts)-1])

		// Parse timestamp
		timestamp, _ := time.Parse("20060102-150405", timestampStr)

		// Create backup info
		backup := Backup{
			Name:      fileName,
			Path:      filepath.Join(dir, fileName),
			Size:      info.Size(),
			CreatedAt: info.ModTime(), // Use file modification time for sorting
			Source:    sourceName,
			Timestamp: timestampStr,
		}

		// If we successfully parsed the timestamp, use it instead of file mod time
		if !timestamp.IsZero() {
			backup.CreatedAt = timestamp
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// formatTimeAgo returns a human-readable string representing time since the given time
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	days := hours / 24
	years := days / 365
	months := days / 30

	switch {
	case years > 0:
		return fmt.Sprintf("%dy", years)
	case months > 0:
		return fmt.Sprintf("%dm", months)
	case days > 0:
		return fmt.Sprintf("%dd", days)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return "just now"
	}
}

// listBackupHistory displays the backup history from the config file
func listBackupHistory() {
	// Read from config file
	configPath := ".backup.yaml"
	if configFile != "" { // Use global configFile var if set
		configPath = configFile
	}

	config, err := configService.ReadBackupConfig(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	// Check if any targets have backup history
	hasHistory := false
	for _, target := range config.Targets {
		if len(target.Backups) > 0 {
			hasHistory = true
			break
		}
	}

	if !hasHistory {
		fmt.Println("No backup history found in config file.")
		fmt.Println("History is recorded when backups are created with the config file specified.")
		return
	}

	fmt.Println("\nBackup History from Config File:")

	// Display backups by target
	for _, target := range config.Targets {
		if len(target.Backups) == 0 {
			continue
		}

		fmt.Printf("\n📁 Location: %s\n", target.Path)

		// Group backups by source
		sourceGroups := make(map[string][]configService.BackupRecord)
		for _, backup := range target.Backups {
			sourceGroups[backup.Source] = append(sourceGroups[backup.Source], backup)
		}

		// Display each source group
		for source, sourceBackups := range sourceGroups {
			fmt.Printf("  📦 Source: %s (%d backups)\n", source, len(sourceBackups))

			// Sort backups by creation time (newest first)
			sort.Slice(sourceBackups, func(i, j int) bool {
				return sourceBackups[i].CreatedAt.After(sourceBackups[j].CreatedAt)
			})

			for i, backup := range sourceBackups {
				// Only show top 5 backups per source unless detailed is enabled
				if !detailed && i >= 5 {
					fmt.Printf("    ... and %d more (use --detailed to see all)\n", len(sourceBackups)-5)
					break
				}

				// Format file size for human readability
				sizeStr := formatSize(backup.Size)

				if detailed {
					// Detailed view
					fmt.Printf("    • %s\n", backup.Filename)
					fmt.Printf("      Size: %s\n", sizeStr)
					fmt.Printf("      Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Println()
				} else {
					// Simple view
					timeAgo := formatTimeAgo(backup.CreatedAt)
					fmt.Printf("    • %s (%s, %s ago)\n", backup.Filename, sizeStr, timeAgo)
				}
			}
		}
	}
}

func init() {
	// Local flags for the list command
	listCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed information")
	listCmd.Flags().StringVarP(&listPath, "path", "p", "", "Custom path to search for backups")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "List all backups, not just those from current directory")
	listCmd.Flags().BoolVar(&showHistory, "history", false, "Show backup history from config file instead of scanning directories")

	// Add command to root
	rootCmd.AddCommand(listCmd)
}
