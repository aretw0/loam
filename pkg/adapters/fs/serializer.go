package fs

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aretw0/loam/pkg/core"
	"gopkg.in/yaml.v3"
)

// Serializer defines how to read and write a specific file format.
type Serializer interface {
	// Parse reads from r and returns a Document.
	Parse(r io.Reader, metadataKey string) (*core.Document, error)
	// Serialize converts the Document to bytes.
	Serialize(doc core.Document, metadataKey string) ([]byte, error)
}

// DefaultSerializers returns the standard set of serializers.
func DefaultSerializers(strict bool) map[string]Serializer {
	return map[string]Serializer{
		".json": NewJSONSerializer(strict),
		".yaml": NewYAMLSerializer(strict),
		".yml":  NewYAMLSerializer(strict),
		".csv":  NewCSVSerializer(strict),
		".md":   NewMarkdownSerializer(strict),
	}
}

// --- JSON Serializer ---

// JSONSerializer handles reading and writing JSON files.
type JSONSerializer struct {
	// Strict enables strict number parsing (as json.Number) to avoid precision loss.
	Strict bool
}

// NewJSONSerializer creates a new JSON serializer.
// Optional strict mode prevents float64 conversion for large integers.
func NewJSONSerializer(strict bool) *JSONSerializer {
	return &JSONSerializer{Strict: strict}
}

func (s *JSONSerializer) Parse(r io.Reader, metadataKey string) (*core.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if s.Strict {
		decoder.UseNumber()
	}
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	doc := &core.Document{Metadata: make(core.Metadata)}
	if metadataKey != "" {
		if meta, ok := payload[metadataKey].(map[string]interface{}); ok {
			doc.Metadata = meta
			delete(payload, metadataKey)
		}
	} else {
		doc.Metadata = payload
	}

	if c, ok := payload["content"].(string); ok {
		doc.Content = c
		if metadataKey == "" {
			delete(doc.Metadata, "content")
		}
	}

	return doc, nil
}

func (s *JSONSerializer) Serialize(doc core.Document, metadataKey string) ([]byte, error) {
	payload := make(map[string]interface{})

	if metadataKey != "" {
		payload[metadataKey] = doc.Metadata
	} else {
		for k, v := range doc.Metadata {
			payload[k] = v
		}
	}

	if doc.Content != "" || metadataKey == "" {
		payload["content"] = doc.Content
	}

	return json.MarshalIndent(payload, "", "  ")
}

// --- YAML Serializer ---

type YAMLSerializer struct {
	// Strict enables strict number parsing (as json.Number) to avoid precision loss.
	Strict bool
}

// NewYAMLSerializer creates a new YAML serializer.
// Optional strict mode prevents float64 conversion for large integers.
func NewYAMLSerializer(strict bool) *YAMLSerializer {
	return &YAMLSerializer{Strict: strict}
}

func (s *YAMLSerializer) Parse(r io.Reader, metadataKey string) (*core.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	doc := &core.Document{Metadata: make(core.Metadata)}
	if metadataKey != "" {
		if meta, ok := payload[metadataKey].(map[string]interface{}); ok {
			doc.Metadata = meta
			delete(payload, metadataKey)
		}
	} else {
		doc.Metadata = payload
	}

	if c, ok := payload["content"].(string); ok {
		doc.Content = c
		if metadataKey == "" {
			delete(doc.Metadata, "content")
		}
	}

	if s.Strict {
		doc.Metadata = recursiveNormalize(doc.Metadata).(core.Metadata)
	}

	return doc, nil
}

func (s *YAMLSerializer) Serialize(doc core.Document, metadataKey string) ([]byte, error) {
	payload := make(map[string]interface{})

	if metadataKey != "" {
		payload[metadataKey] = doc.Metadata
	} else {
		for k, v := range doc.Metadata {
			payload[k] = v
		}
	}

	if doc.Content != "" || metadataKey == "" {
		payload["content"] = doc.Content
	}

	return yaml.Marshal(payload)
}

// --- Markdown Serializer ---

type MarkdownSerializer struct {
	// Strict enables strict number parsing (as json.Number) to avoid precision loss.
	Strict bool
}

// NewMarkdownSerializer creates a new Markdown serializer.
// Optional strict mode prevents float64 conversion for large integers.
func NewMarkdownSerializer(strict bool) *MarkdownSerializer {
	return &MarkdownSerializer{Strict: strict}
}

func (s *MarkdownSerializer) Parse(r io.Reader, metadataKey string) (*core.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	doc := &core.Document{Metadata: make(core.Metadata)}

	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		doc.Content = string(data)
		return doc, nil
	}

	rest := data[3:]
	parts := bytes.SplitN(rest, []byte("---"), 2)
	if len(parts) == 1 {
		return nil, errors.New("frontmatter started but no closing delimiter found")
	}

	yamlData := parts[0]
	contentData := parts[1]

	if err := yaml.Unmarshal(yamlData, &doc.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	doc.Content = strings.TrimPrefix(string(contentData), "\n")
	doc.Content = strings.TrimPrefix(doc.Content, "\r\n")

	if s.Strict {
		doc.Metadata = recursiveNormalize(doc.Metadata).(core.Metadata)
	}

	return doc, nil
}

