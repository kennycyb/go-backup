package main

import (
	"fmt"
	"os"
)

// Version will be set during build process
var Version = "dev"

func main() {
	// If no arguments provided, print help information
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// TODO: Handle arguments and implement main functionality
}

// printHelp displays usage information to the user
func printHelp() {
	fmt.Println("Go Backup Tool")
	fmt.Println("==============")
	fmt.Println("A simple backup utility written in Go")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go-backup [command] [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  backup    Create a new backup")
	fmt.Println("  restore   Restore from a backup")
	fmt.Println("  list      List available backups")
	fmt.Println("  help      Show this help message")
	fmt.Println("")
	fmt.Println("For more information on a command, run:")
	fmt.Println("  go-backup help [command]")
}
