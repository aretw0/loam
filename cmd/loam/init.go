package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a loam vault (git init)",
	Long:  `Initialize a new Loam vault in the current directory. This effectively runs 'git init'.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get CWD", err)
		}

		if gitless {
			fatal("Cannot initialize vault in gitless mode", fmt.Errorf("git is required for init"))
		}

		// loam init -> WithAutoInit(true)
		_, err = loam.NewVault(cwd, slog.Default(), loam.WithAutoInit(true), loam.WithGitless(gitless))
		if err != nil {
			fatal("Failed to initialize vault", err)
		}

		fmt.Println("Initialized empty Loam vault in", cwd)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
