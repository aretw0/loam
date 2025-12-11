package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/spf13/cobra"
)

var (
	writeID      string
	writeContent string
	changeReason string
	writeType    string
	writeScope   string
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a document",
	Long:  `Create or update a document with the given ID and content. Reads from STDIN if content flag is missing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if writeID == "" {
			fmt.Println("Error: --id is required")
			cmd.Usage()
			os.Exit(1)
		}

		// Handle STDIN if content is empty
		if writeContent == "" {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					fatal("Failed to read from STDIN", err)
				}
				writeContent = string(data)
			}
		}

		if writeContent == "" {
			fmt.Println("Error: --content is required or must be piped via STDIN")
			cmd.Usage()
			os.Exit(1)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		// Configure Loam using the new Config struct
		service, err := loam.New(cwd,
			loam.WithAdapter(adapter),
			loam.WithVersioning(!gitless),
			loam.WithLogger(slog.Default()),
			// AutoInit is false by default
		)
		if err != nil {
			fatal("Failed to initialize loam", err)
		}

		// Logic to construct message
		var finalMsg string
		if writeType != "" {
			if changeReason == "" {
				changeReason = fmt.Sprintf("update %s", writeID)
			}
			finalMsg = loam.FormatChangeReason(writeType, writeScope, changeReason, "")
		} else {
			if changeReason != "" {
				// Legacy mode
				finalMsg = loam.AppendFooter(changeReason)
			} else {
				scope := "documents"
				if writeScope != "" {
					scope = writeScope
				}
				finalMsg = loam.FormatChangeReason(loam.CommitTypeDocs, scope, fmt.Sprintf("update %s", writeID), "")
			}
		}

		// Pass commit message via context (Adapter specific requirement)
		ctx := context.WithValue(context.Background(), core.ChangeReasonKey, finalMsg)

		if err := service.SaveDocument(ctx, writeID, writeContent, nil); err != nil {
			fatal("Failed to save document", err)
		}

		fmt.Printf("Document '%s' saved and committed.\n", writeID)
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVar(&writeID, "id", "", "Document ID (filename)")
	writeCmd.Flags().StringVar(&writeContent, "content", "", "Document content")
	writeCmd.Flags().StringVarP(&changeReason, "message", "m", "", "Change reason (audit note)")
	writeCmd.Flags().StringVarP(&writeType, "type", "t", "", "Change type (feat, fix, etc)")
	writeCmd.Flags().StringVarP(&writeScope, "scope", "s", "", "Commit scope")
	writeCmd.MarkFlagRequired("id")
}
