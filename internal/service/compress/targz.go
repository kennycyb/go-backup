package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateTarGzArchive creates a compressed tar archive from the source directory,
// excluding the specified paths. Returns an error if the operation fails.
func CreateTarGzArchive(sourceDir, targetFile string, excludes []string) error {
	// Create the target file
	tarFile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}
	defer tarFile.Close()

	// Create a gzip writer
	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	// Create a tar writer with PAX format for large file support
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the source directory
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path for exclusion checking
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		// Skip if it's the root directory
		if relPath == "." {
			return nil
		}

		// Skip excluded directories and files
		for _, exclude := range excludes {
			// Check for exact match, prefix match with /, or glob pattern
			matched, _ := filepath.Match(exclude, relPath)
			if matched || strings.Contains(relPath, exclude) || strings.HasPrefix(relPath, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip the temporary directory
		if strings.HasPrefix(path, os.TempDir()) {
			return nil
		}

		// Create a header based on the file info
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("error creating tar header: %w", err)
		}

		// Update the header name to use the relative path
		header.Name = relPath

		// Use PAX format for large files
		if info.Size() > RecommendedMaxFileSize {
			header.Format = tar.FormatPAX
		}

		// Write the header to the archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header for %s: %w", path, err)
		}

		// If it's a regular file, write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", path, err)
			}
			defer file.Close()

			// Create a wrapper to handle files that might be too large
			if _, err := io.Copy(tarWriter, file); err != nil {
				if strings.Contains(err.Error(), "write too long") {
					return fmt.Errorf("file %s is too large for tar format (consider splitting large files): %w", path, err)
				}
				return fmt.Errorf("error writing file contents to tar: %w", err)
			}
		}

		return nil
	})
}
