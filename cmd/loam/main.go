package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aretw0/loam/pkg/git"
	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		handleInit(os.Args[2:])
	case "write":
		handleWrite(os.Args[2:])
	case "commit":
		handleCommit(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: loam <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  init             Initialize a loam vault (git init)")
	fmt.Println("  write            Write a note")
	fmt.Println("    -id <id>       Note ID (filename)")
	fmt.Println("    -content <txt> Note content")
	fmt.Println("  commit           Commit changes")
	fmt.Println("    -m <msg>       Commit message")
}

func handleInit(args []string) {
	// For init, we just want to run git init in the current directory
	cwd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get CWD", err)
	}

	client := git.NewClient(cwd)
	if err := client.Init(); err != nil {
		fatal("Failed to init git", err)
	}
	fmt.Println("Initialized empty Loam vault in", cwd)
}

func handleWrite(args []string) {
	cmd := flag.NewFlagSet("write", flag.ExitOnError)
	id := cmd.String("id", "", "Note ID")
	content := cmd.String("content", "", "Note Content")

	if err := cmd.Parse(args); err != nil {
		fatal("Failed to parse args", err)
	}

	if *id == "" {
		fmt.Println("Error: -id is required")
		cmd.Usage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get CWD", err)
	}

	vault, err := loam.NewVault(cwd)
	if err != nil {
		fatal("Failed to open vault", err)
	}

	note := &loam.Note{
		ID:      *id,
		Content: *content,
		// Metadata: empty for CLI simple write for now
	}

	if err := vault.Write(note); err != nil {
		fatal("Failed to write note", err)
	}

	fmt.Printf("Note '%s' written and staged.\n", *id)
}

func handleCommit(args []string) {
	cmd := flag.NewFlagSet("commit", flag.ExitOnError)
	msg := cmd.String("m", "", "Commit message")

	if err := cmd.Parse(args); err != nil {
		fatal("Failed to parse args", err)
	}

	if *msg == "" {
		fmt.Println("Error: -m is required")
		cmd.Usage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get CWD", err)
	}

	vault, err := loam.NewVault(cwd)
	if err != nil {
		fatal("Failed to open vault", err)
	}

	if err := vault.Commit(*msg); err != nil {
		fatal("Failed to commit", err)
	}

	fmt.Println("Committed changes.")
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
