package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/spf13/cobra"
)

var (
	writeID      string
	writeContent string
	changeReason string
	writeType    string
	writeScope   string
	writeSet     map[string]string
	writeRaw     bool
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

		if writeContent != "" && writeRaw {
			fmt.Println("Error: cannot use --content with --raw")
			os.Exit(1)
		}

		// Handle STDIN if content is empty (or we are in raw mode)
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

		wd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		root, err := loam.FindVaultRoot(wd)
		if err != nil {
			fatal("Not a Loam vault (no .loam, .git, or loam.json found). Run 'loam init' first.", nil)
		}

		// Configure Loam
		opts := []loam.Option{
			loam.WithAdapter(adapter),
			loam.WithLogger(slog.Default()),
		}

		// Only force versioning config if the flag was explicitly set.
		// Otherwise, let the platform auto-detect (Smart Gitless Detection).
		if cmd.Flags().Lookup("nover").Changed {
			opts = append(opts, loam.WithVersioning(!nover))
		}

		service, err := loam.New(root, opts...)
		if err != nil {
			fatal("Failed to initialize loam", err)
		}

		// Prepare Metadata
		var meta core.Metadata
		if writeRaw {
			// RAW MODE: Parse content transparently
			// We need to know the extension strategy loam uses.
			// Loam logic: ID extension -> if none, try to guess or default .md
			// Here we act as a "smart pipe". We assume the ID extension is authoritative for parsing.
			ext := filepath.Ext(writeID)
			if ext == "" {
				ext = ".md" // Default assumption for raw piping if no extension provided?
			}

			// Parse content using the shared logic
			doc, err := fs.ParseDocument(strings.NewReader(writeContent), ext, "") // We assume default metadata key for now? Or get from config?
			// NOTE: We don't have access to repo config here easily without creating repo.
			// Ideally loam service exposes this, but for now we assume standard behavior.
			if err != nil {
				fatal("Failed to parse raw content", err)
			}
			meta = doc.Metadata
			writeContent = doc.Content // Update content to be the "body" only if parsed
		} else {
			// IMPERATIVE MODE: Use flags
			if len(writeSet) > 0 {
				meta = make(core.Metadata)
				for k, v := range writeSet {
					meta[k] = v
				}
			}
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

		if err := service.SaveDocument(ctx, writeID, writeContent, meta); err != nil {
			fatal("Failed to save document", err)
		}

		fmt.Printf("Document '%s' saved.\n", writeID)
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVar(&writeID, "id", "", "Document ID (filename)")
	writeCmd.Flags().StringVar(&writeContent, "content", "", "Document content")
	writeCmd.Flags().StringVarP(&changeReason, "message", "m", "", "Change reason (audit note)")
	writeCmd.Flags().StringVarP(&writeType, "type", "t", "", "Change type (feat, fix, etc)")
	writeCmd.Flags().StringVarP(&writeScope, "scope", "s", "", "Commit scope")
	writeCmd.Flags().StringToStringVar(&writeSet, "set", nil, "Set metadata fields (key=value)")
	writeCmd.Flags().BoolVar(&writeRaw, "raw", false, "Treat input as raw document (parse metadata from content)")
	writeCmd.MarkFlagRequired("id")
}
