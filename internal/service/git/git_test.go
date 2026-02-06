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

	Describe("GetCurrentBranch", func() {
		Context("when directory is not a git repository", func() {
			It("returns an error", func() {
				branch, err := GetCurrentBranch(tmpDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a git repository"))
				Expect(branch).To(BeEmpty())
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

				// Create initial commit to establish branch
				testFile := filepath.Join(tmpDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "add", "test.txt")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "commit", "-m", "initial commit")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the current branch name", func() {
				branch, err := GetCurrentBranch(tmpDir)
				Expect(err).NotTo(HaveOccurred())
				// Default branch can be "master" or "main" depending on git config
				Expect(branch).To(Or(Equal("master"), Equal("main")))
			})

			Context("when on a different branch", func() {
				BeforeEach(func() {
					// Create and switch to a new branch
					cmd := exec.Command("git", "checkout", "-b", "feature-branch")
					cmd.Dir = tmpDir
					err := cmd.Run()
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns the current branch name", func() {
					branch, err := GetCurrentBranch(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(branch).To(Equal("feature-branch"))
				})
			})
		})
	})

	Describe("PullLatest", func() {
		Context("when directory is not a git repository", func() {
			It("returns an error", func() {
				hasUpdates, err := PullLatest(tmpDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a git repository"))
				Expect(hasUpdates).To(BeFalse())
			})
		})

		Context("when directory is a git repository with remote", func() {
			var remoteDir string

			BeforeEach(func() {
				var err error
				remoteDir, err = os.MkdirTemp("", "git-remote-test")
				Expect(err).NotTo(HaveOccurred())

				// Initialize remote repository as a bare repo
				cmd := exec.Command("git", "init", "--bare")
				cmd.Dir = remoteDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				// Clone the bare repo to tmpDir
				cmd = exec.Command("git", "clone", remoteDir, tmpDir)
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				// Configure git user in clone
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				// Create initial commit
				testFile := filepath.Join(tmpDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "add", "test.txt")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "commit", "-m", "initial commit")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "push", "origin", "HEAD")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(remoteDir)
			})

			Context("when repository is already up-to-date", func() {
				It("returns false for hasUpdates", func() {
					hasUpdates, err := PullLatest(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasUpdates).To(BeFalse())
				})
			})

			Context("when there are new commits to pull", func() {
				BeforeEach(func() {
					// Create another clone to push changes from
					var err error
					anotherClone, err := os.MkdirTemp("", "git-another-clone")
					Expect(err).NotTo(HaveOccurred())

					cmd := exec.Command("git", "clone", remoteDir, anotherClone)
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Configure git user
					cmd = exec.Command("git", "config", "user.email", "test@example.com")
					cmd.Dir = anotherClone
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "config", "user.name", "Test User")
					cmd.Dir = anotherClone
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Make a new commit and push
					testFile := filepath.Join(anotherClone, "new-file.txt")
					err = os.WriteFile(testFile, []byte("new content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "add", "new-file.txt")
					cmd.Dir = anotherClone
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "commit", "-m", "add new file")
					cmd.Dir = anotherClone
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "push", "origin", "HEAD")
					cmd.Dir = anotherClone
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Clean up the clone immediately after use
					os.RemoveAll(anotherClone)
				})

				It("returns true for hasUpdates", func() {
					hasUpdates, err := PullLatest(tmpDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(hasUpdates).To(BeTrue())
				})
			})

			Context("when repository is in the middle of a merge", func() {
				BeforeEach(func() {
					// Create a conflicting situation by creating another branch
					cmd := exec.Command("git", "checkout", "-b", "test-branch")
					cmd.Dir = tmpDir
					err := cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Make a change on the test branch
					testFile := filepath.Join(tmpDir, "test.txt")
					err = os.WriteFile(testFile, []byte("branch content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "add", "test.txt")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "commit", "-m", "branch change")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Switch back to master/main
					cmd = exec.Command("git", "checkout", "-")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Make a conflicting change on master/main
					err = os.WriteFile(testFile, []byte("master content"), 0644)
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "add", "test.txt")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					cmd = exec.Command("git", "commit", "-m", "master change")
					cmd.Dir = tmpDir
					err = cmd.Run()
					Expect(err).NotTo(HaveOccurred())

					// Start a merge that will conflict
					cmd = exec.Command("git", "merge", "test-branch")
					cmd.Dir = tmpDir
					_ = cmd.Run() // This will fail due to conflict, which is expected
				})

				It("returns an error indicating merge in progress", func() {
					hasUpdates, err := PullLatest(tmpDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("merge operation"))
					Expect(hasUpdates).To(BeFalse())
				})
			})
		})
	})
})
