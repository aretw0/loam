package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/core"
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

		// Configure Loam using the new Config struct
		service, err := loam.New(cwd,
			loam.WithGitless(gitless),
			loam.WithLogger(slog.Default()),
			// AutoInit is false by default
		)
		if err != nil {
			fatal("Failed to initialize loam", err)
		}

		// Logic to construct message
		var finalMsg string
		if writeType != "" {
			if writeMsg == "" {
				writeMsg = fmt.Sprintf("update %s", writeID)
			}
			finalMsg = loam.FormatCommitMessage(writeType, writeScope, writeMsg, "")
		} else {
			if writeMsg != "" {
				// Legacy mode
				finalMsg = loam.AppendFooter(writeMsg)
			} else {
				scope := "notes"
				if writeScope != "" {
					scope = writeScope
				}
				finalMsg = loam.FormatCommitMessage(loam.CommitTypeDocs, scope, fmt.Sprintf("update %s", writeID), "")
			}
		}

		// Pass commit message via context (Adapter specific requirement)
		ctx := context.WithValue(context.Background(), core.CommitMessageKey, finalMsg)

		if err := service.SaveNote(ctx, writeID, writeContent, nil); err != nil {
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
