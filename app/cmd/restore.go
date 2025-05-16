package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	backupFile string
	targetDir  string
	overwrite  bool
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore from a backup",
	Long: `Restore files from a previously created backup.
This command will extract and restore files from a backup archive.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restoring from backup...")
		fmt.Printf("Backup file: %s\n", backupFile)
		fmt.Printf("Target directory: %s\n", targetDir)
		fmt.Printf("Overwrite existing: %v\n", overwrite)
		// TODO: Implement restore functionality
	},
}

func init() {
	// Local flags for the restore command
	restoreCmd.Flags().StringVarP(&backupFile, "file", "f", "", "Backup file to restore from (required)")
	restoreCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory to restore to")
	restoreCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files")

	// Mark required flags
	restoreCmd.MarkFlagRequired("file")
}
