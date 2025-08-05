package analyzer

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GitAnalyzer handles git repository analysis
type GitAnalyzer struct{}

// ChangeAnalysis represents the analysis of staged changes
type ChangeAnalysis struct {
	Added     []string
	Modified  []string
	Deleted   []string
	Renamed   []string
	DiffHints []string
	FileTypes map[string]int
	Scopes    []string
}

// FileChange represents a single file change
type FileChange struct {
	Status   string
	FilePath string
}

// New creates a new GitAnalyzer instance
func New() *GitAnalyzer {
	return &GitAnalyzer{}
}

// IsGitRepository checks if the current directory is a git repository
func (g *GitAnalyzer) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetStagedChanges retrieves all staged changes from git
func (g *GitAnalyzer) GetStagedChanges() ([]FileChange, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var changes []FileChange
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}

		status := parts[0]
		filePath := parts[1]

		changes = append(changes, FileChange{
			Status:   status,
			FilePath: filePath,
		})
	}

	return changes, scanner.Err()
}

// AnalyzeChanges performs comprehensive analysis of the staged changes
func (g *GitAnalyzer) AnalyzeChanges(changes []FileChange) (*ChangeAnalysis, error) {
	analysis := &ChangeAnalysis{
		Added:     []string{},
		Modified:  []string{},
		Deleted:   []string{},
		Renamed:   []string{},
		DiffHints: []string{},
		FileTypes: make(map[string]int),
		Scopes:    []string{},
	}

	scopeSet := make(map[string]bool)

	for _, change := range changes {
		// Categorize file operations
		g.categorizeChange(change, analysis)

		// Extract file information
		g.extractFileInfo(change.FilePath, analysis, scopeSet)
	}

	// Convert scope set to slice
	for scope := range scopeSet {
		analysis.Scopes = append(analysis.Scopes, scope)
	}

	// Extract diff hints
	diffHints, err := g.extractDiffHints()
	if err == nil {
		analysis.DiffHints = diffHints
	}

	return analysis, nil
}

// categorizeChange categorizes a file change by its git status
func (g *GitAnalyzer) categorizeChange(change FileChange, analysis *ChangeAnalysis) {
	status := change.Status
	filePath := change.FilePath

	if strings.Contains(status, "A") {
		analysis.Added = append(analysis.Added, filePath)
	}
	if strings.Contains(status, "M") {
		analysis.Modified = append(analysis.Modified, filePath)
	}
	if strings.Contains(status, "D") {
		analysis.Deleted = append(analysis.Deleted, filePath)
	}
	if strings.Contains(status, "R") {
		analysis.Renamed = append(analysis.Renamed, filePath)
	}
}

// extractFileInfo extracts file types and potential scopes from file paths
func (g *GitAnalyzer) extractFileInfo(filePath string, analysis *ChangeAnalysis, scopeSet map[string]bool) {
	// Extract file extension
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	if ext != "" {
		analysis.FileTypes[ext]++
	}

	// Extract potential scopes from file paths
	pathParts := strings.Split(filePath, "/")
	if len(pathParts) > 1 {
		// Use directory name as potential scope
		scopeSet[pathParts[0]] = true
	}

	// Special file patterns for scope detection
	lowerPath := strings.ToLower(filePath)

	if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "spec") {
		scopeSet["test"] = true
	}
	if strings.Contains(lowerPath, "doc") || strings.Contains(lowerPath, "readme") {
		scopeSet["docs"] = true
	}
	if strings.Contains(lowerPath, "config") || strings.Contains(lowerPath, ".config") {
		scopeSet["config"] = true
	}

	// Package management files
	packageFiles := []string{"package.json", "go.mod", "requirements.txt", "Cargo.toml", "pom.xml"}
	for _, pkgFile := range packageFiles {
		if strings.Contains(lowerPath, strings.ToLower(pkgFile)) {
			scopeSet["deps"] = true
			break
		}
	}
}

// extractDiffHints analyzes git diff output for contextual hints
func (g *GitAnalyzer) extractDiffHints() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--no-color")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	hints := make(map[string]bool)
	diffContent := string(output)

	// Define patterns for common code changes
	patterns := map[string]*regexp.Regexp{
		"added logging":    regexp.MustCompile(`\+.*(?:console\.log|fmt\.Print|log\.|logger\.|Logger)`),
		"added functions":  regexp.MustCompile(`\+.*(?:func |function |def |const .* = |let .* = )`),
		"updated imports":  regexp.MustCompile(`\+.*(?:import |require\(|#include|use )`),
		"updated exports":  regexp.MustCompile(`\+.*(?:export |module\.exports)`),
		"added todos":      regexp.MustCompile(`\+.*(?:TODO|FIXME|XXX|HACK)`),
		"async operations": regexp.MustCompile(`\+.*(?:async |await |Promise|goroutine|threading)`),
		"class changes":    regexp.MustCompile(`\+.*(?:class |struct |interface |type .*struct)`),
		"error handling":   regexp.MustCompile(`\+.*(?:try|catch|except|panic|recover|error|Error)`),
		"database changes": regexp.MustCompile(`\+.*(?:SELECT|INSERT|UPDATE|DELETE|CREATE TABLE|ALTER TABLE)`),
		"api endpoints":    regexp.MustCompile(`\+.*(?:@RequestMapping|@GetMapping|@PostMapping|router\.|app\.|http\.)`),
	}

	for hint, pattern := range patterns {
		if pattern.MatchString(diffContent) {
			hints[hint] = true
		}
	}

	// Convert to slice and limit to top 3 hints
	var result []string
	for hint := range hints {
		result = append(result, hint)
		if len(result) >= 3 {
			break
		}
	}

	return result, nil
}

// Commit creates a git commit with the provided message
func (g *GitAnalyzer) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
