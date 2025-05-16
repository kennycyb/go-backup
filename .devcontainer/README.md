# Development Container for go-backup

This folder contains configuration for a development container that provides a consistent development environment for the go-backup project.

## Features

- Uses the Microsoft Go devcontainer base image
- Includes tools required for Go development:
  - golangci-lint for code linting
  - goimports for import management
  - Ginkgo and Gomega for BDD-style testing
- Pre-configured VS Code settings and extensions
- Automatic go module setup

## Usage

1. Open this repository in VS Code
2. Install the "Remote - Containers" extension if not already installed
3. Run the "Remote-Containers: Reopen in Container" command
4. Wait for the container to build and start

## Additional Tools

The following tools are pre-installed in the container:
- make
- curl
- git
- vim
- Node.js and npm (for any potential web interface)

## VS Code Extensions

The following extensions are automatically installed:
- Go extension
- Makefile Tools
- Trailing Spaces
- Code Spell Checker
