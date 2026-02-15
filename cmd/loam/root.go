package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	nover   bool
	adapter string
	strict  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "loam",
	Short: "A Transactional Storage Engine for Content & Metadata",
	Long: `Loam treats your Markdown documents as a NoSQL database.
It orchestrates filesystem writes and version control to ensure transactional integrity.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := slog.LevelInfo
		if verbose {
			level = slog.LevelDebug
		}

		opts := &slog.HandlerOptions{
			Level: level,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		slog.SetDefault(logger)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main().
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&nover, "nover", false, "Run in no-versioning mode (no git operations)")
	rootCmd.PersistentFlags().StringVar(&adapter, "adapter", "fs", "Storage adapter to use (fs)")
	rootCmd.PersistentFlags().BoolVar(&strict, "strict", false, "Enable strict type checking (preserves numeric fidelity)")
}
