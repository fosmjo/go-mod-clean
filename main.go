package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// flag variables
var (
	modfilePaths []string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "go-mod-cleaner",
	Short: "Clean up unused Go modules.",
	Long: `Clean up unused Go modules. To be specific, it cleans up all modules within $GOPATH/pkg/mod,
except for currently used modules. To specify the modules in use, you need to indicate them
via go.mod files or directories that contain go.mod files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modCachePath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
		cleaner := NewCleaner(modCachePath, modfilePaths, verbose)
		return cleaner.Clean()
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(
		&modfilePaths, "modfile", "m", nil,
		`go.mod files or directories that contain go.mod files,
modules referenced by these files are considered in use`,
	)
	err := rootCmd.MarkFlagRequired("modfile")
	if err != nil {
		panic(err)
	}

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
}
