package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var (
	commitMsg string
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes",
	Long:  `Commit staged changes to the underlying Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		if commitMsg == "" {
			fmt.Println("Error: --message is required")
			cmd.Usage()
			os.Exit(1)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		vault, err := loam.NewVault(cwd, slog.Default(), loam.WithGitless(gitless))
		if err != nil {
			fatal("Failed to open vault", err)
		}

		// Access Git directly for manual commit of staged changes
		if err := vault.Git.Commit(commitMsg); err != nil {
			fatal("Failed to commit", err)
		}

		fmt.Println("Committed changes.")
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringVarP(&commitMsg, "message", "m", "", "Commit message")
	commitCmd.MarkFlagRequired("message")
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
