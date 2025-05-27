package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	backupService "github.com/kennycyb/go-backup/internal/service/backup"
	compressionService "github.com/kennycyb/go-backup/internal/service/compress"
	configService "github.com/kennycyb/go-backup/internal/service/config"
	encryptionService "github.com/kennycyb/go-backup/internal/service/encrypt"
	"github.com/spf13/cobra"
)

var (
	source      string
	destination string
	compress    bool
	configFile  string
	excludeDirs []string
	encrypt     bool
	encryptTo   string
	copyConfig  bool
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

		// Handle encryption if requested or configured
		useEncryption := encrypt
		encryptionReceiver := encryptTo // If not explicitly set via command line, check config
		if !useEncryption && config != nil && config.Encryption != nil {
			if config.Encryption.Method == "gpg" {
				useEncryption = true
				if encryptionReceiver == "" {
					encryptionReceiver = config.Encryption.Receiver
				}
			}
		}

		// Apply encryption if enabled
		if useEncryption {
			if encryptionReceiver == "" {
				fmt.Println("Error: GPG encryption enabled but no recipient specified")
				fmt.Println("Please specify a recipient using --encrypt-to flag or in the config file")
				os.Exit(1)
			}

			fmt.Printf("Encrypting backup with GPG for recipient: %s\n", encryptionReceiver)
			// Encrypt the temporary backup file
			encryptedPath, err := encryptionService.GPGEncrypt(tempBackupPath, encryptionReceiver)
			if err != nil {
				fmt.Printf("Error encrypting backup: %v\n", err)
				os.Exit(1)
			}

			// Remove the original unencrypted file
			os.Remove(tempBackupPath)

			// Update the temporary backup path to point to the encrypted file
			tempBackupPath = encryptedPath

			// Update the backup filename to include .gpg extension
			backupFileName = backupFileName + ".gpg"
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

			if err := backupService.CopyFile(tempBackupPath, destFilePath); err != nil {
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
								// Always use maxBackups from target, as ReadBackupConfig
								// already sets the default value of 7 if it was empty
								maxBackups = target.MaxBackups
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

					// Record this backup in the config file if we're using a config
					if configErr == nil && configFile != "" {
						// Get file information for size
						fileInfo, err := os.Stat(destFilePath)
						if err == nil {
							// Create a backup record
							backupRecord := configService.BackupRecord{
								Filename:  filepath.Base(destFilePath),
								Source:    source,
								CreatedAt: time.Now(),
								Size:      fileInfo.Size(),
							}

							// Add the record to the config
							configService.AddBackupRecord(config, dest, backupRecord)

							// Save updated config
							if err := configService.WriteBackupConfig(configPath, config); err != nil {
								fmt.Printf("  Warning: Failed to update backup history in config - %v\n", err)
							} else {
								fmt.Printf("  History: Updated backup history in %s\n", configPath)
							}

							// Copy the config file to the destination with backup name prefix if enabled
							if copyConfig {
								configBaseName := filepath.Base(backupFileName)
								configBaseName = strings.TrimSuffix(configBaseName, ".tar.gz") // Remove .tar.gz
								configBaseName = strings.TrimSuffix(configBaseName, ".gpg")    // Remove .gpg if encrypted
								destConfigPath := filepath.Join(dest, configBaseName+".backup.yaml")

								// Get the encryption receiver if encryption was used
								currentEncryptionReceiver := encryptionReceiver

								// Copy the config with added helpful comments
								if err := configService.CopyConfigWithHelp(configPath, destConfigPath, useEncryption, currentEncryptionReceiver); err != nil {
									fmt.Printf("  Warning: Failed to copy config file to destination - %v\n", err)
								} else {
									fmt.Printf("  Config: Copied config file with usage info to %s\n", destConfigPath)
								}
							}
						}
					}
				}
			}
		}

		// Clean up the temporary file
		os.Remove(tempBackupPath)
		fmt.Println("\nBackup completed successfully!")
	},
}

func init() {
	// Local flags for the run command
	runCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory to backup (defaults to current directory)")
	runCmd.Flags().StringVarP(&destination, "dest", "d", "", "Destination directory for backup (if not specified, uses config file)")
	runCmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress the backup")
	runCmd.Flags().StringVarP(&configFile, "config", "f", ".backup.yaml", "Config file path")
	runCmd.Flags().BoolVarP(&encrypt, "encrypt", "e", false, "Encrypt the backup using GPG")
	runCmd.Flags().StringVar(&encryptTo, "encrypt-to", "", "GPG recipient email for encryption (defaults to config value)")
	runCmd.Flags().StringSliceVar(&excludeDirs, "exclude", []string{".git", "node_modules", "bin"}, "Directories to exclude from backup")
	runCmd.Flags().BoolVar(&copyConfig, "copy-config", true, "Copy the config file to the target directories with the same name prefix as the backup")

	// Add command to root
	rootCmd.AddCommand(runCmd)
}
