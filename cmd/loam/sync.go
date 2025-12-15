package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize vault with remote",
	Long: `Synchronize the local vault with the configured remote repository.
It integrates remote changes and pushes local changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		fmt.Println("Syncing...")
		if err := loam.Sync(cwd,
			loam.WithAdapter(adapter),
			loam.WithVersioning(!nover),
			loam.WithLogger(slog.Default()),
		); err != nil {
			// User friendly error handling
			fmt.Fprintf(os.Stderr, "Error: Sync failed: %v\n", err)
			fmt.Println("Tip: Ensure you have a remote configured ('git remote add origin <url>') and you are online.")
			fmt.Println("If there are merge conflicts, you may need to resolve them manually in the repository.")
			os.Exit(1)
		}

		fmt.Println("Sync completed successfully.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
