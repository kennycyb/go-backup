package config_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/kennycyb/go-backup/internal/service/config"
)

var _ = Describe("Global Registry", func() {
	var tempDir string
	var globalConfigPath string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "backup-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Create a fake home directory for testing
		homeDir := filepath.Join(tempDir, "home")
		err = os.MkdirAll(homeDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		globalConfigPath = filepath.Join(homeDir, ".backup.yaml")

		// Override home directory for testing
		os.Setenv("HOME", homeDir)
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
		os.Unsetenv("HOME")
	})

	Describe("UpdateGlobalRegistry", func() {
		Context("when global config does not exist", func() {
			It("should return nil without error", func() {
				backupDir := filepath.Join(tempDir, "my-backup")
				err := os.MkdirAll(backupDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				err = config.UpdateGlobalRegistry(backupDir)
				Expect(err).NotTo(HaveOccurred())

				// Verify global config was not created
				_, err = os.Stat(globalConfigPath)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})

		Context("when global config exists", func() {
			BeforeEach(func() {
				// Create initial global config
				initialConfig := `default:
  encryption:
    method: gpg
    receiver: test@example.com
backups: []
`
				err := os.WriteFile(globalConfigPath, []byte(initialConfig), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add new backup location", func() {
				backupDir := filepath.Join(tempDir, "my-backup")
				err := os.MkdirAll(backupDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				err = config.UpdateGlobalRegistry(backupDir)
				Expect(err).NotTo(HaveOccurred())

				// Read and verify the updated config
				data, err := os.ReadFile(globalConfigPath)
				Expect(err).NotTo(HaveOccurred())

				var registry config.GlobalBackupRegistry
				err = yaml.Unmarshal(data, &registry)
				Expect(err).NotTo(HaveOccurred())

				Expect(registry.Backups).To(HaveLen(1))

				absPath, err := filepath.Abs(backupDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(registry.Backups[0].Location).To(Equal(absPath))
				Expect(registry.Backups[0].RunAt).To(BeTemporally("~", time.Now(), time.Second))
			})

			It("should update existing backup location", func() {
				backupDir := filepath.Join(tempDir, "my-backup")
				err := os.MkdirAll(backupDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				// Add initial entry
				err = config.UpdateGlobalRegistry(backupDir)
				Expect(err).NotTo(HaveOccurred())

				// Wait a bit to ensure different timestamp
				time.Sleep(100 * time.Millisecond)

				// Update the same location
				err = config.UpdateGlobalRegistry(backupDir)
				Expect(err).NotTo(HaveOccurred())

				// Read and verify
				data, err := os.ReadFile(globalConfigPath)
				Expect(err).NotTo(HaveOccurred())

				var registry config.GlobalBackupRegistry
				err = yaml.Unmarshal(data, &registry)
				Expect(err).NotTo(HaveOccurred())

				// Should still have only one entry
				Expect(registry.Backups).To(HaveLen(1))
			})

			It("should handle multiple backup locations", func() {
				backup1 := filepath.Join(tempDir, "backup1")
				backup2 := filepath.Join(tempDir, "backup2")

				err := os.MkdirAll(backup1, 0755)
				Expect(err).NotTo(HaveOccurred())
				err = os.MkdirAll(backup2, 0755)
				Expect(err).NotTo(HaveOccurred())

				// Add first location
				err = config.UpdateGlobalRegistry(backup1)
				Expect(err).NotTo(HaveOccurred())

				// Add second location
				err = config.UpdateGlobalRegistry(backup2)
				Expect(err).NotTo(HaveOccurred())

				// Read and verify
				data, err := os.ReadFile(globalConfigPath)
				Expect(err).NotTo(HaveOccurred())

				var registry config.GlobalBackupRegistry
				err = yaml.Unmarshal(data, &registry)
				Expect(err).NotTo(HaveOccurred())

				Expect(registry.Backups).To(HaveLen(2))
			})

			It("should preserve default encryption settings", func() {
				backupDir := filepath.Join(tempDir, "my-backup")
				err := os.MkdirAll(backupDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				err = config.UpdateGlobalRegistry(backupDir)
				Expect(err).NotTo(HaveOccurred())

				// Read and verify
				data, err := os.ReadFile(globalConfigPath)
				Expect(err).NotTo(HaveOccurred())

				var registry config.GlobalBackupRegistry
				err = yaml.Unmarshal(data, &registry)
				Expect(err).NotTo(HaveOccurred())

				Expect(registry.Default.Encryption).NotTo(BeNil())
				Expect(registry.Default.Encryption.Method).To(Equal("gpg"))
				Expect(registry.Default.Encryption.Receiver).To(Equal("test@example.com"))
			})
		})
	})
})
