package main

import (
	"fmt"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a note from the vault",
	Long:  `Delete permanently removes a note from the vault and stages the deletion in Git.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		v, err := loam.NewVault(wd, nil, loam.WithGitless(gitless), loam.WithMustExist())
		if err != nil {
			fmt.Printf("Error initializing vault: %v\n", err)
			os.Exit(1)
		}

		if err := v.Delete(id); err != nil {
			fmt.Printf("Error deleting note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Note deleted: %s\n", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
