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

# Optional settings
options:
  git:
    enable: true   # Only run backup when uncommitted changes are detected
    branch: main   # Branch name for auto-pull feature
    pull: auto     # Enable automatic git pull before backup
```

### Smart Backup with Git Integration

The `options.git` settings allow you to run backups conditionally based on git status:

#### Basic Git Integration

- When `options.git.enable: true`: The backup will only run if there are uncommitted changes in the git repository
- If no uncommitted changes are detected, the backup is skipped with a message
- If the directory is not a git repository, a warning is shown and the backup proceeds normally
- This is useful for automated backups where you only want to backup when there's new work

#### Auto-Pull Feature (Smart Backup 2)

When both `branch` and `pull: auto` are configured, the system will automatically pull the latest changes before running the backup:

```yaml
options:
  git:
    enable: true
    branch: main    # or master, develop, etc.
    pull: auto      # enables auto-pull feature
```

**Behavior:**

1. **Branch Check**: Verifies you're on the configured branch
   - If on the configured branch: proceeds with auto-pull
   - If on a different branch: skips auto-pull, but continues to check for uncommitted changes

2. **Auto-Pull**: Automatically runs `git pull` to fetch latest changes from remote

3. **Backup Decision**:
   - ✅ **Runs backup** if:
     - Uncommitted changes exist, OR
     - Pull brought new updates from remote
   - ⏭️ **Skips backup** if:
     - No uncommitted changes, AND
     - No updates from pull (already up-to-date)

**Important Notes:**
- Auto-pull only works when you're on the configured branch
- If you're on a different branch, auto-pull is skipped but the system still checks for uncommitted changes
- The repository must have a remote configured and SSH keys or credential helpers set up for authentication
- This feature is backward compatible: without `pull: auto`, the original behavior is preserved

Example use cases:
- Automated backups that run on a schedule but only capture when you've made changes
- Development workflows where you want to backup uncommitted work
- Integration with git hooks to backup before certain git operations
- Continuous backup systems that pull latest changes and backup only when repository is updated

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
