package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var (
	readJSON bool
)

var readCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Read a note",
	Long:  `Read a note by its ID. Outputs raw markdown content by default, or JSON object with --json.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		// Use slog.Default() if nil was passed before, or just clean up
		v, err := loam.NewVault(wd, slog.Default(), loam.WithGitless(gitless), loam.WithMustExist())
		if err != nil {
			fmt.Printf("Error initializing vault: %v\n", err)
			os.Exit(1)
		}

		note, err := v.Read(id)
		if err != nil {
			// If JSON requested, maybe output empty JSON or error JSON?
			// For now, standard error to stderr.
			fmt.Fprintf(os.Stderr, "Error reading note: %v\n", err)
			os.Exit(1)
		}

		if readJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(note); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Default: Print Content
		fmt.Print(note.Content)
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().BoolVar(&readJSON, "json", false, "Output in JSON format")
}
