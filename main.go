package main

import (
	"github.com/kennycyb/go-backup/app/cmd"
)

// Version will be set during build process
var Version = "1.0.0"

func main() {
	// Execute the root command with the version information
	cmd.Execute(Version)
}
