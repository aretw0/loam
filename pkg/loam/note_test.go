package loam

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantKey     string
		wantVal     string
		wantErr     bool
	}{
		{
			name: "Basic Frontmatter",
			input: `---
title: Hello World
---
# Content Here`,
			wantContent: "# Content Here",
			wantKey:     "title",
			wantVal:     "Hello World",
			wantErr:     false,
		},
		{
			name:        "No Frontmatter",
			input:       `# Just Markdown`,
			wantContent: "# Just Markdown",
			wantErr:     false,
		},
		{
			name:        "Empty File",
			input:       ``,
			wantContent: "",
			wantErr:     false,
		},
		{
			name: "Invalid YAML",
			input: `---
key: : value
---
Content`,
			wantErr: true,
		},
		{
			name: "Unclosed Frontmatter",
			input: `---
title: Unclosed
Content`,
			wantErr: true,
		},
		{
			name: "Multiline Content",
			input: `---
tag: test
---
Line 1
Line 2`,
			wantContent: "Line 1\nLine 2",
			wantKey:     "tag",
			wantVal:     "test",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := Parse(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Content != tt.wantContent {
				t.Errorf("Parse() content = %q, want %q", got.Content, tt.wantContent)
			}

			if tt.wantKey != "" {
				val, ok := got.Metadata[tt.wantKey]
				if !ok {
					t.Errorf("Missing metadata key %q", tt.wantKey)
				} else if val != tt.wantVal {
					t.Errorf("Metadata[%q] = %v, want %v", tt.wantKey, val, tt.wantVal)
				}
			}
		})
	}
}
