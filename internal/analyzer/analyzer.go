package analyzer

import (
	"path/filepath"
	"strings"

	"gitmit/internal/parser"
)

// CommitMessage represents the analyzed commit message components
type CommitMessage struct {
	Action  string
	Topic   string
	Item    string
	Purpose string
	Scope   string
	IsMajor bool
}

// Analyzer is responsible for analyzing git changes and generating commit message components
type Analyzer struct {
	changes []*parser.Change
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(changes []*parser.Change) *Analyzer {
	return &Analyzer{changes: changes}
}

// AnalyzeChanges analyzes the git changes and returns a CommitMessage
func (a *Analyzer) AnalyzeChanges() *CommitMessage {
	if len(a.changes) == 0 {
		return nil
	}

	// For simplicity, we'll base the analysis on the first change.
	// A more sophisticated implementation would group changes.
	firstChange := a.changes[0]

	action := a.determineAction(firstChange)
	topic := a.determineTopic(firstChange.File)
	item := a.determineItem(firstChange.File)
	purpose := a.determinePurpose(firstChange.Diff)

	// Handle multiple modules by creating a scope
	scope := ""
	if len(a.changes) > 1 {
		topics := make(map[string]struct{})
		for _, change := range a.changes {
			topics[a.determineTopic(change.File)] = struct{}{}
		}
		if len(topics) > 1 {
			var topicList []string
			for t := range topics {
				topicList = append(topicList, t)
			}
			scope = strings.Join(topicList, ", ")
			topic = "core" // or "multiple-modules"
		}
	}

	isMajor := false
	for _, change := range a.changes {
		if change.IsMajor {
			isMajor = true
			break
		}
	}

	return &CommitMessage{
		Action:  action,
		Topic:   topic,
		Item:    item,
		Purpose: purpose,
		Scope:   scope,
		IsMajor: isMajor,
	}
}

func (a *Analyzer) determineAction(change *parser.Change) string {
	switch change.Action {
	case "A":
		return "feat"
	case "M":
		// Simple logic: if "fix" or "bug" is in the diff, it's a fix. Otherwise, refactor.
		if strings.Contains(change.Diff, "fix") || strings.Contains(change.Diff, "bug") {
			return "fix"
		}
		return "refactor"
	case "D":
		return "chore"
	case "R":
		return "refactor"
	default:
		return "chore"
	}
}

func (a *Analyzer) determineTopic(path string) string {
	parts := strings.Split(filepath.Dir(path), string(filepath.Separator))
	if len(parts) > 0 && parts[0] != "." {
		// Example: "internal/analyzer" -> "analyzer"
		if len(parts) > 1 && parts[0] == "internal" {
			return parts[1]
		}
		return parts[0]
	}
	return "core"
}

func (a *Analyzer) determineItem(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (a *Analyzer) determinePurpose(diff string) string {
	keywords := map[string]string{
		"login":      "authentication",
		"validate":   "validation",
		"query":      "database query",
		"cache":      "caching",
		"middleware": "middleware",
		"test":       "testing",
		"config":     "configuration",
		"ci":         "ci/cd",
		"docs":       "documentation",
		"log":        "logging",
		"sql":        "database logic",
		"gorm":       "database logic",
	}

	for keyword, purpose := range keywords {
		if strings.Contains(diff, keyword) {
			return purpose
		}
	}
	return "general update"
}
