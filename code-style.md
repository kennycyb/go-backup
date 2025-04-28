# Go Backup Project Code Style Guide

This document outlines the coding standards and best practices for the go-backup project.

## General Guidelines

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `go vet` and `golint` before submitting changes
- Ensure code passes `go test ./...`

## Code Organization

- Keep package names short, lowercase, and descriptive
- Group related functionality in the same package
- Use `/cmd` for application entrypoints
- Use `/pkg` for library code
- Use `/internal` for private code

## Naming Conventions

- Use MixedCaps (CamelCase) for exported names
- Use mixedCaps (lower camel case) for non-exported names
- Use short, descriptive variable names
- Prefer clarity over brevity
- Avoid abbreviations unless they're well-known

```go
// Good
func ProcessBackup(targetPath string) error {}  // Exported
func validatePath(path string) bool {}          // Non-exported
```

## Comments

- Begin all exported functions, types, and variables with a comment
- Use complete sentences that end with a period
- Use comments to explain why, not how
- Use godoc style comments for packages and exported identifiers

```go
// Package backup provides functionality for backing up files and directories.
package backup

// ProcessBackup creates a backup of the files at the given target path.
// It returns an error if the backup fails.
func ProcessBackup(targetPath string) error {}
```

## Error Handling

- Check errors and handle them gracefully
- Return errors rather than using panic
- Wrap errors with additional context using `fmt.Errorf("context: %w", err)`
- Don't use `_` to ignore errors unless explicitly justified

```go
// Good
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("failed to read config from %s: %w", path, err)
}
```

## Testing

- Write tests for all exported functions
- Use table-driven tests when appropriate
- Use Ginkgo/Gomega for BDD-style testing
- Keep tests in the same package with a `_test.go` suffix
- Aim for high test coverage while maintaining test quality

## Concurrency

- Use channels and goroutines judiciously
- Always ensure proper cleanup and resource management
- Consider using sync.WaitGroup for goroutine synchronization
- Use context for cancellation and timeouts

## Formatting

- Line length should be reasonable (< 120 characters)
- Use tabs for indentation (not spaces)
- Leave empty lines between logical sections of code

## Dependencies

- Minimize external dependencies
- Pin dependency versions using Go modules
- Check license compatibility before adding dependencies

## Configuration

- Use YAML for configuration files
- Support environment variable overrides for configuration
- Validate configuration values early

## Logging

- Use structured logging rather than simple print statements
- Include appropriate context in log messages
- Use appropriate log levels (debug, info, warning, error)

This guide is a living document and may be updated as our practices evolve.