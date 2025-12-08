package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aretw0/loam"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a note from the vault",
	Long:  `Delete permanently removes a note from the vault and stages the deletion in Git.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		service, err := loam.New(wd, loam.WithAdapter(adapter), loam.WithVersioning(!gitless), loam.WithMustExist(true))
		if err != nil {
			fmt.Printf("Error initializing loam: %v\n", err)
			os.Exit(1)
		}

		if err := service.DeleteNote(context.Background(), id); err != nil {
			fmt.Printf("Error deleting note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Note deleted: %s\n", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
