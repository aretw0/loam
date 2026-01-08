package fs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aretw0/loam/pkg/core"
)

func TestSerializers(t *testing.T) {
	doc := core.Document{
		ID:      "test/doc",
		Content: "Hello World",
		Metadata: core.Metadata{
			"title": "Test Title",
			"tags":  []interface{}{"a", "b"},
			"meta": map[string]interface{}{
				"foo": "bar",
			},
			"count": 42.0, // JSON unmarshal uses float64
		},
	}

	serializers := DefaultSerializers(false)

	tests := []struct {
		ext string
	}{
		{".json"},
		{".yaml"},
		{".md"},
		{".csv"},
	}

	for _, tc := range tests {
		t.Run(tc.ext, func(t *testing.T) {
			s := serializers[tc.ext]

			// Serialize
			data, err := s.Serialize(doc, "")
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}

			// Parse back
			parsed, err := s.Parse(bytes.NewReader(data), "")
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Check Content
			// CSV/JSON might trim whitespace differently?
			if strings.TrimSpace(parsed.Content) != strings.TrimSpace(doc.Content) {
				t.Errorf("Content mismatch. Want %q, got %q", doc.Content, parsed.Content)
			}

			// Check Metadata
			if parsed.Metadata["title"] != "Test Title" {
				t.Errorf("Metadata 'title' mismatch")
			}

			// Check Nested Data (The main goal)
			// Markdown frontmatter (yaml) handles nested checks.
			// CSV now handles them via Smart JSON.

			// Tags
			tags, ok := parsed.Metadata["tags"].([]interface{})
			if !ok {
				t.Logf("Tag type: %T", parsed.Metadata["tags"])
				t.Errorf("Metadata 'tags' is not []interface{}")
			} else {
				if len(tags) != 2 {
					t.Errorf("Tags length mismatch")
				}
			}

			// Meta Map
			// YAML unmarshalling into a defined type (core.Metadata) might preserve the type for nested maps
			// depending on the library behavior, or just return map[string]interface{}.
			// We check for both.
			val := parsed.Metadata["meta"]
			var meta map[string]interface{}

			switch v := val.(type) {
			case map[string]interface{}:
				meta = v
			case core.Metadata:
				meta = map[string]interface{}(v)
			default:
				t.Logf("Meta type: %T", val)
				t.Errorf("Metadata 'meta' is not map[string]interface{} or core.Metadata")
			}

			if meta != nil {
				if meta["foo"] != "bar" {
					t.Errorf("Meta 'foo' mismatch")
				}
			}
		})
	}

	t.Run("CSV False Positives", func(t *testing.T) {
		s := &CSVSerializer{}

		// Case 1: Invalid JSON (should remain string)
		csvData := `content,val
foo,"[invalid_json"`

		doc, err := s.Parse(strings.NewReader(csvData), "")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if doc.Metadata["val"] != "[invalid_json" {
			t.Errorf("Expected raw string for invalid/partial JSON")
		}

		// Case 2: Valid JSON-like string which we WANT to be a string?
		val := "{Alice}"
		// IMPORTANT: Check error to avoid panic if Parse fails
		doc2, err := s.Parse(strings.NewReader("content,val\nfoo,"+val), "")
		if err != nil {
			t.Fatalf("Parse failed for case 2: %v", err)
		}
		if _, ok := doc2.Metadata["val"].(string); !ok {
			t.Errorf("Expected string for invalid JSON '{Alice}', got %T", doc2.Metadata["val"])
		}

		// Case 3: Valid JSON Object (Should be parsed as Map)
		// CSV: foo,"{""key"":""value""}" -> Cell: {"key":"value"}
		val3 := `"{""key"":""value""}"`
		doc3, err := s.Parse(strings.NewReader("content,val\nfoo,"+val3), "")
		if err != nil {
			t.Fatalf("Parse failed for case 3: %v", err)
		}

		if m, ok := doc3.Metadata["val"].(map[string]interface{}); !ok {
			t.Errorf("Expected map for valid JSON object, got %T", doc3.Metadata["val"])
		} else {
			if m["key"] != "value" {
				t.Errorf("Map content mismatch")
			}
		}
	})
}

func TestJSONSerializer_Strict(t *testing.T) {
	jsonContent := `{"big_id": 9223372036854775807}` // Max Int64
	reader := strings.NewReader(jsonContent)

	// 1. Strict Mode
	strictSerializer := NewJSONSerializer(true)
	doc, err := strictSerializer.Parse(reader, "")
	if err != nil {
		t.Fatalf("Strict Parse failed: %v", err)
	}

	val := doc.Metadata["big_id"]
	// Should be json.Number
	if _, ok := val.(json.Number); !ok {
		t.Errorf("Strict Mode: Expected json.Number, got %T", val)
	}

	// 2. Loose Mode (Default)
	reader.Reset(jsonContent)
	looseSerializer := NewJSONSerializer(false)
	docLoose, err := looseSerializer.Parse(reader, "")
	if err != nil {
		t.Fatalf("Loose Parse failed: %v", err)
	}

	valLoose := docLoose.Metadata["big_id"]
	// Should be float64
	if _, ok := valLoose.(float64); !ok {
		t.Errorf("Loose Mode: Expected float64, got %T", valLoose)
	} else if valLoose.(float64) != 9223372036854775807 {
		t.Errorf("Loose Mode: Precision loss in JSON! Want %d, got %f", 9223372036854775807, valLoose.(float64))
	}
}

