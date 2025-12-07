package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Inicializa um vault temporário para demonstração
	cfg := loam.Config{
		Path:      "semantic-demo",
		Logger:    logger,
		ForceTemp: true,
		AutoInit:  true,
	}

	// Cleanup
	safePath := loam.ResolveVaultPath(cfg.Path, true)
	os.RemoveAll(safePath)

	service, err := loam.New(cfg)
	if err != nil {
		panic(err)
	}

	// Git Setup
	gitClient := git.NewClient(safePath, logger)
	gitClient.Run("config", "user.name", "Example Bot")
	gitClient.Run("config", "user.email", "bot@example.com")

	// 1. Definição da Nota
	noteID := "design-doc"
	content := "# System Design\n..."
	meta := core.Metadata{
		"status": "draft",
	}

	// 2. Formatar mensagem semântica usando o helper do Loam
	// Isso garante que o commit siga o padrão Conventional Commits + Footer
	msg := loam.FormatCommitMessage(
		loam.CommitTypeDocs,         // type: docs
		"arch",                      // scope: arch
		"create initial design doc", // subject
		"This document outlines the core architecture.", // body
	)

	fmt.Printf("\nMensagem Gerada:\n---\n%s\n---\n", msg)

	// 3. Salvar (Commit)
	ctx := context.WithValue(context.Background(), core.CommitMessageKey, msg)
	if err := service.SaveNote(ctx, noteID, content, meta); err != nil {
		panic(err)
	}

	fmt.Println("Nota salva com commit semântico!")
}
