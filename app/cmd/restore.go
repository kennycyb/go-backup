package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	encryptionService "github.com/kennycyb/go-backup/internal/service/encrypt"
	"github.com/spf13/cobra"
)

var (
	backupFile    string
	targetDir     string
	overwrite     bool
	decrypt       bool
	useConfigFile bool
	passphrase    string
	askPassphrase bool
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore from a backup",
	Long: `Restore files from a previously created backup.
This command will extract and restore files from a backup archive.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restoring from backup...")
		fmt.Printf("Backup file: %s\n", backupFile)
		fmt.Printf("Target directory: %s\n", targetDir)
		fmt.Printf("Overwrite existing: %v\n", overwrite)

		// Process the backup file name
		backupFileBaseName := filepath.Base(backupFile)

		// Remove extension (could be .tar.gz or .tar.gz.gpg)
		nameWithoutExt := strings.TrimSuffix(backupFileBaseName, filepath.Ext(backupFileBaseName))
		if strings.HasSuffix(nameWithoutExt, ".tar") {
			nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".tar")
		}

		// Check for associated config file
		associatedConfigFile := nameWithoutExt + ".backup.yaml"
		associatedConfigPath := filepath.Join(filepath.Dir(backupFile), associatedConfigFile)

		// Check if the associated config file exists and use it if requested
		if useConfigFile {
			if _, err := os.Stat(associatedConfigPath); err == nil {
				fmt.Printf("Found associated config file: %s\n", associatedConfigPath)

				// TODO: In a future implementation, you could use this config file for
				// advanced restore options, such as applying the same exclude rules
				// or finding additional backup metadata
			} else {
				fmt.Printf("No associated config file found at: %s\n", associatedConfigPath)
			}
		}

		// Handle GPG encrypted backups
		if decrypt || strings.HasSuffix(backupFile, ".gpg") {
			fmt.Println("Detected GPG encrypted backup, decrypting...")

			// Create temporary file path for the decrypted archive
			tempOutputFile := filepath.Join(os.TempDir(), filepath.Base(backupFile))
			if strings.HasSuffix(tempOutputFile, ".gpg") {
				tempOutputFile = tempOutputFile[:len(tempOutputFile)-4]
			}

			// Check for passphrase in config if useConfigFile is true
			configPassphrase := ""
			if useConfigFile && passphrase == "" && !askPassphrase {
				if _, err := os.Stat(associatedConfigPath); err == nil {
					// Read config to check for passphrase
					config, err := configService.ReadBackupConfig(associatedConfigPath)
					if err == nil && config != nil && config.Encryption != nil {
						if config.Encryption.Method == "gpg" && config.Encryption.Passphrase != "" {
							configPassphrase = config.Encryption.Passphrase
							fmt.Println("Using passphrase from config file")
						}
					}
				}
			}

			// If askPassphrase flag is set, prompt for passphrase
			promptedPassphrase := ""
			if askPassphrase && passphrase == "" {
				fmt.Print("Enter passphrase for GPG decryption: ")
				fmt.Scanln(&promptedPassphrase)
			}

			// Use provided passphrase, prompted passphrase, or config passphrase
			finalPassphrase := passphrase
			if finalPassphrase == "" {
				finalPassphrase = promptedPassphrase
			}
			if finalPassphrase == "" {
				finalPassphrase = configPassphrase
			}

			// Decrypt the backup file
			decryptedPath, err := encryptionService.GPGDecrypt(backupFile, tempOutputFile, finalPassphrase)
			if err != nil {
				// If decryption failed and we didn't explicitly ask for the passphrase, try prompting
				if finalPassphrase == "" && !askPassphrase {
					fmt.Println("Decryption failed, passphrase may be required.")
					fmt.Print("Enter passphrase for GPG decryption: ")
					fmt.Scanln(&promptedPassphrase)

					// Retry decryption with the entered passphrase
					decryptedPath, err = encryptionService.GPGDecrypt(backupFile, tempOutputFile, promptedPassphrase)
					if err != nil {
						fmt.Printf("Error decrypting backup: %v\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Printf("Error decrypting backup: %v\n", err)
					os.Exit(1)
				}
			}

			fmt.Printf("Decrypted to: %s\n", decryptedPath)

			// Use the decrypted file for restoration
			backupFile = decryptedPath

			// Make sure to clean up the temporary decrypted file when done
			defer os.Remove(decryptedPath)
		}

		// TODO: Implement restore functionality using the (decrypted) backup file
		fmt.Println("Restoration completed!")
	},
}

func init() {
	// Local flags for the restore command
	restoreCmd.Flags().StringVarP(&backupFile, "file", "f", "", "Backup file to restore from (required)")
	restoreCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory to restore to")
	restoreCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files")
	restoreCmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "Force decrypt the backup file (auto-detected for .gpg files)")
	restoreCmd.Flags().BoolVar(&useConfigFile, "use-config", true, "Use the associated backup configuration file if found")
	restoreCmd.Flags().StringVar(&passphrase, "passphrase", "", "Passphrase for GPG decryption (if needed)")
	restoreCmd.Flags().BoolVar(&askPassphrase, "ask-passphrase", false, "Prompt for a passphrase")

	// Mark required flags
	restoreCmd.MarkFlagRequired("file")

	// Add command to root
	rootCmd.AddCommand(restoreCmd)
}
