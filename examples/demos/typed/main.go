package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
)

// User represents our domain object.
type User struct {
	Name  string   `json:"name"`
	Role  string   `json:"role"`
	Level int      `json:"level"`
	Tags  []string `json:"tags"`
}

func main() {
	// 1. Setup Repository
	wd, _ := os.Getwd()

	// Use Loam's Dev Safety: Force temp directory for this demo to ensure no local files are touched.
	repoPath := loam.ResolveVaultPath(filepath.Join(wd, "data"), true)
	fmt.Printf("Repository Path: %s\n", repoPath)

	// 1. Initialize Typed Repository using new OpenTypedRepository factory
	userRepo, err := loam.OpenTypedRepository[User](repoPath,
		loam.WithVersioning(false), // Gitless for demo
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 3. Create & Save a User (Model)
	newUser := &loam.DocumentModel[User]{
		ID:      "users/jdoe",
		Content: "This user was created via Typed Repository",
		Data: User{
			Name:  "John Doe",
			Role:  "Admin",
			Level: 42,
			Tags:  []string{"staff", "verified"},
		},
	}

	fmt.Printf("Saving user %s...\n", newUser.Data.Name)
	if err := userRepo.Save(ctx, newUser); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}

	// 4. Retrieve & Modify
	fmt.Println("Retrieving user...")
	loadedUser, err := userRepo.Get(ctx, "users/jdoe")
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}

	fmt.Printf("Loaded: %+v\n", loadedUser.Data)
	fmt.Printf("Content: %s\n", loadedUser.Content)

	// Modify (Active Record Style)
	loadedUser.Data.Level++
	if err := loadedUser.Save(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("User updated successfully!")

	// 5. Cleanup
	fmt.Printf("\nCheck the file at: %s/users/jdoe.md\n", repoPath)
}
