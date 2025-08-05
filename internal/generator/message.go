package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andev0x/gitmit/internal/analyzer"
)

// MessageGenerator generates conventional commit messages
type MessageGenerator struct{}

// CommitType represents different types of commits
type CommitType string

const (
	Feat     CommitType = "feat"
	Fix      CommitType = "fix"
	Refactor CommitType = "refactor"
	Chore    CommitType = "chore"
	Test     CommitType = "test"
	Docs     CommitType = "docs"
	Style    CommitType = "style"
	Perf     CommitType = "perf"
	CI       CommitType = "ci"
	Build    CommitType = "build"
)

// New creates a new MessageGenerator instance
func New() *MessageGenerator {
	return &MessageGenerator{}
}

// GenerateMessage generates a conventional commit message based on analysis
func (m *MessageGenerator) GenerateMessage(analysis *analyzer.ChangeAnalysis) string {
	commitType := m.determineCommitType(analysis)
	scope := m.determineScope(analysis)
	description := m.generateDescription(analysis)

	// Format according to Conventional Commits specification
	message := string(commitType)
	if scope != "" {
		message += fmt.Sprintf("(%s)", scope)
	}
	message += fmt.Sprintf(": %s", description)

	return message
}

// determineCommitType analyzes changes to determine the appropriate commit type
func (m *MessageGenerator) determineCommitType(analysis *analyzer.ChangeAnalysis) CommitType {
	if commitType, ok := m.determineCommitTypeFromScopes(analysis); ok {
		return commitType
	}
	if commitType, ok := m.determineCommitTypeFromFileTypes(analysis); ok {
		return commitType
	}
	if commitType, ok := m.determineCommitTypeFromDiffHints(analysis); ok {
		return commitType
	}
	if commitType, ok := m.determineCommitTypeFromFileOps(analysis); ok {
		return commitType
	}

	// Default fallback
	return Chore
}

func (m *MessageGenerator) determineCommitTypeFromScopes(analysis *analyzer.ChangeAnalysis) (CommitType, bool) {
	for _, scope := range analysis.Scopes {
		switch scope {
		case "test":
			return Test, true
		case "docs":
			return Docs, true
		case "deps":
			return Chore, true
		case "ci", ".github", ".gitlab":
			return CI, true
		case "build", "webpack", "rollup", "vite":
			return Build, true
		}
	}
	return "", false
}

func (m *MessageGenerator) determineCommitTypeFromFileTypes(analysis *analyzer.ChangeAnalysis) (CommitType, bool) {
	if analysis.FileTypes["md"] > 0 || analysis.FileTypes["txt"] > 0 || analysis.FileTypes["rst"] > 0 {
		return Docs, true
	}
	return "", false
}

func (m *MessageGenerator) determineCommitTypeFromDiffHints(analysis *analyzer.ChangeAnalysis) (CommitType, bool) {
	for _, hint := range analysis.DiffHints {
		lowerHint := strings.ToLower(hint)
		if strings.Contains(lowerHint, "fix") || strings.Contains(lowerHint, "bug") ||
			strings.Contains(lowerHint, "error") || strings.Contains(lowerHint, "patch") {
			return Fix, true
		}
		if strings.Contains(lowerHint, "performance") || strings.Contains(lowerHint, "optimize") {
			return Perf, true
		}
		if strings.Contains(lowerHint, "style") || strings.Contains(lowerHint, "format") {
			return Style, true
		}
	}
	return "", false
}

func (m *MessageGenerator) determineCommitTypeFromFileOps(analysis *analyzer.ChangeAnalysis) (CommitType, bool) {
	hasAdded := len(analysis.Added) > 0
	hasModified := len(analysis.Modified) > 0
	hasDeleted := len(analysis.Deleted) > 0
	hasRenamed := len(analysis.Renamed) > 0

	// New features (primarily new files with minimal modifications)
	if hasAdded && !hasModified && !hasDeleted {
		return Feat, true
	}

	// Refactoring (renames or structural changes)
	if hasRenamed || (hasModified && hasDeleted && !hasAdded) {
		return Refactor, true
	}

	// Modifications to existing files (could be features or fixes)
	if hasModified {
		// Default to feat for positive changes unless hints suggest otherwise
		return Feat, true
	}

	// Pure deletions
	if hasDeleted && !hasAdded && !hasModified {
		return Chore, true
	}

	return "", false
}

// determineScope determines the most appropriate scope for the commit
func (m *MessageGenerator) determineScope(analysis *analyzer.ChangeAnalysis) string {
	if len(analysis.Scopes) == 0 {
		return ""
	}

	// Priority order for scopes
	priorityScopes := []string{
		"api", "ui", "auth", "db", "database", "security",
		"test", "docs", "config", "deps", "ci", "build",
	}

	// Check for priority scopes first
	for _, priority := range priorityScopes {
		for _, scope := range analysis.Scopes {
			if strings.EqualFold(scope, priority) {
				return scope
			}
		}
	}

	// Return the first scope if no priority match
	return analysis.Scopes[0]
}

// generateDescription creates a descriptive commit message
func (m *MessageGenerator) generateDescription(analysis *analyzer.ChangeAnalysis) string {
	// Use diff hints if available and meaningful
	if len(analysis.DiffHints) > 0 {
		return analysis.DiffHints[0]
	}

	// Generate description based on file operations
	var operations []string

	if len(analysis.Added) > 0 {
		if len(analysis.Added) == 1 {
			fileName := m.getFileName(analysis.Added[0])
			operations = append(operations, fmt.Sprintf("add %s", fileName))
		} else {
			operations = append(operations, fmt.Sprintf("add %d files", len(analysis.Added)))
		}
	}

	if len(analysis.Modified) > 0 {
		if len(analysis.Modified) == 1 {
			fileName := m.getFileName(analysis.Modified[0])
			operations = append(operations, fmt.Sprintf("update %s", fileName))
		} else {
			operations = append(operations, fmt.Sprintf("update %d files", len(analysis.Modified)))
		}
	}

	if len(analysis.Deleted) > 0 {
		if len(analysis.Deleted) == 1 {
			fileName := m.getFileName(analysis.Deleted[0])
			operations = append(operations, fmt.Sprintf("remove %s", fileName))
		} else {
			operations = append(operations, fmt.Sprintf("remove %d files", len(analysis.Deleted)))
		}
	}

	if len(analysis.Renamed) > 0 {
		if len(analysis.Renamed) == 1 {
			fileName := m.getFileName(analysis.Renamed[0])
			operations = append(operations, fmt.Sprintf("rename %s", fileName))
		} else {
			operations = append(operations, fmt.Sprintf("rename %d files", len(analysis.Renamed)))
		}
	}

	if len(operations) == 0 {
		return "update files"
	}

	// Join operations with "and"
	if len(operations) == 1 {
		return operations[0]
	} else if len(operations) == 2 {
		return fmt.Sprintf("%s and %s", operations[0], operations[1])
	} else {
		return fmt.Sprintf("%s and %d more changes", operations[0], len(operations)-1)
	}
}

// getFileName extracts the filename from a file path
func (m *MessageGenerator) getFileName(filePath string) string {
	return filepath.Base(filePath)
}
