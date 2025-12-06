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

		vault, err := loam.NewVault(cwd, slog.Default())
		if err != nil {
			fatal("Failed to open vault", err)
		}

		note := &loam.Note{
			ID:      writeID,
			Content: writeContent,
		}

		if writeMsg == "" {
			writeMsg = fmt.Sprintf("feat: update note %s", writeID)
		}

		if err := vault.Save(note, writeMsg); err != nil {
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
	writeCmd.MarkFlagRequired("id")
}
