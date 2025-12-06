package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "loam",
	Short: "A Transactional Storage Engine for Markdown + Frontmatter backed by Git",
	Long: `Loam treats your Markdown notes as a NoSQL database.
It orchestrates filesystem writes and Git commits to ensure transactional integrity.`,
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
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}
