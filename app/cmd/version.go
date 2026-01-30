package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-backup",
	Long:  `Print the version number of go-backup.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-backup version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
