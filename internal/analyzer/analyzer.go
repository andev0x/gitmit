package analyzer

import (
	"bufio"
	"path/filepath"
	"strings"

	"gitmit/internal/config"
	"gitmit/internal/parser"
)

// CommitMessage represents the analyzed commit message components
type CommitMessage struct {
	Action         string
	Topic          string
	Item           string
	Purpose        string
	Scope          string
	IsMajor        bool
	TotalAdded     int
	TotalRemoved   int
	FileExtensions []string
	RenamedFiles   []*parser.Change
	CopiedFiles    []*parser.Change
	IsDocsOnly     bool
	IsConfigOnly   bool
	IsDepsOnly     bool
}

// Analyzer is responsible for analyzing git changes and generating commit message components
type Analyzer struct {
	changes []*parser.Change
	config  *config.Config
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(changes []*parser.Change, cfg *config.Config) *Analyzer {
	return &Analyzer{changes: changes, config: cfg}
}

// AnalyzeChanges analyzes the git changes and returns a CommitMessage
func (a *Analyzer) AnalyzeChanges(totalAdded, totalRemoved int) *CommitMessage {
	if len(a.changes) == 0 {
		return nil
	}

	commitMessage := &CommitMessage{
		TotalAdded:   totalAdded,
		TotalRemoved: totalRemoved,
	}

	var allFileExtensions []string
	var allTopics []string
	var allPurposes []string
	var allItems []string

	for _, change := range a.changes {
		if change.IsRename {
			commitMessage.RenamedFiles = append(commitMessage.RenamedFiles, change)
		}
		if change.IsCopy {
			commitMessage.CopiedFiles = append(commitMessage.CopiedFiles, change)
		}
		if change.IsMajor {
			commitMessage.IsMajor = true
		}

		allFileExtensions = append(allFileExtensions, change.FileExtension)
		allTopics = append(allTopics, a.determineTopic(change.File))
		allPurposes = append(allPurposes, a.determinePurpose(change.Diff))
		allItems = append(allItems, a.determineItem(change.File))
	}

	commitMessage.FileExtensions = uniqueStrings(allFileExtensions)

	// Determine if changes are only documentation, config, or dependencies
	commitMessage.IsDocsOnly = a.isDocsOnly()
	commitMessage.IsConfigOnly = a.isConfigOnly()
	commitMessage.IsDepsOnly = a.isDepsOnly()

	// Apply smart fallback logic
	if msg := a.applySmartFallback(commitMessage); msg != nil {
		return msg
	}

	// Default analysis based on the first change if no specific fallback applies
	firstChange := a.changes[0]
	commitMessage.Action = a.determineAction(firstChange)
	commitMessage.Topic = a.determineTopic(firstChange.File)
	commitMessage.Item = a.determineItem(firstChange.File)
	commitMessage.Purpose = a.determinePurpose(firstChange.Diff)

	// Handle multiple modules by creating a scope
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
			commitMessage.Scope = strings.Join(topicList, ", ")
			commitMessage.Topic = "core" // or "multiple-modules"
		}
	}

	return commitMessage
}

func (a *Analyzer) determineAction(change *parser.Change) string {
	switch change.Action {
	case "A":
		// Enhanced rule: detect added tests
		if strings.HasSuffix(change.File, "_test.go") {
			return "test"
		}
		return "feat"
	case "M":
		// Enhanced rule: detect increased logging
		if a.detectIncreasedLogging(change.Diff) {
			return "feat"
		}
		// Simple logic: if "fix" or "bug" is in the diff, it's a fix. Otherwise, refactor.
		if strings.Contains(change.Diff, "fix") || strings.Contains(change.Diff, "bug") {
			return "fix"
		}
		return "refactor"
	case "D":
		// Enhanced rule: detect removed functions
		if a.detectRemovedFunctions(change.Diff) {
			return "refactor"
		}
		return "chore"
	case "R":
		return "refactor"
	case "C":
		return "feat"
	default:
		return "chore"
	}
}

