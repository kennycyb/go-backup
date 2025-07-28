package compress

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileSizeSummary contains information about file sizes
type FileSizeSummary struct {
	TotalSize       int64
	LargestFile     string
	LargestFileSize int64
	FilesOverSize   []string
}

// LargeFileInfo holds detailed information about a large file
type LargeFileInfo struct {
	Path         string
	RelativePath string
	Size         int64
	ModTime      time.Time
	SizeHuman    string
}

// checkExcluded checks if a path should be excluded based on the provided patterns
func checkExcluded(relPath string, excludes []string) bool {
	for _, exclude := range excludes {
		// Try exact match
		matched, _ := filepath.Match(exclude, relPath)

		// Try prefix match (directory)
		if !matched && strings.HasPrefix(relPath, exclude) {
			// Check if the relative path starts with the exclude pattern followed by path separator
			if len(relPath) == len(exclude) || (len(relPath) > len(exclude) && relPath[len(exclude)] == filepath.Separator) {
				return true
			}
		}

		if matched {
			return true
		}
	}

	return false
}

// TestHelperCheckExcluded exposes the checkExcluded function for testing
func TestHelperCheckExcluded(relPath string, excludes []string) bool {
	return checkExcluded(relPath, excludes)
}

// CheckFileSizes analyzes a directory and returns information about file sizes,
// specifically identifying files that exceed the specified maximum size
func CheckFileSizes(sourceDir string, excludes []string, maxSizeGB int64) (*FileSizeSummary, error) {
	// Convert GB to bytes using our constant
	maxSize := maxSizeGB * GB
	summary := &FileSizeSummary{}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path for exclusion checking
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		// Skip excluded directories and files
		if checkExcluded(relPath, excludes) {
			return nil
		}

		fileSize := info.Size()
		summary.TotalSize += fileSize

		// Track largest file
		if fileSize > summary.LargestFileSize {
			summary.LargestFileSize = fileSize
			summary.LargestFile = relPath
		}

		// Track files over the specified size
		if maxSize > 0 && fileSize > maxSize {
			summary.FilesOverSize = append(summary.FilesOverSize, relPath)
		}

		return nil
	})

	return summary, err
}

// ListLargeFiles finds files larger than the specified threshold and returns detailed information
// sorted by file size (largest first)
func ListLargeFiles(sourceDir string, excludes []string, thresholdMB int64) ([]LargeFileInfo, error) {
	thresholdBytes := thresholdMB * MB
	var largeFiles []LargeFileInfo

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path for exclusion checking and display
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		// Skip excluded directories and files
		if checkExcluded(relPath, excludes) {
			return nil
		}

		fileSize := info.Size()

		// If file is larger than threshold, add to the list
		if fileSize > thresholdBytes {
			largeFiles = append(largeFiles, LargeFileInfo{
				Path:         path,
				RelativePath: relPath,
				Size:         fileSize,
				ModTime:      info.ModTime(),
				SizeHuman:    FormatFileSize(fileSize),
			})
		}

		return nil
	})

	// Sort files by size (largest first)
	sort.Slice(largeFiles, func(i, j int) bool {
		return largeFiles[i].Size > largeFiles[j].Size
	})

	return largeFiles, err
}

// FormatFileSize converts file size in bytes to a human-readable format
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	switch exp {
	case 0:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(div))
	case 1:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(div))
	case 2:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(div))
	case 3:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(div))
	}

	return fmt.Sprintf("%.2f PB", float64(size)/float64(div))
}
