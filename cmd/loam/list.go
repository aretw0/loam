package main

import (
	"fmt"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes in the vault",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		v, err := loam.NewVault(wd, nil)
		if err != nil {
			fmt.Printf("Error initializing vault: %v\n", err)
			os.Exit(1)
		}

		notes, err := v.List()
		if err != nil {
			fmt.Printf("Error listing notes: %v\n", err)
			os.Exit(1)
		}

		for _, note := range notes {
			// Basic output: ID - Title (if available)
			title := ""
			if t, ok := note.Metadata["title"].(string); ok {
				title = fmt.Sprintf("- %s", t)
			}
			fmt.Printf("%s %s\n", note.ID, title)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
