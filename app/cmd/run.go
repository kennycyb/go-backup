// filepath: /workspaces/go-backup/app/cmd/run.go
package cmd

import (
	"bufio"
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
	force       bool
)

// runCmd represents the run command (previously backup command)
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Create a new backup",
	Long: `Create a new backup of specified files or directories.
This command will package and compress the specified sources.`,
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

		fmt.Printf("%s%s\n==============================\n   📦  Starting Backup Job    \n==============================%s\n", ColorCyan, ColorBold, ColorReset)

		// If source is empty, use current directory
		if source == "" {
			sourceDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("%s%s❌ Error getting current directory:%s %v\n", ColorRed, ColorBold, ColorReset, err)
				os.Exit(1)
			}
			source = sourceDir
		}

		// Create a timestamp for the backup file
		timestamp := time.Now().Format("20060102-150405")

		// Get the current folder name for the backup file prefix
		currentDir := filepath.Base(source)
		if currentDir == "." || currentDir == "/" {
			currentDir = "go-backup"
		}

		backupFileName := fmt.Sprintf("%s-%s.tar.gz", currentDir, timestamp)
		tempBackupPath := filepath.Join(os.TempDir(), backupFileName)

		fmt.Printf("%sSource:%s %s\n", ColorDim, ColorReset, source)
		fmt.Printf("%sBackup name:%s %s\n", ColorDim, ColorReset, backupFileName)
		fmt.Printf("%sTemporary backup file:%s %s\n", ColorDim, ColorReset, tempBackupPath)

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
		if configErr != nil {
			fmt.Printf("Error reading config file %s: %v\n", configPath, configErr)
			os.Exit(1)
		}

		if len(config.Excludes) > 0 {
			configExcludes = config.Excludes
			fmt.Printf("%sUsing excludes from config:%s %v\n", ColorDim, ColorReset, configExcludes)
		} else {
			configExcludes = excludeDirs
			fmt.Printf("%sUsing default excludes:%s %v\n", ColorDim, ColorReset, configExcludes)
		}

		// Check for potentially problematic file sizes before creating archive
		fmt.Printf("%sAnalyzing files for potential size issues...%s\n", ColorDim, ColorReset)
		fileSummary, sizeErr := compressionService.CheckFileSizes(source, configExcludes, 8) // 8GB is the standard tar size limit
		if sizeErr != nil {
			fmt.Printf("%s%s⚠️ Warning: Unable to analyze file sizes:%s %v\n", ColorYellow, ColorBold, ColorReset, sizeErr)
		} else if len(fileSummary.FilesOverSize) > 0 {
			fmt.Printf("%s%s⚠️ Warning: %d files exceed the recommended size limit for tar archives:%s\n",
				ColorYellow, ColorBold, len(fileSummary.FilesOverSize), ColorReset)
			for i, file := range fileSummary.FilesOverSize {
				if i < 5 { // Only show the first 5 files
					fmt.Printf("  - %s (%.2f GB)\n", file, float64(fileSummary.LargestFileSize)/(1024*1024*1024))
				} else {
					fmt.Printf("  - ... and %d more\n", len(fileSummary.FilesOverSize)-5)
					break
				}
			}
			fmt.Printf("%sConsider excluding these files or using the --split option for large files%s\n",
				ColorDim, ColorReset)

			// If force flag is not set, ask for confirmation
			if !force {
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("%sContinue with backup anyway? [y/N]:%s ", ColorYellow, ColorReset)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Backup aborted.")
					os.Exit(0)
				}
			}
		}

		// Create the tar.gz archive using the compression service
		err := compressionService.CreateTarGzArchive(source, tempBackupPath, configExcludes)
		if err != nil {
			if strings.Contains(err.Error(), "too large for tar format") {
				fmt.Printf("%s%s❌ Error creating backup archive:%s %v\n", ColorRed, ColorBold, ColorReset, err)
				fmt.Printf("%sSuggestion: Use --exclude to skip large files or consider using a different backup strategy for very large files%s\n",
					ColorYellow, ColorReset)
			} else {
				fmt.Printf("%s%s❌ Error creating backup archive:%s %v\n", ColorRed, ColorBold, ColorReset, err)
			}
			os.Exit(1)
		}

		// Handle encryption if requested or configured
		useEncryption := encrypt
		encryptionReceiver := encryptTo
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
				fmt.Printf("%s%s❌ Error:%s GPG encryption enabled but no recipient specified\n", ColorRed, ColorBold, ColorReset)
				fmt.Println("Please specify a recipient using --encrypt-to flag or in the config file")
				os.Exit(1)
			}

			fmt.Printf("%s🔒 Encrypting backup with GPG for recipient:%s %s\n", ColorYellow, ColorReset, encryptionReceiver)
			// Encrypt the temporary backup file
			encryptedPath, err := encryptionService.GPGEncrypt(tempBackupPath, encryptionReceiver)
			if err != nil {
				fmt.Printf("%s%s❌ Error encrypting backup:%s %v\n", ColorRed, ColorBold, ColorReset, err)
				os.Exit(1)
			}

			os.Remove(tempBackupPath)
			tempBackupPath = encryptedPath
			backupFileName = backupFileName + ".gpg"
		}

		// Determine destinations from config or command line argument
		destinations := []string{}
		if destination != "" {
			destinations = append(destinations, destination)
		} else {
			for _, target := range config.Targets {
				destinations = append(destinations, target.Path)
			}
			if len(destinations) == 0 {
				fmt.Printf("%s%s❌ Error:%s No backup destinations found in config file and no destination specified\n", ColorRed, ColorBold, ColorReset)
				os.Exit(1)
			}
		}

		fmt.Printf("\n%s%sProcessing backup destinations:%s\n", ColorCyan, ColorBold, ColorReset)
		for _, dest := range destinations {
			isFileTarget := false

			// If destination comes from config, try to find the matching target for file/dir info
			var backupFileNameForTarget string = backupFileName
			var destFilePath string

			// Try to match config target for this destination
			var matchedTarget *configService.Target
			for _, t := range config.Targets {
				if t.Path == dest {
					matchedTarget = &t
					break
				}
			}

			if matchedTarget != nil {
				isFileTarget = matchedTarget.IsFileTarget()
			} else {
				// If not found in config, infer: if path exists and is dir, or ends with separator, treat as dir
				info, err := os.Stat(dest)
				if err == nil && info.IsDir() {
					isFileTarget = false
				} else if strings.HasSuffix(dest, string(os.PathSeparator)) {
					isFileTarget = false
				} else {
					isFileTarget = true
				}
			}

			fmt.Printf("\n%s→ Destination:%s %s", ColorBlue, ColorReset, dest)
			if isFileTarget {
				fmt.Printf(" %s(file)%s", ColorDim, ColorReset)
			}
			fmt.Println()
			if !isFileTarget {
				// For directory targets, check if directory exists
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					fmt.Printf("  %s⚠️  Skipping: directory does not exist%s\n", ColorYellow, ColorReset)
					continue
				}
				destFilePath = filepath.Join(dest, backupFileName)
			} else {
				// For file targets, use the file path directly
				// Create directory if it doesn't exist
				destDir := filepath.Dir(dest)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					fmt.Printf("  %s❌ Error: failed to create destination directory -%s %v\n", ColorRed, ColorReset, err)
					continue
				}
				destFilePath = dest
				// For file targets, use the actual filename specified in the target's File field
				backupFileNameForTarget = filepath.Base(dest)
			}

			fmt.Printf("  %sCopying file:%s %s\n", ColorDim, ColorReset, filepath.Base(destFilePath))

			if err := backupService.CopyFile(tempBackupPath, destFilePath); err != nil {
				fmt.Printf("  %s❌ Error: failed to copy backup -%s %v\n", ColorRed, ColorReset, err)
			} else {
				fmt.Printf("  %s✅ Success:%s backup copied successfully\n", ColorGreen, ColorReset)

				// Get maxBackups value from config or use default
				maxBackups := 7 // Default value

				if configFile != "" || destination == "" {
					// Only apply rotation if using config or default destination and not a file target
					if !isFileTarget {
						for _, target := range config.Targets {
							if target.GetDestination() == dest {
								// Always use maxBackups from target, as ReadBackupConfig
								// already sets the default value of 7 if it was empty
								maxBackups = target.MaxBackups
								break
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
							fmt.Printf("  %s⚠️  Warning: Failed to cleanup old backups -%s %v\n", ColorYellow, ColorReset, err)
						} else {
							fmt.Printf("  %s🔄 Rotation:%s Keeping latest %d backups\n", ColorCyan, ColorReset, maxBackups)
						}
					} else {
						fmt.Printf("  %s📄 File target:%s No rotation applied (single file backup)\n", ColorCyan, ColorReset)
					}

					// Record this backup in the config file if we're using a config
					if configFile != "" {
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
								fmt.Printf("  %s⚠️  Warning: Failed to update backup history in config -%s %v\n", ColorYellow, ColorReset, err)
							} else {
								fmt.Printf("  %s📝 History:%s Updated backup history in %s\n", ColorDim, ColorReset, configPath)
							}

							// Copy the config file to the destination with backup name prefix if enabled
							if copyConfig {
								configBaseName := filepath.Base(backupFileNameForTarget)
								configBaseName = strings.TrimSuffix(configBaseName, ".tar.gz") // Remove .tar.gz
								configBaseName = strings.TrimSuffix(configBaseName, ".gpg")    // Remove .gpg if encrypted

								// For file targets, copy config to the directory containing the file
								// For directory targets, copy config to the destination directory
								var destConfigDir string
								if isFileTarget {
									destConfigDir = filepath.Dir(dest)
								} else {
									destConfigDir = dest
								}
								destConfigPath := filepath.Join(destConfigDir, configBaseName+".backup.yaml")

								// Get the encryption receiver if encryption was used
								currentEncryptionReceiver := encryptionReceiver

								// Copy the config with added helpful comments
								if err := configService.CopyConfigWithHelp(configPath, destConfigPath, useEncryption, currentEncryptionReceiver); err != nil {
									fmt.Printf("  %s⚠️  Warning: Failed to copy config file to destination -%s %v\n", ColorYellow, ColorReset, err)
								} else {
									fmt.Printf("  %s📄 Config:%s Copied config file with usage info to %s\n", ColorGreen, ColorReset, destConfigPath)
								}
							}
						}
					}
				}
			}
		}

		// Clean up the temporary file
		os.Remove(tempBackupPath)
		fmt.Printf("\n%s%s🎉 Backup completed successfully!%s\n", ColorGreen, ColorBold, ColorReset)
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
	runCmd.Flags().BoolVar(&force, "force", false, "Force the backup operation, bypassing size warnings")

	// Add command to root
	rootCmd.AddCommand(runCmd)
}
