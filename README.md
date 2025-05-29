# go-backup

A command line tool for creating, managing, and restoring backups.

## Features

- Encrypted backups using GPG
- Multiple backup targets
- Configurable retention policies
- File/directory exclusion patterns
- Automated cleanup of old backups
- Backup history tracking in configuration file

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
    # Backup history is automatically updated
    backups:
      - filename: "backup-20250520-123045.tar.gz"
        source: "/path/to/source"
        createdAt: "2025-05-20T12:30:45Z"
        size: 1048576
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

# View backup history from the configuration file
go-backup list --history
```

The list command shows:
- All configured backup locations
- Backups grouped by source within each location
- By default, only shows backups from the current directory
- File size and creation time information
- Up to 5 most recent backups per source (use --detailed to see all)
- With --history flag, shows the backup records stored in the config file

### Other Commands

Other available commands include:

```bash
# Run a backup
go-backup run
```

The run command:
- Creates a compressed backup of the specified source directory
- Copies the backup to all configured targets
- Performs backup rotation based on maxBackups setting
- Updates the backup history in the configuration file

# Initialize a configuration file
go-backup init

# Restore a backup (not fully implemented yet)
go-backup restore
```

### Config Command

The `config` command allows you to modify your `.backup.yaml` file from the command line:

```bash
# Add a new backup target
go-backup config --add-target /path/to/backup/location

# Delete a backup target
go-backup config --delete-target /path/to/backup/location

# Enable GPG encryption for backups
go-backup config --enable-encryption --gpg-receiver user@example.com

# Disable encryption
go-backup config --disable-encryption
```

Options:
- `--add-target <path>`: Add a new backup target by path
- `--delete-target <path>`: Remove a backup target by path
- `--enable-encryption`: Enable GPG encryption for backups (requires `--gpg-receiver`)
- `--disable-encryption`: Disable encryption for backups
- `--gpg-receiver <email>`: Specify the GPG recipient email for encryption

## Development

This project follows the [Go Backup Project Code Style Guide](code-style.md). Contributors should ensure their code adheres to these guidelines before submitting changes.

### Testing

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.
