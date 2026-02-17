package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	vaultName := "config-loading-demo"
	vaultPath := loam.ResolveVaultPath(vaultName, true)
	_ = os.RemoveAll(vaultPath)

	service, err := loam.New(vaultName,
		loam.WithLogger(logger),
		loam.WithForceTemp(true),
		loam.WithAutoInit(true),
		loam.WithVersioning(false),
		loam.WithContentExtraction(false),
		loam.WithMarkdownBodyKey("body"),
	)
	if err != nil {
		panic(err)
	}

	mdPath := filepath.Join(vaultPath, "config.md")
	mdBody := "---\nname: example\nenabled: true\n---\nThis is the body text.\n"
	if err := os.WriteFile(mdPath, []byte(mdBody), 0644); err != nil {
		panic(err)
	}

	doc, err := service.GetDocument(context.TODO(), "config")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content:")
	fmt.Printf("%q\n", doc.Content)
	fmt.Println("Metadata:")
	fmt.Printf("name=%v enabled=%v body=%q\n", doc.Metadata["name"], doc.Metadata["enabled"], doc.Metadata["body"])
}
