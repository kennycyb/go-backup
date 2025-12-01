# Test Agent for go-backup

## Purpose

This agent assists with writing, running, and maintaining tests for the go-backup project.

## Testing Framework

- **Framework**: Ginkgo v2 with Gomega matchers
- **Test Files**: `*_test.go` files alongside source code
- **Suite Files**: `*_suite_test.go` for test setup

## Setting Up Tests

### Bootstrap a New Test Suite

```bash
# Navigate to the package directory
cd internal/service/mypackage

# Bootstrap creates the suite file (*_suite_test.go)
ginkgo bootstrap

# Generate a test file for a specific source file
ginkgo generate myfile.go
```

### Generated Files

- `ginkgo bootstrap` creates: `mypackage_suite_test.go`
- `ginkgo generate myfile.go` creates: `myfile_test.go`

## Running Tests

```bash
# Run all tests recursively
ginkgo run ./...

# Run tests in specific package
ginkgo run ./internal/service/config/
ginkgo run ./internal/service/backup/
ginkgo run ./internal/service/compress/

# Run with verbose output
ginkgo run -v ./...

# Run specific test by name (using focus)
ginkgo run --focus "should add a backup record" ./internal/service/config/

# Run tests and stop on first failure
ginkgo run --fail-fast ./...

# Run tests in parallel
ginkgo run -p ./...

# Watch mode - re-run tests on file changes
ginkgo watch ./...
```

## Test Structure

### Unit Tests Location

```
internal/service/
├── backup/
│   ├── backup_suite_test.go
│   ├── files_test.go
│   └── rotation_test.go
├── compress/
│   ├── compress_suite_test.go
│   └── filesize_test.go
├── config/
│   ├── config_suite_test.go
│   └── config_test.go
└── encrypt/
    └── (no tests yet)
```

## Writing Tests

### Basic Test Structure

```go
package config_test

import (
    . "github.com/kennycyb/go-backup/internal/service/config"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("FeatureName", func() {
    Describe("FunctionName", func() {
        It("should do something expected", func() {
            // Arrange
            input := setupTestData()

            // Act
            result := FunctionUnderTest(input)

            // Assert
            Expect(result).To(Equal(expected))
        })
    })
})
```

### Table-Driven Tests

```go
DescribeTable("testing multiple scenarios",
    func(input string, expected int) {
        result := ProcessInput(input)
        Expect(result).To(Equal(expected))
    },
    Entry("scenario 1", "input1", 1),
    Entry("scenario 2", "input2", 2),
)
```

## Test Guidelines

### DO

- Use descriptive test names that explain the behavior being tested
- Test both success and error scenarios
- Use table-driven tests for multiple similar test cases
- Clean up temporary files/directories in `AfterEach`
- Use `/tmp/` for test files, never the project's `.backup.yaml`
- Mock external dependencies (GPG, filesystem) when appropriate

### DO NOT

- Modify the actual `.backup.yaml` in the project root
- Leave test artifacts after test completion
- Skip error case testing
- Write tests that depend on external services

## Key Test Scenarios

### Config Tests

- Reading/writing backup configurations
- Adding/removing backup targets
- File target vs directory target handling
- `IsFileTarget()` and `GetDestination()` methods
- Backup record management

### Backup Tests

- File copy operations
- Backup rotation logic
- Cleanup of old backups

### Compress Tests

- Large file detection
- File size formatting
- Exclusion pattern matching

## Creating Test Fixtures

```go
BeforeEach(func() {
    // Create temp directory
    tmpDir, err = os.MkdirTemp("", "test-")
    Expect(err).NotTo(HaveOccurred())

    // Create test config
    testConfig = &BackupConfig{
        Excludes: []string{".git"},
        Targets: []BackupTarget{
            {Path: "/backup/path", MaxBackups: 7},
        },
    }
})

AfterEach(func() {
    // Cleanup
    os.RemoveAll(tmpDir)
})
```

## Debugging Failed Tests

```bash
# Run with verbose output
ginkgo run -v ./internal/service/config/

# Run specific failing test using focus
ginkgo run --focus "TestName" ./internal/service/config/

# Show coverage
ginkgo run --cover ./...

# Generate coverage report
ginkgo run --coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
ginkgo run --race ./...

# Show slow tests (useful for optimization)
ginkgo run --slow-spec-threshold=1s ./...
```
