package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	backupService "github.com/kennycyb/go-backup/internal/service/backup"
	compressionService "github.com/kennycyb/go-backup/internal/service/compress"
	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	source      string
	destination string
	compress    bool
	configFile  string
	excludeDirs []string
)

// runCmd represents the run command (previously backup command)
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Create a new backup",
	Long: `Create a new backup of specified files or directories.
This command will package and compress the specified sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating backup...")

		// If source is empty, use current directory
		if source == "" {
			sourceDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			source = sourceDir
		}

		// Create a timestamp for the backup file
		timestamp := time.Now().Format("20060102-150405")

		// Get the current folder name for the backup file prefix
		currentDir := filepath.Base(source)
		if currentDir == "." || currentDir == "/" {
			// If source is the root directory or current directory symbol,
			// use "go-backup" as the default name
			currentDir = "go-backup"
		}

		backupFileName := fmt.Sprintf("%s-%s.tar.gz", currentDir, timestamp)
		tempBackupPath := filepath.Join(os.TempDir(), backupFileName)

		fmt.Printf("Source: %s\n", source)
		fmt.Printf("Backup name: %s\n", backupFileName)
		fmt.Printf("Temporary backup file: %s\n", tempBackupPath)

		// Get excludes from config file
		configExcludes := []string{} // Default empty list
		var config *configService.BackupConfig

		// Read config file for excludes
		configPath := ".backup.yaml"
		if configFile != "" {
			configPath = configFile
		}

		var configErr error
		config, configErr = configService.ReadBackupConfig(configPath)
		if configErr == nil && len(config.Excludes) > 0 {
			configExcludes = config.Excludes
			fmt.Printf("Using excludes from config: %v\n", configExcludes)
		} else {
			// If no config excludes, use the command line ones
			configExcludes = excludeDirs
			fmt.Printf("Using default excludes: %v\n", configExcludes)
		}

		// Create the tar.gz archive using the compression service
		err := compressionService.CreateTarGzArchive(source, tempBackupPath, configExcludes)
		if err != nil {
			fmt.Printf("Error creating backup archive: %v\n", err)
			os.Exit(1)
		}

		// Determine destinations from config or command line argument
		destinations := []string{}

		if destination != "" {
			// If destination is specified via command line
			destinations = append(destinations, destination)
		} else {
			// Use the config we already loaded or read it again
			if configErr != nil {
				fmt.Printf("Error reading config file %s: %v\n", configPath, configErr)
				fmt.Println("Using default backup location: .backups/")

				// Use a default destination
				destinations = append(destinations, ".backups/")
			} else {
				// Use destinations from config
				for _, target := range config.Targets {
					destinations = append(destinations, target.Path)
				}
			}
		}

		// Copy backup file to all destinations
		fmt.Println("\nProcessing backup destinations:")
		for _, dest := range destinations {
			fmt.Printf("\n→ Destination: %s\n", dest)

			// Check if destination directory exists
			if _, err := os.Stat(dest); os.IsNotExist(err) {
				fmt.Printf("  Skipping: directory does not exist\n")
				continue
			}

			destFilePath := filepath.Join(dest, backupFileName)
			fmt.Printf("  Copying file: %s\n", filepath.Base(destFilePath))

			if err := copyFile(tempBackupPath, destFilePath); err != nil {
				fmt.Printf("  Error: failed to copy backup - %v\n", err)
			} else {
				fmt.Printf("  Success: backup copied successfully\n")

				// Get maxBackups value from config or use default
				maxBackups := 7 // Default value

				if configFile != "" || destination == "" {
					// Only apply rotation if using config or default destination
					if configErr == nil {
						// Find the target that matches this destination
						for _, target := range config.Targets {
							if target.Path == dest {
								if target.MaxBackups > 0 {
									maxBackups = target.MaxBackups
								}
								break
							}
						}
					}

					// Get the current folder name used as prefix from the source path
					prefixName := filepath.Base(source)
					if prefixName == "." || prefixName == "/" {
						prefixName = "go-backup"
					}
					prefix := prefixName + "-"

					// Cleanup old backups
					if err := backupService.CleanupOldBackups(dest, prefix, maxBackups); err != nil {
						fmt.Printf("  Warning: Failed to cleanup old backups - %v\n", err)
					} else {
						fmt.Printf("  Rotation: Keeping latest %d backups\n", maxBackups)
					}
				}
			}
		}

		// Clean up the temporary file
		os.Remove(tempBackupPath)
		fmt.Println("\nBackup completed successfully!")
	},
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	// Sync the file to ensure it's written to disk
	return dstFile.Sync()
}

func init() {
	// Local flags for the run command
	runCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory to backup (defaults to current directory)")
	runCmd.Flags().StringVarP(&destination, "dest", "d", "", "Destination directory for backup (if not specified, uses config file)")
	runCmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress the backup")
	runCmd.Flags().StringVarP(&configFile, "config", "f", ".backup.yaml", "Config file path")
	runCmd.Flags().StringSliceVar(&excludeDirs, "exclude", []string{".git", "node_modules", "bin"}, "Directories to exclude from backup")

	// Add command to root
	rootCmd.AddCommand(runCmd)
}
