package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/kennycyb/go-backup/internal/service/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("AddBackupRecord", func() {
		It("should add a backup record to an existing target", func() {
			// Create a config with a target
			config := &BackupConfig{
				Targets: []BackupTarget{
					{
						Path:       "/backup/path",
						MaxBackups: 3,
						Backups:    []BackupRecord{},
					},
				},
			}

			// Create a test backup record
			record := BackupRecord{
				Filename:  "test-backup-20230101.tar.gz",
				Source:    "/source/path",
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Size:      1024,
			}

			// Add the record
			AddBackupRecord(config, "/backup/path", record)

			// Check that the record was added
			Expect(config.Targets[0].Backups).To(HaveLen(1))
			Expect(config.Targets[0].Backups[0].Filename).To(Equal("test-backup-20230101.tar.gz"))
		})

		It("should prepend new backups and respect maxBackups limit", func() {
			// Create a config with a target that already has some backups
			config := &BackupConfig{
				Targets: []BackupTarget{
					{
						Path:       "/backup/path",
						MaxBackups: 2,
						Backups: []BackupRecord{
							{
								Filename:  "test-backup-20230101.tar.gz",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							},
							{
								Filename:  "test-backup-20230102.tar.gz",
								CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
							},
						},
					},
				},
			}

			// Create a test backup record
			record := BackupRecord{
				Filename:  "test-backup-20230103.tar.gz",
				Source:    "/source/path",
				CreatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
				Size:      1024,
			}

			// Add the record
			AddBackupRecord(config, "/backup/path", record)

			// Check that we still have only 2 records (due to maxBackups)
			Expect(config.Targets[0].Backups).To(HaveLen(2))

			// Check that the new record is at the beginning
			Expect(config.Targets[0].Backups[0].Filename).To(Equal("test-backup-20230103.tar.gz"))
			Expect(config.Targets[0].Backups[1].Filename).To(Equal("test-backup-20230101.tar.gz"))
		})

		It("should do nothing when target is not found", func() {
			// Create a config with a target
			config := &BackupConfig{
				Targets: []BackupTarget{
					{
						Path:       "/backup/path",
						MaxBackups: 3,
						Backups:    []BackupRecord{},
					},
				},
			}

			// Create a test backup record
			record := BackupRecord{
				Filename:  "test-backup-20230101.tar.gz",
				Source:    "/source/path",
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Size:      1024,
			}

			// Add the record to a non-existent target
			AddBackupRecord(config, "/nonexistent/path", record)

			// Check that no records were added
			Expect(config.Targets[0].Backups).To(HaveLen(0))
		})		It("should set the default maxBackups value when adding to a target with 0 maxBackups", func() {
			// Create a config with a target that has zero maxBackups
			config := &BackupConfig{
				Targets: []BackupTarget{
					{
						Path:       "/backup/path",
						MaxBackups: 0, // Zero value should be replaced with default
						Backups:    []BackupRecord{},
					},
				},
			}

			// Add backup records
			for i := 1; i <= 10; i++ {
				record := BackupRecord{
					Filename:  fmt.Sprintf("test-backup-%d.tar.gz", i),
					Source:    "/source/path",
					CreatedAt: time.Now().AddDate(0, 0, -i),
					Size:      int64(i * 1000),
				}
				AddBackupRecord(config, "/backup/path", record)
			}

			// Should have set MaxBackups to 7 and limited the backups
			Expect(config.Targets[0].MaxBackups).To(Equal(7))
			Expect(config.Targets[0].Backups).To(HaveLen(7))

			// First backup in the list should be the most recently added one (10)
			Expect(config.Targets[0].Backups[0].Filename).To(Equal("test-backup-10.tar.gz"))
		})
	})
	var (
		tmpDir     string
		configPath string
	)

	BeforeEach(func() {
		// Create a temporary directory for test files
		var err error
		tmpDir, err = os.MkdirTemp("", "config-test")
		Expect(err).NotTo(HaveOccurred())
		configPath = filepath.Join(tmpDir, "test-config.yaml")
	})

	AfterEach(func() {
		// Clean up the temporary directory
		os.RemoveAll(tmpDir)
	})

	Describe("ReadBackupConfig", func() {
		Context("when the config file exists with valid content", func() {
			BeforeEach(func() {
				// Create a valid config file
				configContent := `
excludes:
  - ".git/**"
  - "node_modules/**"
target:
  - path: "/path/to/backup/location1"
    maxBackups: 5
  - path: "/path/to/backup/location2"
    maxBackups: 10
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads and parses the config correctly", func() {
				config, err := ReadBackupConfig(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())

				// Verify excludes
				Expect(config.Excludes).To(HaveLen(2))
				Expect(config.Excludes).To(ContainElements(".git/**", "node_modules/**"))

				// Verify targets
				Expect(config.Targets).To(HaveLen(2))
				Expect(config.Targets[0].Path).To(Equal("/path/to/backup/location1"))
				Expect(config.Targets[0].MaxBackups).To(Equal(5))
				Expect(config.Targets[1].Path).To(Equal("/path/to/backup/location2"))
				Expect(config.Targets[1].MaxBackups).To(Equal(10))
			})

			It("should apply default maxBackups value of 7 when missing", func() {
				// Create a valid config file with a missing maxBackups value
				configContent := `
excludes:
  - ".git/**"
target:
  - path: "/path/to/backup/location1"
    maxBackups: 5
  - path: "/path/to/backup/location2"
  - path: "/path/to/backup/location3"
    maxBackups: 0
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				Expect(err).NotTo(HaveOccurred())

				config, err := ReadBackupConfig(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())

				// Verify default values are applied
				Expect(config.Targets).To(HaveLen(3))
				Expect(config.Targets[0].MaxBackups).To(Equal(5)) // Specified value kept
				Expect(config.Targets[1].MaxBackups).To(Equal(7)) // Default applied when missing
				Expect(config.Targets[2].MaxBackups).To(Equal(7)) // Default applied when zero
			})
		})

		Context("when the config file does not exist", func() {
			It("returns an error", func() {
				nonExistentPath := filepath.Join(tmpDir, "non-existent.yaml")
				config, err := ReadBackupConfig(nonExistentPath)
				Expect(err).To(HaveOccurred())
				Expect(config).To(BeNil())
			})
		})

		Context("when the config file has invalid YAML", func() {
			BeforeEach(func() {
				// Create an invalid config file
				configContent := `
excludes:
  - ".git/**"
  - "node_modules/**"
target:
  - path: "/path/to/backup/location1"
    maxBackups: invalid_value  # This should be an integer
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				config, err := ReadBackupConfig(configPath)
				Expect(err).To(HaveOccurred())
				Expect(config).To(BeNil())
			})
		})

		Context("when the config file is empty", func() {
			BeforeEach(func() {
				// Create an empty config file
				err := os.WriteFile(configPath, []byte(""), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a config with empty values", func() {
				config, err := ReadBackupConfig(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Excludes).To(BeEmpty())
				Expect(config.Targets).To(BeEmpty())
			})
		})
	})

	Describe("WriteBackupConfig", func() {
		var config *BackupConfig

		BeforeEach(func() {
			// Create a test config
			config = &BackupConfig{
				Excludes: []string{".git/**", "node_modules/**"},
				Targets: []BackupTarget{
					{
						Path:       "/path/to/backup/location1",
						MaxBackups: 5,
					},
					{
						Path:       "/path/to/backup/location2",
						MaxBackups: 10,
					},
				},
			}
		})

		Context("when writing to a valid path", func() {
			It("writes the config file successfully", func() {
				err := WriteBackupConfig(configPath, config)
				Expect(err).NotTo(HaveOccurred())

				// Verify the file exists
				_, err = os.Stat(configPath)
				Expect(err).NotTo(HaveOccurred())

				// Read the file back and verify content
				readConfig, err := ReadBackupConfig(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(readConfig).NotTo(BeNil())

				// Verify excludes
				Expect(readConfig.Excludes).To(HaveLen(2))
				Expect(readConfig.Excludes).To(ContainElements(".git/**", "node_modules/**"))

				// Verify targets
				Expect(readConfig.Targets).To(HaveLen(2))
				Expect(readConfig.Targets[0].Path).To(Equal("/path/to/backup/location1"))
				Expect(readConfig.Targets[0].MaxBackups).To(Equal(5))
				Expect(readConfig.Targets[1].Path).To(Equal("/path/to/backup/location2"))
				Expect(readConfig.Targets[1].MaxBackups).To(Equal(10))
			})
		})

		Context("when writing to a non-existent directory", func() {
			It("creates the directory and writes the file", func() {
				nestedConfigPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")
				err := WriteBackupConfig(nestedConfigPath, config)
				Expect(err).NotTo(HaveOccurred())

				// Verify the file exists
				_, err = os.Stat(nestedConfigPath)
				Expect(err).NotTo(HaveOccurred())

				// Read the file back
				readConfig, err := ReadBackupConfig(nestedConfigPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(readConfig).NotTo(BeNil())
			})
		})

		Context("when the directory cannot be created", func() {
			It("returns an error", func() {
				// Create a file where we want a directory
				dirPath := filepath.Join(tmpDir, "file-not-dir")
				err := os.WriteFile(dirPath, []byte("not a directory"), 0644)
				Expect(err).NotTo(HaveOccurred())

				// Try to write config to a path that would need to create a directory
				// where a file already exists
				badConfigPath := filepath.Join(dirPath, "config.yaml")
				err = WriteBackupConfig(badConfigPath, config)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