func (s *MarkdownSerializer) Serialize(doc core.Document, metadataKey string) ([]byte, error) {
	var buf bytes.Buffer
	if len(doc.Metadata) > 0 {
		buf.WriteString("---\n")
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(doc.Metadata); err != nil {
			return nil, err
		}
		encoder.Close()
		buf.WriteString("---\n")
	}
	buf.WriteString(doc.Content)
	return buf.Bytes(), nil
}

// --- CSV Serializer ---

type CSVSerializer struct {
	// Strict enables strict number parsing (as json.Number) to avoid precision loss.
	Strict bool
}

// NewCSVSerializer creates a new CSV serializer.
// Optional strict mode prevents float64 conversion for large integers.
func NewCSVSerializer(strict bool) *CSVSerializer {
	return &CSVSerializer{Strict: strict}
}

func (s *CSVSerializer) Parse(r io.Reader, metadataKey string) (*core.Document, error) {
	// CSV parsing for a SINGLE document usually implies reading the *first row*?
	// Or is this only used for "raw file" handling?
	//
	// In Loam, ParseDocument is usually called on a file.
	// If the file is a CSV, valid usage is ambiguous.
	//
	// However, existing logic (ParseDocument inside repository.go) assumed:
	// "Parse a generic file".
	// The existing implementation read header + first row.
	// We will preserve that behavior.

	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv header: %w", err)
	}
	row, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv row: %w", err)
	}

	if len(row) != len(headers) {
		return nil, fmt.Errorf("csv row length mismatch")
	}

	doc := &core.Document{Metadata: make(core.Metadata)}
	for i, h := range headers {
		val := row[i]
		if strings.EqualFold(h, "content") {
			doc.Content = val
		} else {
			valTrimmed := strings.TrimSpace(val)
			doc.Metadata[h] = UnmarshalCSVValue(valTrimmed, s.Strict)
		}
	}

	if s.Strict {
		doc.Metadata = recursiveNormalize(doc.Metadata).(core.Metadata)
	}

	return doc, nil
}

func (s *CSVSerializer) Serialize(doc core.Document, metadataKey string) ([]byte, error) {
	keys := []string{"content"}
	// Deterministic order for testing? Maps are random.
	// For serialization of a single doc, it doesn't matter much unless we append.
	for k := range doc.Metadata {
		keys = append(keys, k)
	}

	var row []string
	row = append(row, doc.Content)

	for _, k := range keys[1:] {
		v := doc.Metadata[k]
		if v == nil {
			row = append(row, "")
			continue
		}
		row = append(row, MarshalCSVValue(v))
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(keys); err != nil {
		return nil, err
	}
	if err := w.Write(row); err != nil {
		return nil, err
	}
	w.Flush()
	return buf.Bytes(), nil
}

// --- Helpers ---

// UnmarshalCSVValue attempts to parse a string as JSON if it looks like a Map or Slice.
// Otherwise returns the string as is.
//
// CAVEAT: This uses a heuristic (starts/ends with {} or []). It is possible for a raw string
// that happens to be valid JSON (e.g. "{foo}") to be interpreted as an object.
// Use with caution if your domain allows strings that mimic JSON structure.
func UnmarshalCSVValue(val string, strict bool) interface{} {
	valTrimmed := strings.TrimSpace(val)
	if (strings.HasPrefix(valTrimmed, "{") && strings.HasSuffix(valTrimmed, "}")) ||
		(strings.HasPrefix(valTrimmed, "[") && strings.HasSuffix(valTrimmed, "]")) {
		var parsed interface{}
		decoder := json.NewDecoder(strings.NewReader(val))
		if strict {
			decoder.UseNumber()
		}
		if err := decoder.Decode(&parsed); err == nil {
			return parsed
		}
	}
	return val
}

// MarshalCSVValue converts a value to a string, using JSON for complex types (Map, Slice).
func MarshalCSVValue(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}, []interface{}, map[string]string, []string:
		// Complex types -> JSON String
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}
	}
	// Fallback
	return fmt.Sprintf("%v", v)
}

// recursiveNormalize traverses the map/slice and converts numeric types to json.Number.
// This ensures consistency with JSON Strict mode.
func recursiveNormalize(val interface{}) interface{} {
	switch v := val.(type) {
	case core.Metadata:
		// Convert to map[string]interface{} for processing, then cast back if needed?
		// Actually core.Metadata IS map[string]interface{}
		m := make(core.Metadata)
		for k, val := range v {
			m[k] = recursiveNormalize(val)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[k] = recursiveNormalize(val)
		}
		return m
	case []interface{}:
		l := make([]interface{}, len(v))
		for i, val := range v {
			l[i] = recursiveNormalize(val)
		}
		return l
	case int:
		return json.Number(fmt.Sprintf("%d", v))
	case int64:
		return json.Number(fmt.Sprintf("%d", v))
	case float64:
		// Format without scientific notation if possible for integers, but with decimals for floats.
		// -1 precision uses the smallest number of digits necessary.
		return json.Number(fmt.Sprintf("%v", v))
	case int32:
		return json.Number(fmt.Sprintf("%d", v))
	default:
		return v
	}
}
