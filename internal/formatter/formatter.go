package formatter

import (
	"fmt"
	"strings"
)

// Formatter is responsible for applying final formatting to commit messages
type Formatter struct {
	MaxSubjectLength int
	MaxBodyLength    int
}

// NewFormatter creates a new Formatter
func NewFormatter(maxSubject, maxBody int) *Formatter {
	return &Formatter{
		MaxSubjectLength: maxSubject,
		MaxBodyLength:    maxBody,
	}
}

// FormatMessage applies formatting rules to the commit message
func (f *Formatter) FormatMessage(msg string, isMajor bool) string {
	if msg == "" {
		return ""
	}

	// Split into subject and body
	parts := strings.SplitN(msg, "\n", 2)
	subject := strings.TrimSpace(parts[0])
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}

	// Remove redundant phrases from subject
	subject = strings.ReplaceAll(subject, "add add new", "add new")
	subject = strings.ReplaceAll(subject, "feat feat", "feat")
	subject = strings.ReplaceAll(subject, "fix fix", "fix")

	// Add optional suffixes to subject
	if isMajor {
		subject = fmt.Sprintf("%s (massive refactor)", subject)
	}

	// Wrap subject if too long
	if f.MaxSubjectLength > 0 && len(subject) > f.MaxSubjectLength {
		wrapped := f.wrapString(subject, f.MaxSubjectLength)
		subjectParts := strings.SplitN(wrapped, "\n", 2)
		subject = subjectParts[0]
		if len(subjectParts) > 1 {
			if body != "" {
				body = subjectParts[1] + "\n\n" + body
			} else {
				body = subjectParts[1]
			}
		}
	}

	// Wrap body if exists
	if body != "" && f.MaxBodyLength > 0 {
		body = f.wrapString(body, f.MaxBodyLength)
	}

	if body != "" {
		return subject + "\n\n" + body
	}
	return subject
}

// wrapString wraps a string at the specified limit, preserving paragraphs
func (f *Formatter) wrapString(s string, limit int) string {
	if limit <= 0 {
		return s
	}

	paragraphs := strings.Split(s, "\n\n")
	var result strings.Builder

	for i, p := range paragraphs {
		if i > 0 {
			result.WriteString("\n\n")
		}

		words := strings.Fields(p)
		if len(words) == 0 {
			continue
		}

		currentLineLength := 0
		for j, word := range words {
			if j > 0 {
				if currentLineLength+1+len(word) > limit {
					result.WriteString("\n")
					currentLineLength = 0
				} else {
					result.WriteString(" ")
					currentLineLength++
				}
			}

			result.WriteString(word)
			currentLineLength += len(word)
		}
	}

	return result.String()
}
