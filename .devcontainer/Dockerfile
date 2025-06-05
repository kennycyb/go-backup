# Dockerfile for go-backup development container
#
# This Dockerfile sets up a development environment for the go-backup project.
# It includes security tools (Trivy, Gitleaks) and a Go development environment.

# Stage 1: Trivy for vulnerability scanning
FROM aquasec/trivy:0.57.1 AS trivy

# Stage 2: Gitleaks for secret scanning
FROM zricethezav/gitleaks:v8.18.4 AS gitleaks

# Stage 3: Main development environment based on Microsoft's Go devcontainer
FROM mcr.microsoft.com/devcontainers/go:1.24

# Copy Trivy and Gitleaks binaries into the devcontainer
COPY --from=trivy /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=trivy /contrib/*.tpl /contrib/
COPY --from=gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks

# Add a simple healthcheck (checks Go version as a proxy for container health)
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 CMD go version || exit 1

# Use the vscode user (default in devcontainers)
USER vscode
