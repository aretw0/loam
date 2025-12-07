package loam

import (
	"strings"
)

// CommitType constants for semantic commits
const (
	CommitTypeFeat     = "feat"
	CommitTypeFix      = "fix"
	CommitTypeDocs     = "docs"
	CommitTypeStyle    = "style"
	CommitTypeRefactor = "refactor"
	CommitTypePerf     = "perf"
	CommitTypeTest     = "test"
	CommitTypeChore    = "chore"
)

// FormatCommitMessage builds a Conventional Commit message.
// logic:
//
//	<type>(<scope>): <subject>
//
//	<body>
//
//	Powered-by: Loam
func FormatCommitMessage(ctype, scope, subject, body string) string {
	var sb strings.Builder

	// Header
	if ctype == "" {
		ctype = CommitTypeChore // Default fallback if empty, though CLI might enforce validation
	}
	sb.WriteString(ctype)

	if scope != "" {
		sb.WriteString("(")
		sb.WriteString(scope)
		sb.WriteString(")")
	}

	sb.WriteString(": ")
	sb.WriteString(subject)

	// Body
	if body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(strings.TrimSpace(body))
	}

	// Footer
	sb.WriteString("\n\n")
	sb.WriteString("Powered-by: Loam")

	return sb.String()
}

// AppendFooter appends the Loam footer to an arbitrary message if not present.
// Used for free-form -m "msg" commits.
func AppendFooter(msg string) string {
	if strings.Contains(msg, "Powered-by: Loam") {
		return msg
	}

	// Ensure we don't glue it to the last line
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	// Ensure we have a blank line separation if it looks like a one-liner
	// If it's multi-line, we still want a blank line before footer standardly
	if !strings.HasSuffix(msg, "\n\n") {
		msg += "\n"
	}

	return msg + "Powered-by: Loam"
}