func (a *Analyzer) determineTopic(path string) string {
	// Apply custom topic mappings from config
	for pattern, topic := range a.config.TopicMappings {
		if strings.Contains(path, pattern) {
			return topic
		}
	}

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
	// Apply custom keyword mappings from config
	for keyword, purpose := range a.config.KeywordMappings {
		if strings.Contains(strings.ToLower(diff), strings.ToLower(keyword)) {
			return purpose
		}
	}

	keywords := map[string]string{
		"login":      "authentication",
		"validate":   "validation",
		"query":      "database query",
		"cache":      "caching",
		"refactor":   "code restructuring",
		"logging":    "logging",
		"docs":       "documentation",
		"middleware": "middleware",
		"test":       "testing",
		"config":     "configuration",
		"ci":         "ci/cd",
		"log":        "logging", "sql": "database logic",
		"gorm":     "database logic",
		"feat":     "new feature",
		"bug":      "bug fix",
		"fix":      "bug fix",
		"cleanup":  "code cleanup",
		"perf":     "performance improvement",
		"security": "security update",
		"dep":      "dependency update",
		"build":    "build system",
		"style":    "code style",
	}

	for keyword, purpose := range keywords {
		if strings.Contains(strings.ToLower(diff), keyword) {
			return purpose
		}
	}
	return "general update"
}

func (a *Analyzer) applySmartFallback(msg *CommitMessage) *CommitMessage {
	// If more than 5 files are both added and deleted -> suggest “refactor(core): restructure project”.
	if len(a.changes) > 5 && msg.TotalAdded > 0 && msg.TotalRemoved > 0 && (float64(msg.TotalAdded+msg.TotalRemoved)/float64(len(a.changes))) > 10 { // Heuristic for significant changes across many files
		return &CommitMessage{Action: "refactor", Topic: "core", Purpose: "restructure project"}
	}

	// If .env, .yml, or Dockerfile is changed -> use ci(config): update build configuration.
	for _, ext := range msg.FileExtensions {
		if ext == "env" || ext == "yml" || ext == "yaml" || ext == "Dockerfile" {
			return &CommitMessage{Action: "ci", Topic: "config", Purpose: "update build configuration"}
		}
	}

	// If only markdown or documentation files changed -> use docs: update documentation.
	if msg.IsDocsOnly {
		return &CommitMessage{Action: "docs", Topic: "", Purpose: "update documentation"}
	}

	// If dependencies changed in go.mod -> use chore(deps): update dependencies.
	for _, change := range a.changes {
		if change.File == "go.mod" {
			return &CommitMessage{Action: "chore", Topic: "deps", Purpose: "update dependencies"}
		}
	}

	return nil
}

func (a *Analyzer) isDocsOnly() bool {
	if len(a.changes) == 0 {
		return false
	}
	for _, change := range a.changes {
		if !strings.HasPrefix(change.File, "docs/") && !strings.HasPrefix(change.File, "wiki/") && change.FileExtension != "md" && change.FileExtension != "txt" {
			return false
		}
	}
	return true
}

func (a *Analyzer) isConfigOnly() bool {
	if len(a.changes) == 0 {
		return false
	}
	for _, change := range a.changes {
		if !strings.Contains(change.File, "config") && change.FileExtension != "json" && change.FileExtension != "yaml" && change.FileExtension != "yml" && change.FileExtension != "env" && change.File != "Dockerfile" {
			return false
		}
	}
	return true
}

func (a *Analyzer) isDepsOnly() bool {
	if len(a.changes) == 0 {
		return false
	}
	for _, change := range a.changes {
		if change.File != "go.mod" && change.File != "go.sum" && change.FileExtension != "mod" && change.FileExtension != "sum" {
			return false
		}
	}
	return true
}

func (a *Analyzer) detectIncreasedLogging(diff string) bool {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") && (strings.Contains(line, "log.") || strings.Contains(line, "fmt.Print")) {
			return true
		}
	}
	return false
}

func (a *Analyzer) detectRemovedFunctions(diff string) bool {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "-") && strings.Contains(line, "func ") {
			return true
		}
	}
	return false
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}
