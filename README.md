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

# List backups for current directory
make list

# List all backups regardless of source
make list-all

# List backups with detailed information
make list-detailed

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

## Commands

### List Command

The `list` command displays backups organized by location and source:

```bash
# Basic usage - shows backups for current directory only
go-backup list

# List all backups regardless of source
go-backup list --all

# Show detailed information
go-backup list --detailed

# List backups in a specific location
go-backup list --path /path/to/backups
```

The list command shows:
- All configured backup locations
- Backups grouped by source within each location
- By default, only shows backups from the current directory
- File size and creation time information
- Up to 5 most recent backups per source (use --detailed to see all)

### Other Commands

Other available commands include:

```bash
# Run a backup
go-backup run

# Initialize a configuration file
go-backup init

# Restore a backup (not fully implemented yet)
go-backup restore
```

## Development

This project follows the [Go Backup Project Code Style Guide](code-style.md). Contributors should ensure their code adheres to these guidelines before submitting changes.

### Testing

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.
