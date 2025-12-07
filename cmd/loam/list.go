package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

var (
	listJSON  bool
	filterTag string
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

		service, err := loam.New(wd,
			loam.WithGitless(gitless),
			loam.WithMustExist(true),
			loam.WithLogger(slog.Default()),
		)
		if err != nil {
			fmt.Printf("Error initializing loam: %v\n", err)
			os.Exit(1)
		}

		notes, err := service.ListNotes(context.Background())
		if err != nil {
			fmt.Printf("Error listing notes: %v\n", err)
			os.Exit(1)
		}

		// Filter
		var filtered []core.Note
		for _, note := range notes {
			if filterTag != "" {
				// Check tags
				tags, ok := note.Metadata["tags"]
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
			filtered = append(filtered, note)
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

		for _, note := range filtered {
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
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	listCmd.Flags().StringVar(&filterTag, "tag", "", "Filter notes by tag")
}
