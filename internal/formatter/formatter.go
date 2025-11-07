package formatter

import (
	"strings"
	"unicode"
)

// Formatter is responsible for applying final formatting to commit messages
type Formatter struct{}

// NewFormatter creates a new Formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatMessage applies formatting rules to the commit message
func (f *Formatter) FormatMessage(msg string) string {
	// Capitalize the first letter
	if len(msg) > 0 {
		r := []rune(msg)
		r[0] = unicode.ToUpper(r[0])
		msg = string(r)
	}

	// Remove redundant phrases
	msg = strings.ReplaceAll(msg, "add add new", "add new")
	msg = strings.ReplaceAll(msg, "feat feat", "feat")
	msg = strings.ReplaceAll(msg, "fix fix", "fix")

	// Enforce summary length (soft limit for now, just truncate)
	if len(msg) > 72 {
		msg = msg[:72]
	}

	return msg
}
