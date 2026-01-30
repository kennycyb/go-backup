# Backup Status Tracking Requirements

## 1. Overview

The goal is to persist the execution status of backup jobs in the configuration file (`~/.backup.yaml`) and display this status to the user. This allows users to quickly verify if their scheduled backups are running successfully without inspecting logs.

## 2. User Stories

- **As a user**, I want the application to record the result (success/failure) of each backup run.
- **As a user**, I want to see a summary of all backup targets and their last run status when I execute `go-backup list --all`.

## 3. Functional Requirements

### 3.1 Configuration Schema Update

The `BackupTarget` structure in the configuration file must be updated to include a `lastRun` field.

**New Schema:**

```yaml
target:
  - path: /path/to/backup
    # ... existing fields ...
    lastRun:
      timestamp: 2023-10-27T10:00:00Z
      status: "Success" # or "Failure"
      message: "Backup completed successfully" # or error message
```

### 3.2 Update `run` Command

- When the `run` command executes, it iterates through defined targets.
- After attempting a backup for a target, the application **MUST** update the in-memory configuration with the result.
- The application **MUST** save the updated configuration back to `.backup.yaml`.
  - **Note:** Care must be taken to preserve existing comments and structure if possible, though standard YAML marshalers might not support this perfectly. The user has accepted this trade-off by requesting the feature.
- **Fields to record:**
  - `timestamp`: Time of completion.
  - `status`: "Success" if no errors occurred, "Failure" otherwise.
  - `message`: A brief description or error message.

### 3.3 Update `list` Command

- Modify `go-backup list --all` to display the status.
- **Output Format:**
  - Should show a table or list including: Target Path/File, Last Run Time, Status.
  - Use colors (Green for Success, Red for Failure) if possible.

## 4. Technical Constraints

- Use `gopkg.in/yaml.v3` for YAML handling.
- Ensure concurrent writes don't corrupt the config (though `run` is likely sequential).
- Handle file permission errors gracefully when writing config.

## 5. Acceptance Criteria

1.  Running `go-backup run` updates `.backup.yaml` with `lastRun` information.
2.  Running `go-backup list --all` shows the status from the config file.
3.  If a backup fails, the status shows "Failure" and the error message is saved.
