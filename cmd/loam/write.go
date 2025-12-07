package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var (
	writeID      string
	writeContent string
	writeMsg     string
	writeType    string
	writeScope   string
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a note",
	Long:  `Create or update a note with the given ID and content.`,
	Run: func(cmd *cobra.Command, args []string) {
		if writeID == "" {
			fmt.Println("Error: --id is required")
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

		note := &loam.Note{
			ID:      writeID,
			Content: writeContent,
		}

		// Logic to construct message
		var finalMsg string

		// Strategy:
		// 1. If explicit --type is given, use it + message as subject.
		// 2. If NO --type but --message is given, use legacy mode (append footer).
		// 3. If NO --type AND NO --message, auto-generate semantic message (default: chore or docs).

		if writeType != "" {
			if writeMsg == "" {
				// Auto-generate subject if missing?
				writeMsg = fmt.Sprintf("update %s", writeID)
			}
			finalMsg = loam.FormatCommitMessage(writeType, writeScope, writeMsg, "")
		} else {
			if writeMsg != "" {
				// Legacy mode
				finalMsg = loam.AppendFooter(writeMsg)
			} else {
				// Auto mode: Default to 'docs' type
				// "docs(notes): update {id}"
				scope := "notes"
				if writeScope != "" {
					scope = writeScope
				}
				// Default type: docs
				finalMsg = loam.FormatCommitMessage(loam.CommitTypeDocs, scope, fmt.Sprintf("update %s", writeID), "")
			}
		}

		if err := vault.Save(note, finalMsg); err != nil {
			fatal("Failed to save note", err)
		}

		fmt.Printf("Note '%s' saved and committed.\n", writeID)
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVar(&writeID, "id", "", "Note ID (filename)")
	writeCmd.Flags().StringVar(&writeContent, "content", "", "Note content")
	writeCmd.Flags().StringVarP(&writeMsg, "message", "m", "", "Commit message")
	writeCmd.Flags().StringVarP(&writeType, "type", "t", "", "Commit type")
	writeCmd.Flags().StringVarP(&writeScope, "scope", "s", "", "Commit scope")
	writeCmd.MarkFlagRequired("id")
}
