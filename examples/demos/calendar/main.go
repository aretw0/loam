package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	// 1. Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	vaultPath := "./calendar-vault"

	service, err := loam.New(vaultPath,
		loam.WithAutoInit(true),
		loam.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("📅 Calendar Vault ready at: %s\n", vaultPath)

	// 2. Define an Event
	eventID := fmt.Sprintf("meeting-%d", time.Now().Unix())
	eventContent := `
# Weekly Team Sync

- Discuss project roadmap
- Review blockers
- **Action Items:**
    - [ ] Update documentation
`
	eventMeta := core.Metadata{
		"title":     "Weekly Sync",
		"start":     time.Now().Format(time.RFC3339),
		"end":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"attendees": []string{"alice@example.com", "bob@example.com"},
		"status":    "confirmed",
	}

	// 3. Save with Semantic Reason
	// "feat" indicates a new entry, "fix" would be a reschedule.
	fmt.Println("Creating event...")
	reason := loam.FormatChangeReason(loam.CommitTypeFeat, "cal", "schedule weekly sync", "Recurring event")
	ctx := context.WithValue(context.Background(), core.ChangeReasonKey, reason)

	if err := service.SaveNote(ctx, eventID, eventContent, eventMeta); err != nil {
		panic(err)
	}

	fmt.Printf("✅ Event '%s' scheduled!\n", eventID)
	fmt.Println("\nRun 'git log' in the vault to see the semantic history.")
}
