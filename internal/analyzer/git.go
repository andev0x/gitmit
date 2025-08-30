package analyzer

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GitAnalyzer handles git repository analysis
type GitAnalyzer struct{}

// ChangeAnalysis represents the analysis of staged changes
type ChangeAnalysis struct {
	Added            []string
	Modified         []string
	Deleted          []string
	Renamed          []string
	DiffHints        []string
	FileTypes        map[string]int
	Scopes           []string
	LanguagePatterns map[string]int
	CodeComplexity   string
	ChangeImpact     string
	SemanticHints    []string
	FunctionChanges  []string
	VariableChanges  []string
	ImportChanges    []string
	CommentChanges   []string
	ErrorPatterns    []string
	PerformanceHints []string
	SecurityHints    []string
	TestChanges      []string
	ConfigChanges    []string
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

// GetStagedDiff retrieves the diff of all staged changes
func (g *GitAnalyzer) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
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

	// Enhanced patterns for better language understanding
	patterns := map[string]*regexp.Regexp{
		// Function and method changes
		"added functions":   regexp.MustCompile(`\+.*(?:func |function |def |const .* = |let .* = |var .* = )`),
		"updated functions": regexp.MustCompile(`^[+-].*(?:func |function |def )`),
		"removed functions": regexp.MustCompile(`^-.*(?:func |function |def )`),

		// Import and dependency changes
		"updated imports":    regexp.MustCompile(`^[+-].*(?:import |require\(|#include|use |from )`),
		"added dependencies": regexp.MustCompile(`\+.*(?:package\.json|go\.mod|requirements\.txt|Cargo\.toml|pom\.xml)`),
		"updated exports":    regexp.MustCompile(`^[+-].*(?:export |module\.exports|pub |public )`),

		// Error handling and logging
		"added logging":      regexp.MustCompile(`\+.*(?:console\.log|fmt\.Print|log\.|logger\.|Logger|println|printf)`),
		"error handling":     regexp.MustCompile(`^[+-].*(?:try|catch|except|panic|recover|error|Error|throw|raise)`),
		"added error checks": regexp.MustCompile(`\+.*(?:if.*error|if.*err|check.*error|assert.*error)`),

		// Async and concurrency
		"async operations": regexp.MustCompile(`^[+-].*(?:async |await |Promise|goroutine|threading|future|coroutine)`),
		"concurrency":      regexp.MustCompile(`^[+-].*(?:mutex|lock|semaphore|channel|select|spawn|thread)`),

		// Class and structure changes
		"class changes":       regexp.MustCompile(`^[+-].*(?:class |struct |interface |type .*struct|trait |enum )`),
		"method changes":      regexp.MustCompile(`^[+-].*(?:public |private |protected |static |final )`),
		"constructor changes": regexp.MustCompile(`^[+-].*(?:constructor|init|new |__init__)`),

		// Database and data operations
		"database changes": regexp.MustCompile(`^[+-].*(?:SELECT|INSERT|UPDATE|DELETE|CREATE TABLE|ALTER TABLE|DROP TABLE|CREATE INDEX)`),
		"query changes":    regexp.MustCompile(`^[+-].*(?:query|Query|sql|SQL|mongo|Mongo|redis|Redis)`),
		"data validation":  regexp.MustCompile(`^[+-].*(?:validate|validation|check|verify|assert)`),

		// API and web changes
		"api endpoints":     regexp.MustCompile(`^[+-].*(?:@RequestMapping|@GetMapping|@PostMapping|@PutMapping|@DeleteMapping|router\.|app\.|http\.|endpoint|route)`),
		"http methods":      regexp.MustCompile(`^[+-].*(?:GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)`),
		"response handling": regexp.MustCompile(`^[+-].*(?:response|Response|json|JSON|xml|XML|status|Status)`),

		// Security and authentication
		"security features": regexp.MustCompile(`^[+-].*(?:password|hash|encrypt|decrypt|jwt|token|auth|security|signature|verify)`),
		"authentication":    regexp.MustCompile(`^[+-].*(?:login|logout|register|signin|signout|oauth|saml|ldap)`),
		"authorization":     regexp.MustCompile(`^[+-].*(?:permission|role|access|authorize|grant|deny)`),

		// Configuration and settings
		"configuration":    regexp.MustCompile(`^[+-].*(?:config|settings|env|\.env|\.config|\.toml|\.yaml|\.yml|\.json|\.ini)`),
		"environment vars": regexp.MustCompile(`^[+-].*(?:process\.env|os\.environ|getenv|setenv|environment)`),
		"feature flags":    regexp.MustCompile(`^[+-].*(?:feature.*flag|toggle|switch|enable|disable)`),

		// Deployment and infrastructure
		"deployment":     regexp.MustCompile(`^[+-].*(?:docker|kubernetes|k8s|deploy|helm|dockerfile|docker-compose|kustomize)`),
		"infrastructure": regexp.MustCompile(`^[+-].*(?:terraform|ansible|puppet|chef|cloudformation|serverless)`),
		"monitoring":     regexp.MustCompile(`^[+-].*(?:metrics|monitoring|alert|prometheus|grafana|jaeger|zipkin)`),

		// Performance and optimization
		"performance":       regexp.MustCompile(`^[+-].*(?:cache|optimize|performance|speed|fast|efficient|benchmark|profile)`),
		"memory management": regexp.MustCompile(`^[+-].*(?:memory|gc|garbage|alloc|free|malloc|new|delete)`),
		"algorithm changes": regexp.MustCompile(`^[+-].*(?:algorithm|sort|search|filter|map|reduce|fold)`),

		// Testing and quality
		"testing":        regexp.MustCompile(`^[+-].*(?:test|spec|mock|stub|fixture|assert|expect|describe|it|should)`),
		"test coverage":  regexp.MustCompile(`^[+-].*(?:coverage|coverage\.go|\.test\.|_test\.|test_|spec_|\.spec\.)`),
		"quality checks": regexp.MustCompile(`^[+-].*(?:lint|linter|eslint|gofmt|black|rustfmt|clang-format|prettier)`),

		// Documentation and comments
		"documentation":     regexp.MustCompile(`^[+-].*(?:readme|docs|documentation|comment|javadoc|godoc|rustdoc|docstring)`),
		"code comments":     regexp.MustCompile(`^[+-].*(?://|/\*|\*/|#|<!--|-->|"""|'''|///|//!)`),
		"api documentation": regexp.MustCompile(`^[+-].*(?:swagger|openapi|api.*doc|postman|insomnia)`),

		// Style and formatting
		"style formatting": regexp.MustCompile(`^[+-].*(?:prettier|eslint|gofmt|black|rustfmt|clang-format|indent|spacing)`),
		"code style":       regexp.MustCompile(`^[+-].*(?:style|format|lint|beautify|uglify|minify)`),

		// Dependency and package management
		"dependency updates": regexp.MustCompile(`^[+-].*(?:package\.json|go\.mod|requirements\.txt|Cargo\.toml|pom\.xml|composer\.json|Gemfile)`),
		"version changes":    regexp.MustCompile(`^[+-].*(?:version|Version|v\d+\.\d+\.\d+|semver|major|minor|patch)`),

		// Revert and rollback
		"revert changes":   regexp.MustCompile(`^[+-].*(?:revert|rollback|undo|restore|backup|rollback|restore)`),
		"work in progress": regexp.MustCompile(`^[+-].*(?:wip|work in progress|draft|temporary|temp|TODO|FIXME|XXX|HACK)`),

		// UI and frontend
		"ui components":    regexp.MustCompile(`^[+-].*(?:component|Component|widget|Widget|view|View|page|Page)`),
		"styling":          regexp.MustCompile(`^[+-].*(?:css|CSS|scss|less|stylus|styled|className|class=|style=)`),
		"user interaction": regexp.MustCompile(`^[+-].*(?:onClick|onChange|onSubmit|event|Event|handler|Handler)`),

		// Mobile and platform specific
		"mobile features":   regexp.MustCompile(`^[+-].*(?:android|Android|ios|iOS|react-native|flutter|mobile|Mobile)`),
		"platform specific": regexp.MustCompile(`^[+-].*(?:windows|Windows|macos|macOS|linux|Linux|darwin|Darwin)`),

		// Internationalization
		"i18n":         regexp.MustCompile(`^[+-].*(?:i18n|internationalization|localization|locale|translation|translate)`),
		"localization": regexp.MustCompile(`^[+-].*(?:gettext|po|mo|locale|language|lang|trans)`),
	}

	for hint, pattern := range patterns {
		if pattern.MatchString(diffContent) {
			hints[hint] = true
		}
	}

	// Convert to slice and limit to top 5 hints
	var result []string
	for hint := range hints {
		result = append(result, hint)
		if len(result) >= 5 {
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

// GetLastCommitMessage retrieves the message of the last commit
func (g *GitAnalyzer) GetLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// AmendCommit amends the last commit with the provided message
func (g *GitAnalyzer) AmendCommit(message string) error {
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	return cmd.Run()
}

// GetRecentCommits retrieves the last n commit messages
func (g *GitAnalyzer) GetRecentCommits(n int) (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%s", "-n", fmt.Sprintf("%d", n))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
