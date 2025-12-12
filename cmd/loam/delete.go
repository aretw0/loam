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
	Short: "Delete a document from the vault",
	Long:  `Delete permanently removes a document from the vault and stages the deletion in Git.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		root, err := loam.FindVaultRoot(wd)
		if err != nil {
			fmt.Printf("Error: Not a Loam vault: %v\n", err)
			os.Exit(1)
		}

		service, err := loam.New(root, loam.WithAdapter(adapter), loam.WithVersioning(!nover), loam.WithMustExist(true))
		if err != nil {
			fmt.Printf("Error initializing loam: %v\n", err)
			os.Exit(1)
		}

		if err := service.DeleteDocument(context.Background(), id); err != nil {
			fmt.Printf("Error deleting document: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Document deleted: %s\n", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
