# go-backup Requirements Document

## 1. Introduction

### 1.1 Purpose
This document outlines the requirements for the go-backup utility, a command-line tool designed to create, manage, and restore encrypted backups efficiently and securely.

### 1.2 Scope
The go-backup utility is intended for users who need to create secure backups of their data with support for multiple encryption methods, configurable retention policies, and flexible targeting options.

### 1.3 Definitions and Acronyms
- **GPG**: GNU Privacy Guard, an encryption program
- **Age**: A modern file encryption tool that provides a simple, secure method for encrypting files
- **Tar**: A common archive format used to bundle files together
- **Retention Policy**: Rules governing how long backups are kept before being deleted

## 2. Overall Description

### 2.1 Product Perspective
go-backup is a standalone utility that integrates with system commands like GPG and tar to provide backup functionality. It can be used as part of automated backup workflows or manual backup procedures.

### 2.2 Product Functions
- Create encrypted backups of directories
- Support multiple encryption methods (GPG, Age, or none)
- Configure multiple backup destinations
- Implement retention policies for backup management
- Provide flexible file/directory exclusion patterns
- Restore backups to specified locations

### 2.3 User Classes and Characteristics
- **System administrators** who manage backups on servers
- **Individual users** who want to securely back up personal data
- **Developers** who need to back up project directories

### 2.4 Operating Environment
- Linux-based operating systems (primary)
- macOS (secondary)
- Windows with appropriate Unix utilities installed (limited support)

### 2.5 Design and Implementation Constraints
- Written in Go for cross-platform compatibility
- Must integrate with external encryption tools (GPG, Age)
- Must use standard Unix utilities (tar) for archive operations

## 3. Specific Requirements

### 3.1 External Interfaces

#### 3.1.1 Command Line Interface
- The primary interface shall be a command-line utility
- The CLI shall support the following commands:
  - `create`: Create a new backup
  - `list`: List available backups
  - `restore`: Restore from a backup
  - `generate-config`: Generate a default configuration file
  - `version`: Display version information

#### 3.1.2 Configuration File
- The system shall use a YAML configuration file (`.backup.yaml`)
- The configuration file shall support:
  - Global and target-specific exclusion patterns
  - Multiple backup targets
  - Encryption configuration
  - Retention policies (maximum number of backups)

### 3.2 Functional Requirements

#### 3.2.1 Backup Creation
- **REQ-1**: The system shall create compressed tar archives of specified directories
- **REQ-2**: The system shall support excluding files/directories using patterns
- **REQ-3**: The system shall encrypt backups using the configured encryption method
- **REQ-4**: The system shall support backup naming with configurable patterns
- **REQ-5**: The system shall support multiple backup destinations

#### 3.2.2 Encryption
- **REQ-6**: The system shall support GPG encryption with configurable recipients
- **REQ-7**: The system shall support Age encryption with configurable recipients
- **REQ-8**: The system shall support unencrypted backups
- **REQ-9**: The system shall support ASCII armor mode for GPG and Age encryption

#### 3.2.3 Backup Management
- **REQ-10**: The system shall automatically clean up old backups based on retention policy
- **REQ-11**: The system shall provide a way to list available backups
- **REQ-12**: The system shall sort backups by name/date when applying retention policies

#### 3.2.4 Backup Restoration
- **REQ-13**: The system shall support restoring backups to a specified directory
- **REQ-14**: The system shall automatically detect encryption type during restore
- **REQ-15**: The system shall handle decryption of GPG and Age encrypted backups

#### 3.2.5 Configuration Management
- **REQ-34**: The system shall provide a command to generate a default configuration file
- **REQ-35**: The generated configuration shall include sensible defaults for paths, exclusions, and encryption settings
- **REQ-36**: The system shall prompt for confirmation before overwriting an existing configuration file
- **REQ-37**: The system shall support specifying a custom path for the generated configuration file

### 3.3 Non-Functional Requirements

#### 3.3.1 Performance
- **REQ-16**: The system shall efficiently process large backup directories
- **REQ-17**: The system shall minimize disk usage during backup creation

#### 3.3.2 Security
- **REQ-18**: The system shall never store encryption keys or credentials
- **REQ-19**: The system shall use secure algorithms for all encryption operations
- **REQ-20**: The system shall securely handle temporary files used during backup

#### 3.3.3 Usability
- **REQ-21**: The system shall provide clear error messages
- **REQ-22**: The system shall provide informative progress messages during operations
- **REQ-23**: The system shall include comprehensive help documentation

#### 3.3.4 Reliability
- **REQ-24**: The system shall clean up temporary files even in case of failures
- **REQ-25**: The system shall validate configuration files before use

## 4. Configuration Requirements

### 4.1 File Format
- The configuration shall be stored in YAML format
- The configuration file shall be named `.backup.yaml` by default

### 4.2 Configuration Options
- **Encryption settings**:
  ```yaml
  encryption:
    type: "gpg" | "age" | "none"
    options:
      recipients: [list of recipient keys]
      armor: true | false
  ```

- **Target settings**:
  ```yaml
  target:
    - path: "/path/to/destination"
      maxBackups: <number>
      namePattern: "<pattern>"
      exclude: [list of patterns]
  ```

- **Global exclusion settings**:
  ```yaml
  exclude:
    - "<pattern1>"
    - "<pattern2>"
    # etc.
  ```

## 5. Development Requirements

### 5.1 Code Quality
- **REQ-26**: Code shall follow the project's code style guidelines
- **REQ-27**: Code shall include appropriate unit and integration tests
- **REQ-28**: Code shall implement proper error handling

### 5.2 Documentation
- **REQ-29**: Code shall be properly documented with comments
- **REQ-30**: External interfaces shall be documented
- **REQ-31**: User documentation shall be provided for all features

### 5.3 Testing
- **REQ-32**: Unit tests shall be implemented for core functionality
- **REQ-33**: Test coverage shall be maintained at a minimum of 70%

## 6. Constraints and Limitations

- The system relies on external tools (GPG, Age, tar) being properly installed
- The system is primarily designed for Unix-like environments
- The backup file size is limited by available disk space
- Encryption and decryption performance depends on the selected encryption tool

## 7. Roadmap for Future Enhancements

- Support for remote backup destinations (SSH, S3, etc.)
- Incremental backup functionality
- Backup verification features
- GUI interface
- Scheduled backups
- Backup compression options