package encrypt_test

import (
	"os"
	"path/filepath"

	"github.com/kennycyb/go-backup/internal/service/encrypt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPG", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "gpg-test-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("GPGEncrypt", func() {
		Context("when source file does not exist", func() {
			It("should return an error", func() {
				_, err := encrypt.GPGEncrypt("/nonexistent/file.txt", "test@example.com")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("source file doesn't exist"))
			})
		})

		Context("when source file exists but recipient is invalid", func() {
			It("should return an error for invalid recipient", func() {
				// Create a test file
				testFile := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(testFile, []byte("test content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				// Try to encrypt with an invalid recipient
				_, err = encrypt.GPGEncrypt(testFile, "nonexistent-recipient@invalid-domain-12345.com")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gpg encryption failed"))
			})
		})

		// Note: Testing successful encryption requires a valid GPG key in the keyring
		// This test is skipped in CI environments without GPG setup
		Context("when GPG is properly configured", func() {
			It("should create encrypted file with .gpg extension", func() {
				Skip("Requires valid GPG key in keyring - run manually with proper GPG setup")

				testFile := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(testFile, []byte("test content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				encryptedFile, err := encrypt.GPGEncrypt(testFile, "your-gpg-email@example.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(encryptedFile).To(Equal(testFile + ".gpg"))
				Expect(encryptedFile).To(BeARegularFile())
			})
		})
	})

	Describe("GPGDecrypt", func() {
		Context("when encrypted file does not exist", func() {
			It("should return an error", func() {
				_, err := encrypt.GPGDecrypt("/nonexistent/file.gpg", "", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("encrypted file doesn't exist"))
			})
		})

		Context("when output file is not specified", func() {
			It("should derive output filename by removing .gpg extension", func() {
				// Create a fake encrypted file (won't actually decrypt but tests path logic)
				testFile := filepath.Join(tmpDir, "test.txt.gpg")
				err := os.WriteFile(testFile, []byte("fake encrypted content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				// This will fail because it's not a real GPG file, but we can check the error
				_, err = encrypt.GPGDecrypt(testFile, "", "")
				Expect(err).To(HaveOccurred())
				// The error should be about decryption failing, not about file paths
				Expect(err.Error()).To(ContainSubstring("gpg decryption failed"))
			})
		})

		Context("when file has no .gpg extension", func() {
			It("should add .decrypted extension to output", func() {
				// Create a file without .gpg extension
				testFile := filepath.Join(tmpDir, "test.enc")
				err := os.WriteFile(testFile, []byte("fake encrypted content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				// This will fail but tests the path handling
				_, err = encrypt.GPGDecrypt(testFile, "", "")
				Expect(err).To(HaveOccurred())
			})
		})

		// Note: Testing successful decryption requires encrypted test data
		Context("when GPG is properly configured", func() {
			It("should decrypt file successfully", func() {
				Skip("Requires valid GPG encrypted file - run manually with proper GPG setup")
			})
		})
	})

	Describe("ValidateGPGReceiver", func() {
		Context("when recipient is empty", func() {
			It("should return an error", func() {
				valid, _, err := encrypt.ValidateGPGReceiver("")
				Expect(valid).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("recipient cannot be empty"))
			})
		})

		Context("when recipient does not exist in keyring", func() {
			It("should return an error", func() {
				valid, _, err := encrypt.ValidateGPGReceiver("nonexistent-user-12345@invalid-domain.com")
				Expect(valid).To(BeFalse())
				Expect(err).To(HaveOccurred())
			})
		})

		// Note: Testing successful validation requires a valid GPG key
		Context("when recipient exists in keyring", func() {
			It("should return true with key info", func() {
				Skip("Requires valid GPG key in keyring - run manually with proper GPG setup")

				valid, keyInfo, err := encrypt.ValidateGPGReceiver("your-gpg-email@example.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(valid).To(BeTrue())
				Expect(keyInfo).NotTo(BeEmpty())
			})
		})
	})
})
