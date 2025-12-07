package loam

import "testing"

func TestFormatCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		ctype   string
		scope   string
		subject string
		body    string
		want    string
	}{
		{
			name:    "simple",
			ctype:   "feat",
			scope:   "",
			subject: "add feature",
			body:    "",
			want:    "feat: add feature\n\nPowered-by: Loam",
		},
		{
			name:    "with scope",
			ctype:   "fix",
			scope:   "api",
			subject: "fix bug",
			body:    "",
			want:    "fix(api): fix bug\n\nPowered-by: Loam",
		},
		{
			name:    "with body",
			ctype:   "docs",
			scope:   "",
			subject: "update readme",
			body:    "Added new examples.",
			want:    "docs: update readme\n\nAdded new examples.\n\nPowered-by: Loam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCommitMessage(tt.ctype, tt.scope, tt.subject, tt.body)
			if got != tt.want {
				t.Errorf("FormatCommitMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendFooter(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{
			name: "plain",
			msg:  "simple message",
			want: "simple message\n\nPowered-by: Loam",
		},
		{
			name: "already has newline",
			msg:  "line 1\n",
			want: "line 1\n\nPowered-by: Loam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendFooter(tt.msg)
			if got != tt.want {
				t.Errorf("AppendFooter() = %q, want %q", got, tt.want)
			}
		})
	}
}
