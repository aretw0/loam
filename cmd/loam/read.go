package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/spf13/cobra"
)

var (
	readFormat string
)

var readCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Read a document",
	Long:  `Read a document by its ID. Outputs raw markdown content by default. Use --format=json for structured output.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		// Configure Loam
		service, err := loam.New(wd,
			loam.WithAdapter(adapter),
			loam.WithVersioning(!gitless),
			loam.WithMustExist(true),
			loam.WithLogger(slog.Default()),
		)
		if err != nil {
			fmt.Printf("Error initializing loam: %v\n", err)
			os.Exit(1)
		}

		note, err := service.GetDocument(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading document: %v\n", err)
			os.Exit(1)
		}

		switch readFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(note); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
		case "raw":
			fmt.Print(note.Content)
		default:
			fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'. Use 'raw' or 'json'.\n", readFormat)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVar(&readFormat, "format", "raw", "Output format (raw, json)")
}
