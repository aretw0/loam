package loam

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Metadata aliases the flexible YAML frontmatter map.
type Metadata map[string]interface{}

// Note represents a Markdown file with optional YAML frontmatter.
type Note struct {
	ID       string   // Typically the filename without extension
	Metadata Metadata // Parsed YAML frontmatter
	Content  string   // The markdown body
}

// Parse reads a stream and decodes it into a Note.
// It detects if the stream starts with a frontmatter block (delimeted by ---).
func Parse(r io.Reader) (*Note, error) {
	// Read everything to memory to allow multi-pass or splitting
	// For files < 10MB this is fine. For larger, we might stream, but notes are usually small.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	n := &Note{
		Metadata: make(Metadata),
	}

	// Check for Frontmatter delimiter at the very beginning
	// Standard: Must start with ---
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		// No frontmatter, treat everything as content
		n.Content = string(data)
		return n, nil
	}

	// Find the closing delimiter
	// We look for "\n---" or "\r\n---"
	rest := data[3:] // Skip first ---

	// Careful with line endings. We expect the closing fence to be on its own line.
	// Simple approach: find the next "---" and verify it's preceded by newline.
	parts := bytes.SplitN(rest, []byte("---"), 2)
	if len(parts) == 1 {
		// Found start, but no end. Invalid frontmatter or just literal --- in text?
		// Loam decision: If it starts with ---, it expects a closing ---. If not found, error or treat as text?
		// Let's treat as text if no closing found to be safe, OR error.
		// Jekyll/Hugo usually require valid FM. Let's return error to be strict.
		return nil, errors.New("frontmatter started but no closing delimiter found")
	}

	// parts[0] is yaml, parts[1] is content
	yamlData := parts[0]
	contentData := parts[1]

	// Parse YAML
	if err := yaml.Unmarshal(yamlData, &n.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Trim leading whitespace from content (usually a newline after the second ---)
	n.Content = strings.TrimPrefix(string(contentData), "\n")
	n.Content = strings.TrimPrefix(n.Content, "\r\n")

	return n, nil
}

// String serializes the note back to Markdown with Frontmatter.
func (n *Note) String() (string, error) {
	var buf bytes.Buffer

	// Write Frontmatter if exists
	if len(n.Metadata) > 0 {
		buf.WriteString("---\n")
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2) // Pretty print
		if err := encoder.Encode(n.Metadata); err != nil {
			return "", err
		}
		encoder.Close()
		buf.WriteString("---\n")
	}

	buf.WriteString(n.Content)
	return buf.String(), nil
}
