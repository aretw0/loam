package fs_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// TestPolyglotConsistency verifies that when strict mode is enabled,
// both JSON and YAML serializers return consistent numeric types (json.Number).
func TestPolyglotConsistency(t *testing.T) {
	// 1. Setup Data
	jsonBody := `{"count": 123, "price": 10.5}`
	yamlBody := "count: 123\nprice: 10.5"

	// 2. Initialize Serializers with Strict Mode
	// We simulate what the repository does internally
	jsonSerializer := fs.NewJSONSerializer(true)
	// Now strict should also apply to YAML
	yamlSerializer := &fs.YAMLSerializer{Strict: true}

	// 3. Parse
	docJSON, err := jsonSerializer.Parse(bytes.NewReader([]byte(jsonBody)), "")
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	docYAML, err := yamlSerializer.Parse(bytes.NewReader([]byte(yamlBody)), "")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// 4. Compare Types
	checkField(t, "count", docJSON, docYAML)
	checkField(t, "price", docJSON, docYAML)
}

func checkField(t *testing.T, field string, docA, docB *core.Document) {
	t.Helper()
	valA := docA.Metadata[field]
	valB := docB.Metadata[field]

	typeA := reflect.TypeOf(valA)
	typeB := reflect.TypeOf(valB)

	// Currently, this is expected to FAIL.
	// JSON Strict -> json.Number
	// YAML -> int or float64
	if typeA != typeB {
		t.Errorf("Type mismatch for field '%s':\n\tJSON Strict: %T (%v)\n\tYAML Default: %T (%v)\n\tExpected them to be identical (json.Number)",
			field, valA, valA, valB, valB)
	}

	// Optional: Check if values are logically equal, just to be sure
	// But the main goal is type consistency.
}
