package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	detailed bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long: `List all available backups with their metadata.
This command will display information about existing backups.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing backups...")
		fmt.Printf("Detailed view: %v\n", detailed)
		// TODO: Implement list functionality
	},
}

func init() {
	// Local flags for the list command
	listCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed information")
}
