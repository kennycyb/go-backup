# AGENTS.md - AI Assistant Guidelines for go-backup

## Project Overview

go-backup is a Go-based backup utility that creates compressed and encrypted backups of directories to multiple target locations with rotation support.

## Key Concepts

### Backup Target Types

1. **Directory Targets** (`path:`): Backups are stored in a directory with timestamped filenames and rotation

   ```yaml
   target:
     - path: /backups/location1
       maxBackups: 7
   ```

2. **File Targets** (`file:`): Single file backup that overwrites the same file each time (no rotation)
   ```yaml
   target:
     - file: /backups/single-backup.tar.gz.gpg
   ```

### Configuration Structure

The `.backup.yaml` file contains:

- `excludes`: List of directories/patterns to exclude from backup
- `target`: List of backup targets (either `path:` or `file:`)
- `encryption`: GPG encryption settings (method, receiver)

## Code Guidelines

### File Locations

- Commands: `app/cmd/`
- Services: `internal/service/`
  - `backup/`: Backup operations (copy, rotation)
  - `compress/`: Tar/gzip compression
  - `config/`: Configuration reading/writing
  - `encrypt/`: GPG encryption

### Important Types

```go
// BackupTarget - supports both path and file targets
type BackupTarget struct {
    Path       string         `yaml:"path,omitempty"`
    File       string         `yaml:"file,omitempty"`
    MaxBackups int            `yaml:"maxBackups,omitempty"`
    Backups    []BackupRecord `yaml:"backups,omitempty"`
}

// Helper methods
func (t BackupTarget) IsFileTarget() bool      // Returns true if File field is set
func (t BackupTarget) GetDestination() string  // Returns File or Path
```

### Testing

- Use Ginkgo/Gomega for tests
- Test files: `*_test.go` with `_suite_test.go` for setup
- Run tests: `go test ./...`
- Run specific package: `go test ./internal/service/config/`

### DO NOT

- Overwrite `.backup.yaml` in the project root during development
- Use `/tmp/.backup.yaml` for testing instead
- Hardcode sensitive information (keys, passphrases)

### DO

- Use `GetDestination()` method to get target path (handles both file and path targets)
- Use `IsFileTarget()` to check target type before applying rotation logic
- Preserve encryption config structure (not as array)
- Clean up associated config files when rotating backups

## Common Tasks

### Adding a New Command

1. Create new file in `app/cmd/`
2. Define cobra command with `Use`, `Short`, `Long`, `Run`
3. Register in `init()` with `rootCmd.AddCommand()`

### Modifying Config Structure

1. Update types in `internal/service/config/config.go`
2. Add helper methods if needed
3. Update tests in `config_test.go`
4. Ensure backward compatibility with existing configs

## Build & Run

```bash
go build                           # Build
go run main.go run                 # Run backup
go run main.go init                # Initialize config
go test ./...                      # Run all tests
```
