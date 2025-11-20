package compress_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kennycyb/go-backup/internal/service/compress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filesize", func() {
	var (
		tempDir string
		cleanup func()
	)

	// setupTestFileSystem creates a temporary file system structure for testing
	setupTestFileSystem := func() (string, func()) {
		// Create a temporary directory that will be cleaned up after the test
		tempDir, err := os.MkdirTemp("", "filesize-test-")
		Expect(err).NotTo(HaveOccurred(), "Failed to create temporary directory")

		// Setup cleanup function
		cleanup := func() {
			os.RemoveAll(tempDir)
		}

		// Create some test files with different sizes
		files := map[string]int64{
			"small.txt":           1024,              // 1 KB
			"medium.dat":          5 * 1024 * 1024,   // 5 MB
			"large.dat":           100 * 1024 * 1024, // 100 MB
			"verylarge.dat":       500 * 1024 * 1024, // 500 MB
			"node_modules/pkg.js": 1024 * 1024,       // 1 MB in excluded dir
			".git/config":         512,               // 0.5 KB in excluded dir
			"project/docs.md":     2 * 1024 * 1024,   // 2 MB
			"project/src/main.go": 1*1024*1024 + 1,   // 1 MB + 1 byte
		}

		// Create directories
		Expect(os.MkdirAll(filepath.Join(tempDir, "node_modules"), 0755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tempDir, "project", "src"), 0755)).To(Succeed())

		// Create files with specific sizes
		for path, size := range files {
			createTestFile(filepath.Join(tempDir, path), size)
		}

		return tempDir, cleanup
	}

	// createTestFile creates a test file of the specified size
	createTestFile := func(path string, size int64) {
		file, err := os.Create(path)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test file %s", path)
		defer file.Close()

		// Set the file size
		err = file.Truncate(size)
		Expect(err).NotTo(HaveOccurred(), "Failed to resize file %s to %d bytes", path, size)

		// Set file modification time to ensure deterministic testing
		modTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
		err = os.Chtimes(path, modTime, modTime)
		Expect(err).NotTo(HaveOccurred(), "Failed to set modification time for %s", path)
	}

	BeforeEach(func() {
		// Setup test file system before each test
		tempDir, cleanup = setupTestFileSystem()
	})

	AfterEach(func() {
		// Clean up after each test
		cleanup()
	})

	Describe("CheckFileSizes", func() {
		Context("with different thresholds and excludes", func() {
			DescribeTable("checking various file size scenarios",
				func(excludes []string, maxSizeGB int64, expectFileOver string) {
					summary, err := compress.CheckFileSizes(tempDir, excludes, maxSizeGB)

					// Check there's no error
					Expect(err).NotTo(HaveOccurred())

					// If expecting a file over threshold
					if expectFileOver != "" {
						Expect(summary.FilesOverSize).NotTo(BeEmpty(), "Expected to find files over size threshold")
						Expect(summary.LargestFile).To(Equal(expectFileOver), "Expected largest file to be "+expectFileOver)
						Expect(summary.LargestFileSize).To(Equal(int64(500*1024*1024)), "Expected largest file to be 500 MB")
					} else {
						Expect(summary.FilesOverSize).To(BeEmpty(), "Did not expect to find files over size threshold")
					}
				},
				Entry("No files over 1 GB", []string{}, int64(1), ""),
				Entry("Large file over 50 MB", []string{}, int64(0), "verylarge.dat"), // 0 = 50 MB threshold
				Entry("With node_modules excluded", []string{"node_modules", ".git"}, int64(0), "verylarge.dat"),
			)
		})
	})

	Describe("ListLargeFiles", func() {
		DescribeTable("listing large files with different thresholds and excludes",
			func(excludes []string, thresholdMB int64, expectedCount int, expectedOrder []string) {
				largeFiles, err := compress.ListLargeFiles(tempDir, excludes, thresholdMB)

				// Check there's no error
				Expect(err).NotTo(HaveOccurred())

				// Check the number of files found
				Expect(largeFiles).To(HaveLen(expectedCount), "Expected to find %d large files", expectedCount)

				// Check that files are sorted properly by size (largest first)
				if len(largeFiles) > 0 && len(expectedOrder) > 0 {
					for i, expectedFileName := range expectedOrder {
						if i < len(largeFiles) {
							Expect(largeFiles[i].RelativePath).To(ContainSubstring(expectedFileName),
								"Expected file at position %d to be %s", i, expectedFileName)
						}
					}
				}

				// Verify human-readable file size format
				if len(largeFiles) > 0 {
					Expect(largeFiles[0].SizeHuman).To(ContainSubstring("MB"),
						"Expected human-readable size to contain unit")
				}
			},
			Entry("List files over 50 MB", []string{}, int64(50), 2,
				[]string{"verylarge.dat", "large.dat"}),
			Entry("List files over 1 MB", []string{}, int64(1), 4,
				[]string{"verylarge.dat", "large.dat", "medium.dat", "project/docs.md"}),
			Entry("With exclusions", []string{"node_modules", ".git", "project"}, int64(1), 2,
				[]string{"verylarge.dat", "large.dat"}),
		)
	})

	Describe("FormatFileSize", func() {
		DescribeTable("formatting file sizes to human readable format",
			func(size int64, expected string) {
				result := compress.FormatFileSize(size)
				Expect(result).To(ContainSubstring(expected),
					"Formatted file size should contain %s", expected)
			},
			Entry("512 bytes", int64(512), "0.50 KB"),
			Entry("1 KB", int64(1024), "1.00 KB"),
			Entry("1 MB", int64(1024*1024), "1.00 MB"),
			Entry("1 GB", int64(1024*1024*1024), "1.00 GB"),
			Entry("5 GB", int64(1024*1024*1024*5), "5.00 GB"),
			Entry("1 TB", int64(1024*1024*1024*1024), "1.00 TB"),
		)
	})

	Describe("CheckExcluded", func() {
		DescribeTable("checking file exclusion patterns",
			func(path string, excludes []string, expected bool) {
				// Call private function through exported test helper
				result := compress.TestHelperCheckExcluded(path, excludes)
				Expect(result).To(Equal(expected),
					"Path %s with excludes %v should return %v", path, excludes, expected)
			},
			Entry("No excludes", "file.txt", []string{}, false),
			Entry("Exact match", "file.txt", []string{"file.txt"}, true),
			Entry("Directory match", "dir/file.txt", []string{"dir"}, true),
			Entry("Different file in dir", "dir/file.txt", []string{"dir/other.txt"}, false),
			Entry("Subdirectory match", "dir/subdir/file.txt", []string{"dir"}, true),
			Entry("Glob match", "dir/file.txt", []string{"*.txt"}, true),
			Entry("Glob no match", "dir/file.go", []string{"*.txt"}, false),
			Entry("Node modules match", "node_modules/pkg.js", []string{"node_modules"}, true),
			Entry("Project node modules", "project/node_modules/pkg.js", []string{"node_modules"}, true),
		)
	})
})