func TestYAMLSerializer_Strict(t *testing.T) {
	yamlContent := "big_id: 9223372036854775807\nprice: 10.50"
	reader := strings.NewReader(yamlContent)

	// 1. Strict Mode
	s := NewYAMLSerializer(true)
	doc, err := s.Parse(reader, "")
	if err != nil {
		t.Fatalf("Strict Parse failed: %v", err)
	}

	// Verify Int64 -> json.Number
	val := doc.Metadata["big_id"]
	if n, ok := val.(json.Number); !ok {
		t.Errorf("Strict Mode (Int): Expected json.Number, got %T", val)
	} else if n.String() != "9223372036854775807" {
		t.Errorf("Strict Mode (Int): Precision loss! Want %s, got %s", "9223372036854775807", n.String())
	}

	// Verify Float -> json.Number
	valFloat := doc.Metadata["price"]
	if _, ok := valFloat.(json.Number); !ok {
		t.Errorf("Strict Mode (Float): Expected json.Number, got %T", valFloat)
	} else if valFloat.(json.Number).String() != "10.5" {
		t.Errorf("Strict Mode (Float): Precision loss! Want %s, got %s", "10.5", valFloat.(json.Number).String())
	}

	// 2. Loose Mode (Default)
	reader.Reset(yamlContent)
	sLoose := NewYAMLSerializer(false)
	docLoose, err := sLoose.Parse(reader, "")
	if err != nil {
		t.Fatalf("Loose Parse failed: %v", err)
	}

	// Verify Int -> int/int64 (YAML parser specific, likely int or int64)
	valLoose := docLoose.Metadata["big_id"]
	switch valLoose.(type) {
	case int:
		// OK
	case int64:
		// OK
	case float64:
		// OK but check value
		if valLoose != 9223372036854775807 {
			t.Errorf("Loose Mode (Int): Precision loss! Want %d, got %d", 9223372036854775807, valLoose)
		}
	default:
		t.Errorf("Loose Mode: Expected native number type, got %T", valLoose)
	}
}

func TestMarkdownSerializer_Strict(t *testing.T) {
	mdContent := `---
count: 9223372036854775807
score: 99.9
---
# Hello
`
	reader := strings.NewReader(mdContent)

	// 1. Strict Mode
	s := NewMarkdownSerializer(true)
	doc, err := s.Parse(reader, "")
	if err != nil {
		t.Fatalf("Strict Parse failed: %v", err)
	}

	if n, ok := doc.Metadata["count"].(json.Number); !ok {
		t.Errorf("Strict Mode: Expected json.Number for 'count', got %T", doc.Metadata["count"])
	} else if n.String() != "9223372036854775807" {
		t.Errorf("Strict Mode: Precision loss in Markdown! Want %s, got %s", "9223372036854775807", n.String())
	}

	if _, ok := doc.Metadata["score"].(json.Number); !ok {
		t.Errorf("Strict Mode: Expected json.Number for 'score', got %T", doc.Metadata["score"])
	} else if doc.Metadata["score"].(json.Number).String() != "99.9" {
		t.Errorf("Strict Mode: Precision loss in Markdown! Want %s, got %s", "99.9", doc.Metadata["score"].(json.Number).String())
	}
}

func TestCSVSerializer_Strict(t *testing.T) {
	// Value is MaxInt64 (9223372036854775807), which loses precision as float64
	csvContent := `content,val
foo,"{""big_id"": 9223372036854775807}"`

	reader := strings.NewReader(csvContent)

	// 1. Strict Mode
	s := NewCSVSerializer(true)
	doc, err := s.Parse(reader, "")
	if err != nil {
		t.Fatalf("Strict Parse failed: %v", err)
	}

	// Verify the nested JSON was parsed with number fidelity
	valMap, ok := doc.Metadata["val"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", doc.Metadata["val"])
	}

	bigID := valMap["big_id"]
	if n, ok := bigID.(json.Number); !ok {
		t.Errorf("Strict Mode: Expected json.Number for nested 'big_id', got %T (Value: %v)", bigID, bigID)
	} else if n.String() != "9223372036854775807" {
		t.Errorf("Strict Mode: Precision loss in CSV! Want %s, got %s", "9223372036854775807", n.String())
	}

	// 2. Loose Mode (Default)
	reader.Reset(csvContent)
	sLoose := NewCSVSerializer(false)
	docLoose, err := sLoose.Parse(reader, "")
	if err != nil {
		t.Fatalf("Loose Parse failed: %v", err)
	}

	valMapLoose := docLoose.Metadata["val"].(map[string]interface{})
	bigIDLoose := valMapLoose["big_id"]

	// Should be float64 in loose mode
	if _, ok := bigIDLoose.(float64); !ok {
		t.Errorf("Loose Mode: Expected float64, got %T", bigIDLoose)
	} else if bigIDLoose.(float64) != 9223372036854775807 {
		t.Errorf("Loose Mode: Precision loss in CSV! Want %d, got %f", 9223372036854775807, bigIDLoose.(float64))
	}
}
