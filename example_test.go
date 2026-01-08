package loam_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

// Example_basic demonstrates how to initialize a Vault, save a note, and read it back.
func Example_basic() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", "loam-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize the Loam service (Vault) targeting the temporary directory.
	// WithAutoInit(true) ensures the underlying storage (git repo) is initialized.
	vault, err := loam.New(tmpDir, loam.WithAutoInit(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 1. Save a Document
	err = vault.SaveDocument(ctx, "hello-world", "This is my first note in Loam.", core.Metadata{
		"tags":   []string{"example"},
		"author": "Gopher",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Read it back
	doc, err := vault.GetDocument(ctx, "hello-world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found document: %s\n", doc.ID)
	// Output:
	// Found document: hello-world
}

// ExampleNewTypedRepository demonstrates how to use the Generic Typed Wrapper for type safety.
func ExampleNewTypedRepository() {
	// Setup: Temporary repository
	tmpDir, err := os.MkdirTemp("", "loam-typed-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Use loam.Init to get the Repository directly
	repo, err := loam.Init(filepath.Join(tmpDir, "vault"), loam.WithAutoInit(true))
	if err != nil {
		log.Fatal(err)
	}

	// Define your Domain Model
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Wrap the repository
	userRepo := loam.NewTypedRepository[User](repo)
	ctx := context.Background()

	// Save a typed document
	err = userRepo.Save(ctx, &loam.DocumentModel[User]{
		ID:      "users/alice",
		Content: "Alice's Profile",
		Data: User{
			Name:  "Alice",
			Email: "alice@example.com",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve it back
	doc, err := userRepo.Get(ctx, "users/alice")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User Name: %s\n", doc.Data.Name)
	// Output:
	// User Name: Alice
}

// Example_csvNestedData demonstrates Loam's "Smart CSV" capability,
// which automatically handles nested structures (like maps or slices)
// by serializing them as JSON within the CSV column.
func Example_csvNestedData() {
	// Setup: Temporary repository
	tmpDir, err := os.MkdirTemp("", "loam-csv-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := loam.Init(filepath.Join(tmpDir, "vault"), loam.WithAutoInit(true))
	if err != nil {
		log.Fatal(err)
	}

	type Metrics struct {
		Host string            `json:"host"`
		Tags map[string]string `json:"tags"` // Nested Map
		Load []int             `json:"load"` // Nested Slice
	}

	metricsRepo := loam.NewTypedRepository[Metrics](repo)
	ctx := context.Background()

	// 1. Save complex data to CSV
	err = metricsRepo.Save(ctx, &loam.DocumentModel[Metrics]{
		ID: "metrics/server-01.csv", // .csv extension triggers CSV adapter
		Data: Metrics{
			Host: "server-01",
			Tags: map[string]string{"env": "prod", "region": "us-east"},
			Load: []int{10, 20, 15},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Read it back
	// Loam automatically parses the JSON strings inside the CSV back into Maps and Slices.
	doc, err := metricsRepo.Get(ctx, "metrics/server-01.csv")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", doc.Data.Host)
	fmt.Printf("Tag Region: %s\n", doc.Data.Tags["region"])
	fmt.Printf("Load: %v\n", doc.Data.Load)
	// Output:
	// Host: server-01
	// Tag Region: us-east
	// Load: [10 20 15]
}

// Example_strictMode demonstrates how to enable global strict mode for type fidelity.
// This ensures that large integers (int64) are not lost as float64 during parsing
// across ALL supported formats (JSON, YAML, Markdown).
func Example_strictMode() {
	// Setup
	tmpDir, err := os.MkdirTemp("", "loam-strict-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize with Global Strict Mode
	// This applies strict parsing (json.Number) to all serializers.
	repo, err := loam.Init(filepath.Join(tmpDir, "vault"),
		loam.WithAutoInit(true),
		loam.WithStrict(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 1. JSON Example (Large Int)
	jsonContent := `{"big_id": 9223372036854775807, "type": "json"}`
	_ = os.WriteFile(filepath.Join(tmpDir, "vault", "strict.json"), []byte(jsonContent), 0644)

	// 2. YAML Example (Large Int)
	yamlContent := "big_id: 9223372036854775807\ntype: yaml"
	_ = os.WriteFile(filepath.Join(tmpDir, "vault", "strict.yaml"), []byte(yamlContent), 0644)

	// Read back and verify types
	for _, file := range []string{"strict.json", "strict.yaml"} {
		doc, err := repo.Get(ctx, file)
		if err != nil {
			log.Fatal(err)
		}

		val := doc.Metadata["big_id"]
		fmt.Printf("[%s] Type: %T\n", doc.Metadata["type"], val)
	}

	// Output:
	// [json] Type: json.Number
	// [yaml] Type: json.Number
}
