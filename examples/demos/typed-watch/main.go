package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/typed"
)

// AppConfig represents our typed document.
type AppConfig struct {
	Theme string `json:"theme"`
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`
}

func main() {
	// 1. Setup a temporary vault for the demo
	tmpDir, err := os.MkdirTemp("", "loam-demo-watch")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("🌱 Loam Vault: %s\n", tmpDir)

	// 2. Initialize Typed Repository
	// We use the generic wrapper directly via loam.OpenTypedRepository helper (simulated here for clarity)
	// or by wrapping a core service/repo.
	startOpts := []loam.Option{
		loam.WithAdapter("fs"),
		loam.WithAutoInit(true),
		loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))), // Quiet logger
	}

	// Create the Typed Repository
	// We use the helper loam.OpenTypedRepository to initialize the repository.
	configRepo, err := loam.OpenTypedRepository[AppConfig](tmpDir, startOpts...)
	if err != nil {
		panic(err)
	}

	// 3. Create Initial Config
	ctx := context.Background()
	initialConfig := &typed.DocumentModel[AppConfig]{
		ID:      "settings", // settings.json
		Content: "",         // JSON docs don't strictly need content body
		Data: AppConfig{
			Theme: "light",
			Port:  8080,
			Debug: false,
		},
	}
	// We must attach the saver or use repo.Save directly. Using repo.Save is cleaner.
	if err := configRepo.Save(ctx, initialConfig); err != nil {
		panic(err)
	}
	fmt.Println("✅ Initial settings saved: Theme=light")

	// 4. Start Watching
	// We want to be notified when "settings" changes.
	events, err := configRepo.Watch(ctx, "settings.*") // Watch settings.json (or .md etc)
	if err != nil {
		panic(err)
	}

	fmt.Println("👀 Watching for changes...")

	// 5. Simulate External Change (e.g. User editing file)
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("\n✏️  Simulating external edit...")

		// We use the core service to simulate a "raw" write or another process
		newConfig := AppConfig{
			Theme: "dark", // Changed!
			Port:  9090,
			Debug: true,
		}

		// Update using the same repo for simplicity, but conceptually this could be VS Code saving the file.
		updateDoc := &typed.DocumentModel[AppConfig]{
			ID:   "settings",
			Data: newConfig,
		}
		if err := configRepo.Save(ctx, updateDoc); err != nil {
			fmt.Printf("Error saving: %v\n", err)
		}
	}()

	// 6. React Loop
	timeout := time.After(3 * time.Second)

	for {
		select {
		case event := <-events:
			if event.Type == core.EventModify || event.Type == core.EventCreate {
				fmt.Printf("⚡ Event received: %s on %s\n", event.Type, event.ID)

				// Reload Configuration (Hot Reload)
				doc, err := configRepo.Get(ctx, event.ID)
				if err != nil {
					fmt.Printf("Failed to reload: %v\n", err)
					continue
				}

				fmt.Printf("🔄 Hot Reloaded! New Theme: %s, Port: %d\n", doc.Data.Theme, doc.Data.Port)
				return // Demo complete
			}
		case <-timeout:
			fmt.Println("❌ Timeout waiting for event")
			return
		}
	}
}
