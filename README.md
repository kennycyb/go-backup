# go-backup

A command line tool for creating, managing, and restoring backups.

## Features

- Encrypted backups using GPG
- Multiple backup targets
- Configurable retention policies
- File/directory exclusion patterns
- Automated cleanup of old backups

## Installation

```bash
go install github.com/kennycyb/go-backup/cmd/backup@latest
```

## Usage

The tool can be used via the provided Makefile:

```bash
# Build the tool
make build

# Create a backup using default configuration
make run

# List all backups
make list

# List backups from a specific location
make list-location LOCATION=/path/to/backups

# Restore a backup
make restore BACKUP=/path/to/backup.tar.gpg
```

## Configuration

Configuration is stored in `.backup.yaml`:

```yaml
# Global exclude patterns
exclude:
  - ".backups/**"
  - "**/.git/**"

target:
  - path: "/path/to/backup/location1"
    maxBackups: 7
    transformation: "s@/source/path@/target/path@"
    namePattern: "%s-%s.tar.xz.gpg"
    # Target-specific excludes
    exclude:
      - "*.tmp"
      - "cache/"
```

## Development

This project follows the [Go Backup Project Code Style Guide](code-style.md). Contributors should ensure their code adheres to these guidelines before submitting changes.

### Testing

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.
