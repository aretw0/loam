package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var (
	commitMsg   string
	commitType  string
	commitScope string
	commitBody  string
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes",
	Long:  `Commit staged changes to the underlying Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		vault, err := loam.NewVault(cwd, slog.Default(), loam.WithGitless(gitless))
		if err != nil {
			fatal("Failed to open vault", err)
		}

		if vault.IsGitless() {
			fatal("Cannot commit in gitless mode", fmt.Errorf("git is required"))
		}

		// Logic to construct message
		var finalMsg string

		// If --type is provided, use semantic formatting
		if commitType != "" {
			// If -m is provided, it's the subject. If not, error?
			// Subject is required.
			if commitMsg == "" {
				fmt.Println("Error: --message (subject) is required when using --type")
				cmd.Usage()
				os.Exit(1)
			}
			finalMsg = loam.FormatCommitMessage(commitType, commitScope, commitMsg, commitBody)
		} else {
			// Legacy/Free-form mode
			if commitMsg == "" {
				fmt.Println("Error: --message is required")
				cmd.Usage()
				os.Exit(1)
			}
			finalMsg = loam.AppendFooter(commitMsg)
		}

		// Access Git directly for manual commit of staged changes
		if err := vault.Git.Commit(finalMsg); err != nil {
			fatal("Failed to commit", err)
		}

		fmt.Println("Committed changes.")
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringVarP(&commitMsg, "message", "m", "", "Commit message (Subject)")
	commitCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	commitCmd.Flags().StringVarP(&commitScope, "scope", "s", "", "Commit scope")
	commitCmd.Flags().StringVarP(&commitBody, "body", "b", "", "Commit body")
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
