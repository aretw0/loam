package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aretw0/introspection"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	// Create a loam service
	svc, err := loam.New("./demo-vault", loam.WithAutoInit(true), loam.WithVersioning(false))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Save some documents
	fmt.Println("Saving documents...")
	svc.SaveDocument(ctx, "doc1", "Hello World", core.Metadata{"author": "Alice"})
	svc.SaveDocument(ctx, "doc2", "Second Doc", core.Metadata{"author": "Bob"})

	// Force cache population via List
	docs, _ := svc.ListDocuments(ctx)
	fmt.Printf("Loaded %d documents into cache\n", len(docs))

	// Give reconcile a moment to complete
	time.Sleep(100 * time.Millisecond)

	// Demonstrate observability - Service implements introspection.Introspectable
	if intro, ok := interface{}(svc).(introspection.Introspectable); ok {
		fmt.Println("\n=== Service State ===")
		printState(intro)
	}

	// Access underlying repository for more detailed state
	// Note: In production, you'd use a public method to get the repository
	// For now, we'll demonstrate with a typed service setup

	repo, err := loam.Init("./demo-vault", loam.WithAutoInit(true), loam.WithVersioning(false))
	if err != nil {
		log.Fatal(err)
	}

	if intro, ok := repo.(introspection.Introspectable); ok {
		fmt.Println("\n=== Repository State ===")
		printState(intro)
	}

	var repoState fs.RepositoryState
	var repoStateOK bool
	if intro, ok := repo.(introspection.Introspectable); ok {
		if state, ok := intro.State().(fs.RepositoryState); ok {
			repoState = state
			repoStateOK = true
		}
	}

	// Start watching to see watcher state
	fmt.Println("\n=== Starting Watcher ===")
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	events, err := svc.Watch(watchCtx, "*")
	if err != nil {
		log.Fatal(err)
	}

	// Give watcher time to start (lifecycle.Go() is async)
	time.Sleep(200 * time.Millisecond)

	// Check state again
	if intro, ok := repo.(introspection.Introspectable); ok {
		fmt.Println("\n=== Repository State (with active watcher) ===")
		printState(intro)
	}

	if repoStateOK {
		fmt.Println("\n=== Vault Diagram (Mermaid) ===")
		config := introspection.DefaultDiagramConfig()
		config.SecondaryID = "vault"
		config.SecondaryLabel = "Vault Topology"
		diagram := introspection.TreeDiagram(buildVaultTree(repoState), config)
		fmt.Println(diagram)
	}

	// Show component type
	if comp, ok := repo.(introspection.Component); ok {
		fmt.Printf("\nComponent Type: %s\n", comp.ComponentType())
	}

	// Consume one event to demonstrate it's working
	select {
	case event := <-events:
		fmt.Printf("\nReceived watch event: %+v\n", event)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("\nNo events received (expected if no changes)")
	}

	fmt.Println("\n✅ Observability demonstration complete!")
}

func printState(intro introspection.Introspectable) {
	state := intro.State()
	fmt.Printf("%+v\n", state)
}

type vaultNode struct {
	Name     string
	Status   string
	Metadata map[string]string
	Children []vaultNode
}

func buildVaultTree(state fs.RepositoryState) vaultNode {
	// Status must match classes in introspection.DefaultStyles()
	// Available: created, pending, starting, running, suspended, stopping, stopped, finished, killed, failed
	// - running: Actively processing
	// - suspended: Paused/idle, waiting to be activated
	watcherStatus := "suspended"
	if state.WatcherActive {
		watcherStatus = "running"
	}

	repoNode := vaultNode{
		Name:   "Repository",
		Status: "running",
		Metadata: map[string]string{
			"type":  "process",
			"path":  state.Path,
			"cache": fmt.Sprintf("%d", state.CacheSize),
		},
		Children: []vaultNode{
			{
				Name:   "Watcher",
				Status: watcherStatus,
				Metadata: map[string]string{
					"type": "goroutine",
				},
			},
			{
				Name:   "Cache",
				Status: "running",
				Metadata: map[string]string{
					"type":    "container",
					"entries": fmt.Sprintf("%d", state.CacheSize),
				},
			},
		},
	}

	return vaultNode{
		Name:   "Vault",
		Status: "running",
		Metadata: map[string]string{
			"type": "container",
			"path": state.Path,
		},
		Children: []vaultNode{repoNode},
	}
}
