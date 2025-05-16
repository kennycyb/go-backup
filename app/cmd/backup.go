package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	source      string
	destination string
	compress    bool
	configFile  string
	excludeDirs []string
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a new backup",
	Long: `Create a new backup of specified files or directories.
This command will package and compress the specified sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating backup...")

		// If source is empty, use current directory
		if source == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			source = currentDir
		}

		// Create a timestamp for the backup file
		timestamp := time.Now().Format("20060102-150405")
		backupFileName := fmt.Sprintf("backup-%s.tar.gz", timestamp)
		tempBackupPath := filepath.Join(os.TempDir(), backupFileName)

		fmt.Printf("Source: %s\n", source)
		fmt.Printf("Temporary backup file: %s\n", tempBackupPath)

		// Create the tar.gz archive
		err := createTarGzArchive(source, tempBackupPath)
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
			// Read from config file
			configPath := ".backup.yaml"
			if configFile != "" {
				configPath = configFile
			}

			config, err := readBackupConfig(configPath)
			if err != nil {
				fmt.Printf("Error reading config file %s: %v\n", configPath, err)
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
		for _, dest := range destinations {
			// Create destination directory if it doesn't exist
			if err := os.MkdirAll(dest, 0755); err != nil {
				fmt.Printf("Error creating destination directory %s: %v\n", dest, err)
				continue
			}

			destFilePath := filepath.Join(dest, backupFileName)
			fmt.Printf("Copying backup to: %s\n", destFilePath)

			if err := copyFile(tempBackupPath, destFilePath); err != nil {
				fmt.Printf("Error copying backup to %s: %v\n", destFilePath, err)
			} else {
				fmt.Printf("Backup successfully copied to %s\n", destFilePath)
			}
		}

		// Clean up the temporary file
		os.Remove(tempBackupPath)
		fmt.Println("Backup completed successfully!")
	},
}

// createTarGzArchive creates a compressed tar archive from the source directory
func createTarGzArchive(sourceDir, targetFile string) error {
	// Create the target file
	tarFile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}
	defer tarFile.Close()

	// Create a gzip writer
	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the source directory
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		for _, exclude := range excludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip the temporary directory and target backup locations
		if strings.HasPrefix(path, os.TempDir()) {
			return nil
		}

		// Get the relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		// Skip if it's the root directory
		if relPath == "." {
			return nil
		}

		// Create a header based on the file info
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("error creating tar header: %w", err)
		}

		// Update the header name to use the relative path
		header.Name = relPath

		// Write the header to the archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header: %w", err)
		}

		// If it's a regular file, write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("error writing file contents to tar: %w", err)
			}
		}

		return nil
	})
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

// readBackupConfig reads the backup configuration from the specified file
func readBackupConfig(filePath string) (*BackupConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config BackupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func init() {
	// Local flags for the backup command
	backupCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory to backup (defaults to current directory)")
	backupCmd.Flags().StringVarP(&destination, "dest", "d", "", "Destination directory for backup (if not specified, uses config file)")
	backupCmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress the backup")
	backupCmd.Flags().StringVarP(&configFile, "config", "f", ".backup.yaml", "Config file path")
	backupCmd.Flags().StringSliceVar(&excludeDirs, "exclude", []string{".git", "node_modules", "bin"}, "Directories to exclude from backup")
}
