package formatter

import (
	"fmt"
	"strings"
)

// Formatter is responsible for applying final formatting to commit messages
type Formatter struct{}

// NewFormatter creates a new Formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatMessage applies formatting rules to the commit message
func (f *Formatter) FormatMessage(msg string, isMajor bool) string {
	// Capitalize the first letter
	/* if len(msg) > 0 {
		r := []rune(msg)
		r[0] = unicode.ToUpper(r[0])
		msg = string(r)
	} */

	// Remove redundant phrases
	msg = strings.ReplaceAll(msg, "add add new", "add new")
	msg = strings.ReplaceAll(msg, "feat feat", "feat")
	msg = strings.ReplaceAll(msg, "fix fix", "fix")

	// Enforce summary length (soft limit for now, try to break at word boundaries)
	if len(msg) > 72 {
		truncatedMsg := msg
		if len(truncatedMsg) > 72 {
			truncatedMsg = truncatedMsg[:72]
			lastSpace := strings.LastIndex(truncatedMsg, " ")
			if lastSpace != -1 {
				truncatedMsg = truncatedMsg[:lastSpace]
			}
			msg = fmt.Sprintf("%s...", truncatedMsg)
		}
	}

	// Add optional suffixes
	if isMajor {
		msg = fmt.Sprintf("%s (massive refactor)", msg)
	}

	return msg
}
