package backup_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kennycyb/go-backup/internal/service/backup"
)

var _ = Describe("Files", func() {
	Describe("CopyFile", func() {
		var (
			tempDir  string
			srcFile  string
			destFile string
		)

		BeforeEach(func() {
			// Create a temporary directory for testing
			var err error
			tempDir, err = os.MkdirTemp("", "files-test")
			Expect(err).NotTo(HaveOccurred())

			// Create a source test file with some content
			srcFile = filepath.Join(tempDir, "source.txt")
			destFile = filepath.Join(tempDir, "destination.txt")

			err = os.WriteFile(srcFile, []byte("test content"), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// Clean up the temporary directory
			os.RemoveAll(tempDir)
		})

		It("should copy a file from source to destination", func() {
			// Copy the file
			err := backup.CopyFile(srcFile, destFile)
			Expect(err).NotTo(HaveOccurred())

			// Verify the destination file exists
			Expect(destFile).To(BeARegularFile())

			// Read both files and verify contents are the same
			srcContent, err := os.ReadFile(srcFile)
			Expect(err).NotTo(HaveOccurred())

			destContent, err := os.ReadFile(destFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(destContent).To(Equal(srcContent))
		})

		It("should return an error when source file doesn't exist", func() {
			// Try to copy from a non-existent source file
			nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
			err := backup.CopyFile(nonExistentFile, destFile)

			// Should return an error
			Expect(err).To(HaveOccurred())
		})

		It("should return an error when destination directory doesn't exist", func() {
			// Try to copy to a destination in a non-existent directory
			invalidDestFile := filepath.Join(tempDir, "nonexistent-dir", "dest.txt")
			err := backup.CopyFile(srcFile, invalidDestFile)

			// Should return an error
			Expect(err).To(HaveOccurred())
		})
	})
})
