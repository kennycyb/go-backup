package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	cfgFile string

	// Version is set during build
	Version string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-backup",
	Short: "A simple backup utility written in Go",
	Long: `Go Backup Tool
==============
A simple backup utility written in Go that helps you manage
your backup needs easily and efficiently.`,
	Version: Version,
	// If no subcommands or arguments are provided, show help
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	Version = version
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("go-backup version {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-backup.yaml)")

	// Commands are added in their respective files' init() functions
}
