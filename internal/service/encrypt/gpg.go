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
// If a passphrase is provided, it will be used for decryption.
// If passphrase is empty, GPG will use the agent or prompt for a passphrase.
func GPGDecrypt(encryptedFile, outputFile string, passphrase string) (string, error) {
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

	var cmd *exec.Cmd

	if passphrase != "" {
		// Use passphrase-fd=0 to read the passphrase from stdin
		cmd = exec.Command("gpg", "--batch", "--yes", "--passphrase-fd", "0",
			"--output", outputFile, "--decrypt", encryptedFile)

		// Create a pipe to send the passphrase
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("failed to get stdin pipe: %w", err)
		}

		// Capture the standard error
		stderr, err := cmd.StderrPipe()
		if err != nil {
			stdin.Close()
			return "", fmt.Errorf("failed to get stderr pipe: %w", err)
		}

		// Start the command before writing to stdin
		if err := cmd.Start(); err != nil {
			stdin.Close()
			return "", fmt.Errorf("failed to start gpg command: %w", err)
		}

		// Write the passphrase to stdin and close the pipe
		_, err = stdin.Write([]byte(passphrase + "\n"))
		if err != nil {
			return "", fmt.Errorf("failed to write passphrase: %w", err)
		}
		stdin.Close()

		// Read the error output
		errorOutput := make([]byte, 1024)
		stderr.Read(errorOutput)

		// Wait for the command to finish
		if err := cmd.Wait(); err != nil {
			return "", fmt.Errorf("gpg decryption failed: %w, details: %s", err, errorOutput)
		}
	} else {
		// Default command without passphrase support
		cmd = exec.Command("gpg", "--batch", "--yes", "--output", outputFile,
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
	}

	// Verify the decrypted file was created
	if _, err := os.Stat(outputFile); err != nil {
		return "", fmt.Errorf("decrypted file wasn't created: %w", err)
	}

	return outputFile, nil
}
