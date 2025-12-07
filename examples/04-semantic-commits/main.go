package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Inicializa um vault temporário para demonstração
	vault, err := loam.NewVault("semantic-demo", logger,
		loam.WithTempDir(),
		loam.WithAutoInit(true),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vault criado em: %s\n", vault.Path)

	// 1. Criar nota
	note := &loam.Note{
		ID:      "design-doc",
		Content: "# System Design\n...",
		Metadata: loam.Metadata{
			"status": "draft",
		},
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
	if err := vault.Save(note, msg); err != nil {
		panic(err)
	}

	fmt.Println("Nota salva com commit semântico!")
}
