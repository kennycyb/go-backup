# GitHub Copilot Instructions for Go-Backup

## Terminal Usage
Use @terminal when answering questions about Git.

## General Guidelines
- Follow Go best practices and idiomatic Go code style
- Keep the codebase maintainable and testable
- Provide meaningful comments but avoid unnecessary ones

## Project-Specific Instructions
- **DO NOT** overwrite `.backup.yaml` files when generating code unless explicitly asked to with `--overwrite` flag
- **ALWAYS** use `/tmp/.backup.yaml` for self-testing and examples, never the actual `.backup.yaml` in the project directory
- Preserve the encryption configuration structure in the form:
  ```yaml
  encryption:
    method: gpg
    receiver: user@example.com
  ```
  Not as an array of encryption methods
- When suggesting encryption implementations, remember that passphrase is only applicable for decryption, not encryption
- Always make sure config files are copied alongside backups with appropriate names
- When rotating backups, associated config files should also be cleaned up

## Code Style
- Function names: use camelCase for private functions and PascalCase for exported ones
- Error handling: always check errors and provide meaningful error messages
- Comments: follow Go comment style for packages, types, functions, and constants
- Formatting: follow standard Go formatting with gofmt/goimports

## Testing Guidelines
- Write unit tests for new functionality
- Mock external dependencies (like file system, GPG operations) for testing
- Test both success and error scenarios
- Use table-driven tests where appropriate

## Security Considerations
- Never hardcode sensitive information like encryption keys or passphrases
- When storing passphrases in config files, always include a warning about security implications
- Validate user input, especially paths and commands that will be executed
- Handle temporary files securely, ensuring they are properly cleaned up