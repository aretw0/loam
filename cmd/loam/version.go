package main

import (
	"fmt"
	"strings"

	"github.com/aretw0/loam"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of loam",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("loam version %s\n", strings.TrimSpace(loam.Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
