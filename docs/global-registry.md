# Global Backup Registry

## Overview

go-backup now supports a global backup registry at `~/.backup.yaml`. This file tracks all backup locations and their last run times across your system.

## How It Works

When you run a backup using `go-backup run`, the tool will:

1. Check if `~/.backup.yaml` exists in your home directory
2. If it exists, update the registry with:
   - The full path to the directory containing the local `.backup.yaml`
   - The timestamp of when the backup was run
3. If `~/.backup.yaml` doesn't exist, the backup runs normally without updating any global registry

## Example

Create a global registry file at `~/.backup.yaml`:

```yaml
default:
  encryption:
    method: gpg
    receiver: user@example.com
backups:
  - location: /Users/john/projects/my-app
    run_at: 2026-01-30T14:20:39.333321+08:00
  - location: /Users/john/documents/important
    run_at: 2026-01-30T15:30:12.123456+08:00
```

## Structure

### `default` Section

Optional default configuration that can be used across all backups:

- `encryption`: Default encryption settings
  - `method`: Encryption method (e.g., `gpg`)
  - `receiver`: Default GPG recipient email

### `backups` Section

Array of backup locations being tracked:

- `location`: Full path to the directory containing the local `.backup.yaml`
- `run_at`: ISO 8601 timestamp of the last backup run

## Usage

### Initial Setup

1. Create `~/.backup.yaml` manually or copy from a template:

   ```bash
   cat > ~/.backup.yaml << 'EOF'
   default:
     encryption:
       method: gpg
       receiver: your-email@example.com
   backups: []
   EOF
   ```

2. Run backups as usual from any location with a local `.backup.yaml`:

   ```bash
   cd /path/to/your/project
   go-backup run
   ```

3. The global registry will be automatically updated

### Viewing Tracked Backups

Simply view the file:

```bash
cat ~/.backup.yaml
```

### Running All Tracked Backups

Use the `run-all` command to run backups for all tracked locations:

```bash
go-backup run-all
```

This command will:

- Read all backup locations from `~/.backup.yaml`
- Execute a backup for each location
- Display errors if a location is missing or if .backup.yaml is not found
- Stop at the first error by default

To continue running backups even if one fails:

```bash
go-backup run-all --continue
```

The command provides a summary at the end showing:

- Number of successful backups
- Number of failed backups
- Number of missing locations
- Total locations processed

### Removing a Backup Location

Edit `~/.backup.yaml` and remove the entry from the `backups` array, or delete the entire file if you don't want global tracking.

## Notes

- The global registry is **optional**. If `~/.backup.yaml` doesn't exist, backups work normally without global tracking
- Each backup location must have its own local `.backup.yaml` configuration file
- The `location` field stores the absolute path to the directory containing the local `.backup.yaml`
- Timestamps use ISO 8601 format with timezone information
- Multiple backup locations can be tracked in a single global registry

## Use Cases

- Track all backup jobs across your system
- Quickly see when each backup location was last backed up
- Maintain consistent encryption settings across multiple backup locations
- Audit backup activity across different projects
- **Run all tracked backups with a single command** using `go-backup run-all`

## Future Enhancements

Potential features for future versions:

- Status summary showing which backups are overdue
- Backup health checks based on last run times
- Scheduling and automation support
