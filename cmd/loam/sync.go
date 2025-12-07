package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/git"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize vault with remote (pull & push)",
	Long: `Synchronize the local vault with the configured remote repository.
It performs a 'git pull --rebase' to integrate remote changes, followed by a 'git push' to upload local changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		logger := slog.Default()

		// Determine path using loam helper (ensures we get the right vault root)
		// We can reuse ResolveVaultPath logic if exposed, or just rely on git client finding root.
		// For consistency, let's use the Factory to "validate" the vault first?
		// No, `sync` shouldn't require instantiating the whole service if we just want to sync.
		// But we need to know if it's gitless.

		// Check gitless flag
		if gitless {
			fmt.Fprintf(os.Stderr, "Error: Cannot sync in gitless mode\n")
			os.Exit(1)
		}

		// Instantiate Git Client directly
		// Note: We use 'cwd' but strictly we should use 'loam.ResolveVaultPath'
		// to be consistent with other commands if user is in a subdir.
		// However, standard git commands work from subdirs.
		gitClient := git.NewClient(cwd, logger)

		if !gitClient.IsRepo() {
			fatal("Not a git repository", fmt.Errorf("%s is not a git repo", cwd))
		}

		fmt.Println("Syncing...")
		if err := gitClient.Sync(); err != nil {
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
