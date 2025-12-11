package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
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

		if nover {
			fatal("Cannot initialize vault in no-versioning mode", fmt.Errorf("git is required for init"))
		}

		// loam init -> AutoInit=true
		_, err = loam.Init(cwd,
			loam.WithAdapter(adapter),
			loam.WithAutoInit(true),
			loam.WithVersioning(!nover),
			loam.WithLogger(slog.Default()),
		)
		if err != nil {
			fatal("Failed to initialize vault", err)
		}

		// Since we removed resolvedPath return, we assume it's CWD or has been handled.
		// For CLI UX just print CWD.
		fmt.Printf("Initialized empty Loam vault in %s\n", cwd)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
