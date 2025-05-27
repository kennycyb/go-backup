package encrypt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GPGEncrypt encrypts a file using GPG with the specified recipient's public key.
// It returns the path to the encrypted file.
func GPGEncrypt(sourceFile, recipient string) (string, error) {
	// Ensure the source file exists
	if _, err := os.Stat(sourceFile); err != nil {
		return "", fmt.Errorf("source file doesn't exist: %w", err)
	}

	// Create the output file path by appending .gpg extension
	encryptedFile := sourceFile + ".gpg"

	// Build and execute gpg command
	cmd := exec.Command("gpg", "--batch", "--yes", "--trust-model", "always",
		"--recipient", recipient, "--output", encryptedFile,
		"--encrypt", sourceFile)

	// Capture the standard error
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start gpg command: %w", err)
	}

	// Read the error output
	errorOutput := make([]byte, 1024)
	stderr.Read(errorOutput)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("gpg encryption failed: %w, details: %s", err, errorOutput)
	}

	// Verify the encrypted file was created
	if _, err := os.Stat(encryptedFile); err != nil {
		return "", fmt.Errorf("encrypted file wasn't created: %w", err)
	}

	return encryptedFile, nil
}

// GPGDecrypt decrypts a file using GPG.
// It returns the path to the decrypted file.
func GPGDecrypt(encryptedFile, outputFile string) (string, error) {
	// Ensure the encrypted file exists
	if _, err := os.Stat(encryptedFile); err != nil {
		return "", fmt.Errorf("encrypted file doesn't exist: %w", err)
	}

	// If no output file is specified, create one by removing the .gpg extension
	if outputFile == "" {
		outputFile = encryptedFile
		if filepath.Ext(outputFile) == ".gpg" {
			outputFile = outputFile[:len(outputFile)-4]
		} else {
			outputFile = outputFile + ".decrypted"
		}
	}

	// Build and execute gpg command
	cmd := exec.Command("gpg", "--batch", "--yes", "--output", outputFile,
		"--decrypt", encryptedFile)

	// Capture the standard error
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start gpg command: %w", err)
	}

	// Read the error output
	errorOutput := make([]byte, 1024)
	stderr.Read(errorOutput)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("gpg decryption failed: %w, details: %s", err, errorOutput)
	}

	// Verify the decrypted file was created
	if _, err := os.Stat(outputFile); err != nil {
		return "", fmt.Errorf("decrypted file wasn't created: %w", err)
	}

	return outputFile, nil
}
