package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/spf13/cobra"
)

var (
	listJSON  bool
	filterTag string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents in the vault",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		root, err := loam.FindVaultRoot(wd)
		if err != nil {
			fmt.Println("Error: Not a Loam vault (no .loam, .git, or loam.json found).")
			os.Exit(1)
		}

		// Auto-detect versioning for list
		useVersioning := !nover
		if useVersioning {
			if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
				useVersioning = false
			}
		}

		service, err := loam.New(root,
			loam.WithAdapter(adapter),
			loam.WithVersioning(useVersioning),
			loam.WithMustExist(true),
			loam.WithLogger(slog.Default()),
		)
		if err != nil {
			fmt.Printf("Error initializing loam: %v\n", err)
			os.Exit(1)
		}

		docs, err := service.ListDocuments(context.Background())
		if err != nil {
			fmt.Printf("Error listing documents: %v\n", err)
			os.Exit(1)
		}

		var filtered []core.Document
		for _, doc := range docs {
			if filterTag != "" {
				// Check tags
				tags, ok := doc.Metadata["tags"]
				hasTag := false
				if ok {
					// Handle []interface{} (from YAML) or []string
					switch t := tags.(type) {
					case []interface{}:
						for _, item := range t {
							if s, ok := item.(string); ok && s == filterTag {
								hasTag = true
								break
							}
						}
					case []string:
						for _, s := range t {
							if s == filterTag {
								hasTag = true
								break
							}
						}
					}
				}
				if !hasTag {
					continue
				}
			}
			filtered = append(filtered, doc)
		}

		if listJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(filtered); err != nil {
				fmt.Printf("Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
			return
		}

		for _, doc := range filtered {
			// Basic output: ID - Title (if available)
			title := ""
			if t, ok := doc.Metadata["title"].(string); ok {
				title = fmt.Sprintf("- %s", t)
			}
			fmt.Printf("%s %s\n", doc.ID, title)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	listCmd.Flags().StringVar(&filterTag, "tag", "", "Filter documents by tag")
}
