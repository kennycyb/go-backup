package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	compressionService "github.com/kennycyb/go-backup/internal/service/compress"
	configService "github.com/kennycyb/go-backup/internal/service/config"
	"github.com/spf13/cobra"
)

var (
	largeMinSize int64
	largeSort    string
	largeLimit   int
)

// largeFilesCmd represents the large-files command
var largeFilesCmd = &cobra.Command{
	Use:   "large-files",
	Short: "List large files that may cause backup issues",
	Long: `List large files in the backup source directory that may exceed size limits.
This command helps identify files that could cause issues when creating tar archives.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Color and emoji constants
		const (
			ColorReset  = "\033[0m"
			ColorRed    = "\033[31m"
			ColorGreen  = "\033[32m"
			ColorYellow = "\033[33m"
			ColorBlue   = "\033[34m"
			ColorCyan   = "\033[36m"
			ColorWhite  = "\033[37m"
			ColorBold   = "\033[1m"
			ColorDim    = "\033[2m"
		)

		// Check if source is specified
		if source == "" {
			fmt.Printf("%s%sError: Source directory not specified%s\n", ColorRed, ColorBold, ColorReset)
			fmt.Printf("Use the --source flag to specify a directory\n")
			os.Exit(1)
		}

		// Validate source directory exists
		sourceStat, err := os.Stat(source)
		if err != nil {
			fmt.Printf("%s%sError: Unable to access source directory %s: %v%s\n", ColorRed, ColorBold, source, err, ColorReset)
			os.Exit(1)
		}

		if !sourceStat.IsDir() {
			fmt.Printf("%s%sError: %s is not a directory%s\n", ColorRed, ColorBold, source, ColorReset)
			os.Exit(1)
		}

		// Determine which config file to use
		var configPath string
		if configFile != "" {
			configPath = configFile
		} else {
			configPath = filepath.Join(source, ".backup.yaml")
		}

		// Get excludes from config if available
		configExcludes := []string{}

		config, configErr := configService.ReadBackupConfig(configPath)
		if configErr == nil && len(config.Excludes) > 0 {
			configExcludes = config.Excludes
			fmt.Printf("%sUsing excludes from config:%s %v\n", ColorDim, ColorReset, configExcludes)
		} else {
			configExcludes = excludeDirs
			fmt.Printf("%sUsing default excludes:%s %v\n", ColorDim, ColorReset, excludeDirs)
		}

		// Create absolute source path
		absSource, err := filepath.Abs(source)
		if err != nil {
			fmt.Printf("%s%sError: Unable to determine absolute path: %v%s\n", ColorRed, ColorBold, err, ColorReset)
			os.Exit(1)
		}

		fmt.Printf("%sAnalyzing files in %s...%s\n", ColorDim, absSource, ColorReset)

		// Find large files
		largeFiles, err := compressionService.ListLargeFiles(absSource, configExcludes, largeMinSize)
		if err != nil {
			fmt.Printf("%s%sError analyzing files: %v%s\n", ColorRed, ColorBold, err, ColorReset)
			os.Exit(1)
		}

		if len(largeFiles) == 0 {
			fmt.Printf("%s%sNo files found larger than %d MB%s\n", ColorGreen, ColorBold, largeMinSize, ColorReset)
			return
		}

		// Limit number of files to display
		if largeLimit > 0 && largeLimit < len(largeFiles) {
			largeFiles = largeFiles[:largeLimit]
		}

		// Print results
		fmt.Printf("%s%sFound %d files larger than %d MB%s\n", ColorYellow, ColorBold, len(largeFiles), largeMinSize, ColorReset)

		// Setup tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "%sSize\tFile\tLast Modified%s\n", ColorBold, ColorReset)
		fmt.Fprintf(w, "%s---\t----\t-------------%s\n", ColorDim, ColorReset) // Display large files
		warnSize := int64(compressionService.RecommendedMaxFileSize)

		for _, file := range largeFiles {
			sizeColor := ColorWhite
			if file.Size > warnSize {
				sizeColor = ColorRed
			} else if file.Size > warnSize/2 {
				sizeColor = ColorYellow
			}

			fmt.Fprintf(w, "%s%s%s\t%s\t%s\n",
				sizeColor,
				file.SizeHuman,
				ColorReset,
				file.RelativePath,
				file.ModTime.Format("Jan 02, 2006 15:04"))
		}
		w.Flush()

		// Print warning for files that may cause issues
		criticalFiles := 0
		for _, file := range largeFiles {
			if file.Size > warnSize {
				criticalFiles++
			}
		}

		if criticalFiles > 0 {
			fmt.Printf("\n%s%s⚠️ Warning: %d file(s) exceed the recommended size limit for tar archives%s\n",
				ColorRed, ColorBold, criticalFiles, ColorReset)
			fmt.Printf("%sFiles over %.2f GB may cause 'write too long' errors during backup.%s\n",
				ColorDim, float64(warnSize)/(1024*1024*1024), ColorReset)
			fmt.Printf("%sConsider adding these files to your exclude list or using a different backup method for them.%s\n",
				ColorDim, ColorReset)
		}
	},
}

func init() {
	rootCmd.AddCommand(largeFilesCmd)

	// Add flags specific to the large-files command
	largeFilesCmd.Flags().Int64Var(&largeMinSize, "min-size", 100, "Minimum size in MB to include in the list")
	largeFilesCmd.Flags().StringVar(&largeSort, "sort", "size", "Sort results by: size, name, or date")
	largeFilesCmd.Flags().IntVar(&largeLimit, "limit", 50, "Limit the number of files to display (0 for no limit)")

	// Add common flags
	largeFilesCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory to analyze")
	largeFilesCmd.Flags().StringVarP(&configFile, "config", "f", "", "Path to configuration file (defaults to .backup.yaml in source directory)")
	largeFilesCmd.Flags().StringSliceVarP(&excludeDirs, "exclude", "e", []string{".git", "node_modules", "tmp", "temp", "logs"}, "Directories to exclude")
}
