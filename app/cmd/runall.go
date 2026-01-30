package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var continueOnError bool

// runAllCmd represents the run-all command
var runAllCmd = &cobra.Command{
	Use:   "run-all",
	Short: "Run backups for all locations in global registry",
	Long: `Run backups for all locations tracked in ~/.backup.yaml.
This command reads the global registry and executes backups for each
tracked location. If a location no longer exists, an error is displayed.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Color constants
		const (
			ColorReset  = "\033[0m"
			ColorRed    = "\033[31m"
			ColorGreen  = "\033[32m"
			ColorYellow = "\033[33m"
			ColorCyan   = "\033[36m"
			ColorBold   = "\033[1m"
			ColorDim    = "\033[2m"
		)

		fmt.Printf("%s%s\n======================================\n   📦  Running All Tracked Backups   \n======================================%s\n\n", ColorCyan, ColorBold, ColorReset)

		// Read global registry
		registry, err := configService.ReadGlobalRegistry()
		if err != nil {
			fmt.Printf("%s%s❌ Error:%s %v\n", ColorRed, ColorBold, ColorReset, err)
			fmt.Printf("%sHint:%s Create ~/.backup.yaml to track backup locations.\n", ColorDim, ColorReset)
			fmt.Printf("%sSee docs/global-registry.md for more information.%s\n", ColorDim, ColorReset)
			os.Exit(1)
		}

		if len(registry.Backups) == 0 {
			fmt.Printf("%s%s⚠️  No backup locations found in global registry.%s\n", ColorYellow, ColorBold, ColorReset)
			fmt.Printf("%sRun backups in directories with .backup.yaml to register them.%s\n", ColorDim, ColorReset)
			return
		}

		fmt.Printf("%sFound %d backup location(s) in registry:%s\n\n", ColorDim, len(registry.Backups), ColorReset)

		successCount := 0
		errorCount := 0
		missingCount := 0

		for i, entry := range registry.Backups {
			fmt.Printf("%s[%d/%d]%s %s\n", ColorBold, i+1, len(registry.Backups), ColorReset, entry.Location)

			// Check if location exists
			if _, err := os.Stat(entry.Location); os.IsNotExist(err) {
				fmt.Printf("  %s%s❌ Error:%s Directory does not exist\n", ColorRed, ColorBold, ColorReset)
				missingCount++
				if !continueOnError {
					fmt.Printf("\n%s%s⚠️  Stopping due to error. Use --continue to skip errors.%s\n", ColorYellow, ColorBold, ColorReset)
					break
				}
				fmt.Println()
				continue
			}

			// Check if .backup.yaml exists in the location
			configPath := filepath.Join(entry.Location, ".backup.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Printf("  %s%s❌ Error:%s .backup.yaml not found in directory\n", ColorRed, ColorBold, ColorReset)
				missingCount++
				if !continueOnError {
					fmt.Printf("\n%s%s⚠️  Stopping due to error. Use --continue to skip errors.%s\n", ColorYellow, ColorBold, ColorReset)
					break
				}
				fmt.Println()
				continue
			}

			// Get the path to the current executable
			execPath, err := os.Executable()
			if err != nil {
				// Fall back to "go-backup" if we can't determine the executable path
				execPath = "go-backup"
			}

			// Run backup for this location
			backupCmd := exec.Command(execPath, "run", "-s", entry.Location, "-f", configPath, "--force")
			backupCmd.Stdout = os.Stdout
			backupCmd.Stderr = os.Stderr

			err = backupCmd.Run()
			if err != nil {
				fmt.Printf("  %s%s❌ Error:%s Backup failed: %v\n", ColorRed, ColorBold, ColorReset, err)
				errorCount++
				if !continueOnError {
					fmt.Printf("\n%s%s⚠️  Stopping due to error. Use --continue to skip errors.%s\n", ColorYellow, ColorBold, ColorReset)
					break
				}
			} else {
				successCount++
			}

			fmt.Println()
		}

		// Summary
		fmt.Printf("%s%s======================================\n", ColorCyan, ColorBold)
		fmt.Printf("             Summary\n")
		fmt.Printf("======================================%s\n", ColorReset)
		fmt.Printf("%s✅ Successful:%s %d\n", ColorGreen, ColorReset, successCount)
		if errorCount > 0 {
			fmt.Printf("%s❌ Failed:%s %d\n", ColorRed, ColorReset, errorCount)
		}
		if missingCount > 0 {
			fmt.Printf("%s⚠️  Missing:%s %d\n", ColorYellow, ColorReset, missingCount)
		}
		fmt.Printf("%s📊 Total:%s %d\n", ColorDim, ColorReset, len(registry.Backups))

		if errorCount > 0 || missingCount > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	runAllCmd.Flags().BoolVar(&continueOnError, "continue", false, "Continue running backups even if one fails")
	rootCmd.AddCommand(runAllCmd)
}
