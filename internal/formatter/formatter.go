package formatter

import (
	"fmt"
	"regexp"
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
		body = strings.TrimLeft(parts[1], "\n\r")
		body = strings.TrimRight(body, "\n\r\t ")
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
			// Subject overflow becomes the start of the body
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

// wrapString wraps a string at the specified limit, preserving paragraphs and structures
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

		lines := strings.Split(p, "\n")
		var currentParagraph strings.Builder

		for _, line := range lines {
			if f.isStructural(line) {
				// Flush any pending paragraph text
				if currentParagraph.Len() > 0 {
					result.WriteString(f.reflow(currentParagraph.String(), limit))
					result.WriteString("\n")
					currentParagraph.Reset()
				}
				// Wrap the structural line itself (preserving its prefix if possible)
				result.WriteString(f.wrapLine(line, limit))
				result.WriteString("\n")
			} else {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				if currentParagraph.Len() > 0 {
					currentParagraph.WriteString(" ")
				}
				currentParagraph.WriteString(trimmed)
			}
		}

		if currentParagraph.Len() > 0 {
			result.WriteString(f.reflow(currentParagraph.String(), limit))
		}

		// Cleanup trailing newline from structural lines at the end of a paragraph
		resStr := result.String()
		if strings.HasSuffix(resStr, "\n") {
			result.Reset()
			result.WriteString(strings.TrimSuffix(resStr, "\n"))
		}
	}

	return result.String()
}

// isStructural identifies lines that should not be reflowed into paragraphs
func (f *Formatter) isStructural(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// List markers at the start of the trimmed line
	markers := []string{"- ", "* ", "+ "}
	for _, m := range markers {
		if strings.HasPrefix(trimmed, m) {
			return true
		}
	}

	// Numeric list markers (e.g., "1. ")
	numericRegex := regexp.MustCompile(`^\d+\.\s`)
	if numericRegex.MatchString(trimmed) {
		return true
	}

	// Significant indentation (at least 2 spaces or a tab)
	if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
		return true
	}

	return false
}

// reflow joins words and wraps them at the limit
func (f *Formatter) reflow(s string, limit int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var res strings.Builder
	currLen := 0
	for i, w := range words {
		if i > 0 {
			if currLen+1+len(w) > limit {
				res.WriteString("\n")
				currLen = 0
			} else {
				res.WriteString(" ")
				currLen++
			}
		}
		res.WriteString(w)
		currLen += len(w)
	}
	return res.String()
}

// wrapLine wraps a single structural line, attempting to preserve indentation
func (f *Formatter) wrapLine(line string, limit int) string {
	if len(line) <= limit {
		return line
	}

	// Find indentation and prefix
	indent := ""
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indent += string(char)
		} else {
			break
		}
	}

	// Also check for list markers
	content := line[len(indent):]
	prefix := ""
	markers := []string{"- ", "* ", "+ "}
	for _, m := range markers {
		if strings.HasPrefix(content, m) {
			prefix = m
			content = content[len(m):]
			break
		}
	}

	// Handle numeric markers
	numericRegex := regexp.MustCompile(`^\d+\.\s`)
	if loc := numericRegex.FindStringIndex(content); loc != nil && loc[0] == 0 {
		prefix = content[loc[0]:loc[1]]
		content = content[loc[1]:]
	}

	words := strings.Fields(content)
	if len(words) == 0 {
		return line
	}

	var res strings.Builder
	res.WriteString(indent)
	res.WriteString(prefix)
	currLen := len(indent) + len(prefix)

	for i, w := range words {
		if i > 0 {
			if currLen+1+len(w) > limit {
				res.WriteString("\n")
				res.WriteString(indent)
				// Extra indentation for wrapped list items
				if prefix != "" {
					res.WriteString("  ")
				}
				currLen = len(indent)
				if prefix != "" {
					currLen += 2
				}
			} else {
				res.WriteString(" ")
				currLen++
			}
		}
		res.WriteString(w)
		currLen += len(w)
	}

	return res.String()
}
