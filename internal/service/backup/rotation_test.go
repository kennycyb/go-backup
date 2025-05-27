package backup_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/kennycyb/go-backup/internal/service/backup"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		// Create a temporary directory for test files
		var err error
		tmpDir, err = os.MkdirTemp("", "backup-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up the temporary directory
		os.RemoveAll(tmpDir)
	})

	Describe("CleanupOldBackups", func() {
		var (
			testFiles      []string
			testPrefix     string
			createTestFile func(name string, modTime time.Time) string
		)

		BeforeEach(func() {
			testFiles = []string{}
			testPrefix = "test-backup"

			// Helper function to create a test backup file with a specific modification time
			createTestFile = func(name string, modTime time.Time) string {
				filePath := filepath.Join(tmpDir, name)
				err := os.WriteFile(filePath, []byte("test backup content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				// Set the modification time
				err = os.Chtimes(filePath, modTime, modTime)
				Expect(err).NotTo(HaveOccurred())

				testFiles = append(testFiles, filePath)
				return filePath
			}
		})

		Context("when there are more backups than the limit", func() {
			It("deletes older backups", func() {
				now := time.Now()

				// Create some backup files with different timestamps
				createTestFile(testPrefix+"-20240101-120000.tar.gz", now.Add(-10*24*time.Hour)) // Oldest
				createTestFile(testPrefix+"-20240102-120000.tar.gz", now.Add(-9*24*time.Hour))
				createTestFile(testPrefix+"-20240103-120000.tar.gz", now.Add(-8*24*time.Hour))
				createTestFile(testPrefix+"-20240104-120000.tar.gz", now.Add(-7*24*time.Hour))
				createTestFile(testPrefix+"-20240105-120000.tar.gz", now.Add(-6*24*time.Hour)) // Newest

				// Should keep only the 3 newest backups
				err := CleanupOldBackups(tmpDir, testPrefix+"-", 3)
				Expect(err).NotTo(HaveOccurred())

				// Check that only 3 files are left
				files, err := os.ReadDir(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				var remainingFiles []string
				for _, file := range files {
					remainingFiles = append(remainingFiles, file.Name())
				}

				// We should have only 3 files left
				Expect(remainingFiles).To(HaveLen(3))

				// These should be the newer files
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240104-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240105-120000.tar.gz"))

				// The oldest files should be deleted
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.tar.gz"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240102-120000.tar.gz"))
			})
		})

		Context("when there are fewer backups than the limit", func() {
			It("doesn't delete any backups", func() {
				now := time.Now()

				// Create fewer backup files than the limit
				createTestFile(testPrefix+"-20240101-120000.tar.gz", now.Add(-2*24*time.Hour))
				createTestFile(testPrefix+"-20240102-120000.tar.gz", now.Add(-1*24*time.Hour))

				// Set limit higher than the number of backups
				err := CleanupOldBackups(tmpDir, testPrefix+"-", 5)
				Expect(err).NotTo(HaveOccurred())

				// Check that all 2 files are still there
				files, err := os.ReadDir(tmpDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(HaveLen(2))
			})
		})

		Context("when there are files with different prefixes", func() {
			It("only considers files with the specified prefix", func() {
				now := time.Now()

				// Create files with the target prefix
				createTestFile(testPrefix+"-20240101-120000.tar.gz", now.Add(-3*24*time.Hour))
				createTestFile(testPrefix+"-20240102-120000.tar.gz", now.Add(-2*24*time.Hour))
				createTestFile(testPrefix+"-20240103-120000.tar.gz", now.Add(-1*24*time.Hour))

				// Create files with a different prefix
				createTestFile("other-prefix-20240101-120000.tar.gz", now.Add(-3*24*time.Hour))
				createTestFile("other-prefix-20240102-120000.tar.gz", now.Add(-2*24*time.Hour))

				// Should keep only the 2 newest backups with the specified prefix
				err := CleanupOldBackups(tmpDir, testPrefix+"-", 2)
				Expect(err).NotTo(HaveOccurred())

				// Check the files in the directory
				files, err := os.ReadDir(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				var remainingFiles []string
				for _, file := range files {
					remainingFiles = append(remainingFiles, file.Name())
				}

				// We should have 4 files left: 2 with the target prefix and 2 with the other prefix
				Expect(remainingFiles).To(HaveLen(4))

				// Should have kept the 2 newest from the target prefix
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240102-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.tar.gz"))

				// Should have deleted the oldest from the target prefix
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.tar.gz"))

				// Should not have touched files with different prefixes
				Expect(remainingFiles).To(ContainElement("other-prefix-20240101-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement("other-prefix-20240102-120000.tar.gz"))
			})
		})

		Context("when there are associated config files", func() {
			It("deletes both backups and their associated config files", func() {
				now := time.Now()

				// Create some backup files with different timestamps
				createTestFile(testPrefix+"-20240101-120000.tar.gz", now.Add(-10*24*time.Hour)) // Oldest - to be deleted
				createTestFile(testPrefix+"-20240102-120000.tar.gz", now.Add(-9*24*time.Hour))  // To be deleted
				createTestFile(testPrefix+"-20240103-120000.tar.gz", now.Add(-8*24*time.Hour))  // Keep
				createTestFile(testPrefix+"-20240104-120000.tar.gz", now.Add(-7*24*time.Hour))  // Keep
				createTestFile(testPrefix+"-20240105-120000.tar.gz", now.Add(-6*24*time.Hour))  // Keep - Newest

				// Create associated config files for each backup
				createTestFile(testPrefix+"-20240101-120000.backup.yaml", now.Add(-10*24*time.Hour)) // Should be deleted
				createTestFile(testPrefix+"-20240102-120000.backup.yaml", now.Add(-9*24*time.Hour))  // Should be deleted
				createTestFile(testPrefix+"-20240103-120000.backup.yaml", now.Add(-8*24*time.Hour))  // Should be kept
				createTestFile(testPrefix+"-20240104-120000.backup.yaml", now.Add(-7*24*time.Hour))  // Should be kept
				createTestFile(testPrefix+"-20240105-120000.backup.yaml", now.Add(-6*24*time.Hour))  // Should be kept

				// Should keep only the 3 newest backups and their config files
				err := CleanupOldBackups(tmpDir, testPrefix+"-", 3)
				Expect(err).NotTo(HaveOccurred())

				// Check the files in the directory
				files, err := os.ReadDir(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				var remainingFiles []string
				for _, file := range files {
					remainingFiles = append(remainingFiles, file.Name())
				}

				// We should have 6 files left: 3 backup files and 3 config files
				Expect(remainingFiles).To(HaveLen(6))

				// These should be the newer files - both backups and configs
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240104-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240105-120000.tar.gz"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.backup.yaml"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240104-120000.backup.yaml"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240105-120000.backup.yaml"))

				// The oldest files should be deleted - both backups and configs
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.tar.gz"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240102-120000.tar.gz"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.backup.yaml"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240102-120000.backup.yaml"))
			})

			It("handles encrypted backups with .gpg extension", func() {
				now := time.Now()

				// Create some encrypted backup files with different timestamps
				createTestFile(testPrefix+"-20240101-120000.tar.gz.gpg", now.Add(-10*24*time.Hour)) // Oldest - to be deleted
				createTestFile(testPrefix+"-20240102-120000.tar.gz.gpg", now.Add(-9*24*time.Hour))  // To be deleted
				createTestFile(testPrefix+"-20240103-120000.tar.gz.gpg", now.Add(-8*24*time.Hour))  // Keep

				// Create associated config files for each backup
				createTestFile(testPrefix+"-20240101-120000.backup.yaml", now.Add(-10*24*time.Hour)) // Should be deleted
				createTestFile(testPrefix+"-20240102-120000.backup.yaml", now.Add(-9*24*time.Hour))  // Should be deleted
				createTestFile(testPrefix+"-20240103-120000.backup.yaml", now.Add(-8*24*time.Hour))  // Should be kept

				// Should keep only the newest backup and its config file
				err := CleanupOldBackups(tmpDir, testPrefix+"-", 1)
				Expect(err).NotTo(HaveOccurred())

				// Check the files in the directory
				files, err := os.ReadDir(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				var remainingFiles []string
				for _, file := range files {
					remainingFiles = append(remainingFiles, file.Name())
				}

				// We should have 2 files left: 1 encrypted backup file and 1 config file
				Expect(remainingFiles).To(HaveLen(2))

				// These should be the newer files - both encrypted backup and config
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.tar.gz.gpg"))
				Expect(remainingFiles).To(ContainElement(testPrefix + "-20240103-120000.backup.yaml"))

				// The oldest files should be deleted - both encrypted backups and configs
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.tar.gz.gpg"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240102-120000.tar.gz.gpg"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240101-120000.backup.yaml"))
				Expect(remainingFiles).NotTo(ContainElement(testPrefix + "-20240102-120000.backup.yaml"))
			})
		})
	})
})
