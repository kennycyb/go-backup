package config_test

import (
	"os"
	"path/filepath"

	. "github.com/kennycyb/go-backup/internal/service/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
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
				Targets: []struct {
					Path       string `yaml:"path"`
					MaxBackups int    `yaml:"maxBackups"`
				}{
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
