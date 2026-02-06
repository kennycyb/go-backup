package git_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/kennycyb/go-backup/internal/service/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Git", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "git-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("HasUncommittedChanges", func() {
		Context("when directory is not a git repository", func() {
			It("returns an error", func() {
				hasChanges, err := HasUncommittedChanges(tmpDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a git repository"))
				Expect(hasChanges).To(BeFalse())
			})
		})

		Context("when directory is a git repository", func() {
			BeforeEach(func() {
				// Initialize git repository
				cmd := exec.Command("git", "init")
				cmd.Dir = tmpDir
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				// Configure git user for commits
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("with no files", func() {
				It("returns false", func() {
					hasChanges, err := HasUncommittedChanges(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasChanges).To(BeFalse())
				})
			})

			Context("with untracked files", func() {
				BeforeEach(func() {
					// Create a new file
					testFile := filepath.Join(tmpDir, "test.txt")
					err := os.WriteFile(testFile, []byte("test content"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns true", func() {
					hasChanges, err := HasUncommittedChanges(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasChanges).To(BeTrue())
				})
			})

			Context("with staged changes", func() {
				BeforeEach(func() {
					// Create and stage a file
					testFile := filepath.Join(tmpDir, "test.txt")
					err := os.WriteFile(testFile, []byte("test content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd := exec.Command("git", "add", "test.txt")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns true", func() {
					hasChanges, err := HasUncommittedChanges(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasChanges).To(BeTrue())
				})
			})

			Context("with modified files", func() {
				BeforeEach(func() {
					// Create and commit a file
					testFile := filepath.Join(tmpDir, "test.txt")
					err := os.WriteFile(testFile, []byte("test content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd := exec.Command("git", "add", "test.txt")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "commit", "-m", "initial commit")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Modify the file
					err = os.WriteFile(testFile, []byte("modified content"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns true", func() {
					hasChanges, err := HasUncommittedChanges(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasChanges).To(BeTrue())
				})
			})

			Context("with all changes committed", func() {
				BeforeEach(func() {
					// Create, stage, and commit a file
					testFile := filepath.Join(tmpDir, "test.txt")
					err := os.WriteFile(testFile, []byte("test content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd := exec.Command("git", "add", "test.txt")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "commit", "-m", "initial commit")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns false", func() {
					hasChanges, err := HasUncommittedChanges(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasChanges).To(BeFalse())
				})
			})
		})
	})
})
