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
<<<<<<< HEAD
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
=======
	Action            string
	Topic             string
	Item              string
	Purpose           string
	Scope             string
	IsMajor           bool
	TotalAdded        int
	TotalRemoved      int
	FileExtensions    []string
	RenamedFiles      []*parser.Change
	CopiedFiles       []*parser.Change
	IsDocsOnly        bool
	IsConfigOnly      bool
	IsDepsOnly        bool
	DetectedFunctions []string
	DetectedStructs   []string
	DetectedMethods   []string
	ChangePatterns    []string
>>>>>>> 1028df8 (fix(config): updated imports)
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
	var allFunctions []string
	var allStructs []string
	var allMethods []string
	var allPatterns []string

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

		// Detect code structures
		funcs := a.detectFunctions(change.Diff)
		allFunctions = append(allFunctions, funcs...)

		structs := a.detectStructs(change.Diff)
		allStructs = append(allStructs, structs...)

		methods := a.detectMethods(change.Diff)
		allMethods = append(allMethods, methods...)

		// Detect change patterns
		patterns := a.detectChangePatterns(change)
		allPatterns = append(allPatterns, patterns...)
	}

	commitMessage.FileExtensions = uniqueStrings(allFileExtensions)
	commitMessage.DetectedFunctions = uniqueStrings(allFunctions)
	commitMessage.DetectedStructs = uniqueStrings(allStructs)
	commitMessage.DetectedMethods = uniqueStrings(allMethods)
	commitMessage.ChangePatterns = uniqueStrings(allPatterns)

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

	// Enhanced scope detection for multiple modules
	if len(a.changes) > 1 {
		scope := a.detectIntelligentScope()
		if scope != "" {
			commitMessage.Scope = scope
		}
	}

	// Detect multi-file patterns
	multiPatterns := a.detectMultiFilePatterns()
	if len(multiPatterns) > 0 {
		// Adjust action and purpose based on multi-file patterns
		if contains(multiPatterns, "feature-addition") {
			commitMessage.Action = "feat"
			commitMessage.Purpose = "add new feature across multiple modules"
		} else if contains(multiPatterns, "bug-fix-cascade") {
			commitMessage.Action = "fix"
			commitMessage.Purpose = "resolve issue across multiple components"
		} else if contains(multiPatterns, "refactor-sweep") {
			commitMessage.Action = "refactor"
			commitMessage.Purpose = "restructure and improve code organization"
		} else if contains(multiPatterns, "test-suite-update") {
			commitMessage.Action = "test"
			commitMessage.Purpose = "update test suite"
		}
	}

	return commitMessage
}

// detectIntelligentScope determines the best scope based on file paths and patterns
func (a *Analyzer) detectIntelligentScope() string {
	if len(a.changes) == 0 {
		return ""
	}

	topics := make(map[string]int)
	directories := make(map[string]int)

	for _, change := range a.changes {
		topic := a.determineTopic(change.File)
		topics[topic]++

		dir := filepath.Dir(change.File)
		if dir != "." {
			parts := strings.Split(dir, string(filepath.Separator))
			if len(parts) > 0 {
				// Count first-level directory
				directories[parts[0]]++
			}
		}
	}

	// If all changes are in the same topic, use that topic
	if len(topics) == 1 {
		for topic := range topics {
			return topic
		}
	}

	// If changes span multiple topics but are in the same directory tree
	if len(directories) == 1 {
		for dir := range directories {
			return dir
		}
	}

	// If changes span multiple but related topics, create combined scope
	if len(topics) <= 3 {
		var topicList []string
		for topic := range topics {
			topicList = append(topicList, topic)
		}
		// Sort for consistency
		for i := 0; i < len(topicList); i++ {
			for j := i + 1; j < len(topicList); j++ {
				if topicList[i] > topicList[j] {
					topicList[i], topicList[j] = topicList[j], topicList[i]
				}
			}
		}
		return strings.Join(topicList, ",")
	}

	// For many topics, use "core" or most common topic
	maxCount := 0
	mostCommonTopic := "core"
	for topic, count := range topics {
		if count > maxCount {
			maxCount = count
			mostCommonTopic = topic
		}
	}

	return mostCommonTopic
}

// detectMultiFilePatterns identifies patterns across multiple files
func (a *Analyzer) detectMultiFilePatterns() []string {
	if len(a.changes) <= 1 {
		return nil
	}

	var patterns []string

	// Count file types and actions
	addedFiles := 0
	modifiedFiles := 0
	deletedFiles := 0
	testFiles := 0
	configFiles := 0

	for _, change := range a.changes {
		switch change.Action {
		case "A":
			addedFiles++
		case "M":
			modifiedFiles++
		case "D":
			deletedFiles++
		}

		if strings.HasSuffix(change.File, "_test.go") {
			testFiles++
		}

		if strings.Contains(change.File, "config") ||
			change.FileExtension == "json" ||
			change.FileExtension == "yaml" ||
			change.FileExtension == "yml" {
			configFiles++
		}
	}

	// Feature addition pattern: multiple new files
	if addedFiles >= 3 && float64(addedFiles)/float64(len(a.changes)) > 0.6 {
		patterns = append(patterns, "feature-addition")
	}

	// Bug fix cascade: modifications across multiple files
	if modifiedFiles >= 3 && float64(modifiedFiles)/float64(len(a.changes)) > 0.6 {
		// Check if files are related
		hasFixKeyword := false
		for _, change := range a.changes {
			if strings.Contains(change.Diff, "fix") || strings.Contains(change.Diff, "bug") {
				hasFixKeyword = true
				break
			}
		}
		if hasFixKeyword {
			patterns = append(patterns, "bug-fix-cascade")
		}
	}

	// Refactor sweep: mix of additions, modifications, and deletions
	if addedFiles > 0 && modifiedFiles > 0 && deletedFiles > 0 && len(a.changes) >= 4 {
		patterns = append(patterns, "refactor-sweep")
	}

	// Test suite update: majority of changes are tests
	if testFiles > 0 && float64(testFiles)/float64(len(a.changes)) > 0.7 {
		patterns = append(patterns, "test-suite-update")
	}

	// Configuration update: majority are config files
	if configFiles > 0 && float64(configFiles)/float64(len(a.changes)) > 0.7 {
		patterns = append(patterns, "config-update")
	}

	// API redesign: multiple handler/api files modified
	apiFiles := 0
	for _, change := range a.changes {
		if strings.Contains(change.File, "handler") ||
			strings.Contains(change.File, "api") ||
			strings.Contains(change.File, "route") {
			apiFiles++
		}
	}
	if apiFiles >= 3 {
		patterns = append(patterns, "api-redesign")
	}

	// Database migration: multiple db/migration files
	dbFiles := 0
	for _, change := range a.changes {
		if strings.Contains(change.File, "migration") ||
			strings.Contains(change.File, "database") ||
			strings.Contains(change.File, "schema") {
			dbFiles++
		}
	}
	if dbFiles >= 2 {
		patterns = append(patterns, "database-migration")
	}

	return patterns
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (a *Analyzer) determineAction(change *parser.Change) string {
	switch change.Action {
	case "A":
		// Enhanced rule: detect added tests
		if strings.HasSuffix(change.File, "_test.go") {
			return "test"
		}
		// Detect new API endpoints
		if strings.Contains(change.Diff, "func ") && (strings.Contains(change.File, "handler") ||
			strings.Contains(change.File, "api") || strings.Contains(change.File, "route")) {
			return "feat"
		}
		return "feat"
	case "M":
		// Use detected patterns for better action determination
		diff := change.Diff

		// Check for security updates
		if strings.Contains(diff, "security") || strings.Contains(diff, "vulnerability") {
			return "security"
		}

		// Check for performance improvements
		if strings.Contains(diff, "optimize") || strings.Contains(diff, "performance") ||
			strings.Contains(diff, "cache") || strings.Contains(diff, "goroutine") {
			return "perf"
		}

		// Enhanced rule: detect increased logging
		if a.detectIncreasedLogging(diff) {
			return "feat"
		}

		// Check for bug fixes
		if strings.Contains(diff, "fix") || strings.Contains(diff, "bug") ||
			strings.Contains(diff, "issue") || strings.Contains(diff, "resolve") {
			return "fix"
		}

		// Check for style changes
		if a.isStyleChange(diff) {
			return "style"
		}

		// Check for test updates
		if strings.HasSuffix(change.File, "_test.go") {
			return "test"
		}

		// Default to refactor for modifications
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

// isStyleChange detects if changes are primarily formatting/style related
func (a *Analyzer) isStyleChange(diff string) bool {
	totalChanges := 0
	styleChanges := 0

	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			totalChanges++

			// Detect common style changes
			trimmed := strings.TrimSpace(line[1:])

			// Empty lines or whitespace only
			if trimmed == "" || len(strings.TrimSpace(trimmed)) == 0 {
				styleChanges++
				continue
			}

			// Comment formatting
			if strings.HasPrefix(trimmed, "//") {
				styleChanges++
				continue
			}

			// Import formatting
			if strings.Contains(trimmed, "import") {
				styleChanges++
				continue
			}

			// Bracket/brace only lines
			if trimmed == "{" || trimmed == "}" || trimmed == "(" || trimmed == ")" {
				styleChanges++
				continue
			}
		}
	}

	// If more than 70% of changes are style-related
	if totalChanges > 0 && float64(styleChanges)/float64(totalChanges) > 0.7 {
		return true
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
<<<<<<< HEAD
=======

// detectFunctions extracts function names from diff
func (a *Analyzer) detectFunctions(diff string) []string {
	var functions []string
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		// Look for Go function declarations
		if strings.Contains(line, "func ") {
			// Extract function name
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				funcLine := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "+"), "-"))
				if strings.HasPrefix(funcLine, "func ") {
					parts := strings.Fields(funcLine)
					if len(parts) >= 2 {
						funcName := strings.Split(parts[1], "(")[0]
						functions = append(functions, funcName)
					}
				}
			}
		}
	}
	return functions
}

// detectStructs extracts struct names from diff
func (a *Analyzer) detectStructs(diff string) []string {
	var structs []string
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		// Look for Go struct declarations
		if strings.Contains(line, "type ") && strings.Contains(line, "struct") {
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				structLine := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "+"), "-"))
				if strings.HasPrefix(structLine, "type ") {
					parts := strings.Fields(structLine)
					if len(parts) >= 2 {
						structName := parts[1]
						structs = append(structs, structName)
					}
				}
			}
		}
	}
	return structs
}

// detectMethods extracts method names from diff
func (a *Analyzer) detectMethods(diff string) []string {
	var methods []string
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		// Look for Go method declarations (func with receiver)
		if strings.Contains(line, "func (") {
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				methodLine := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "+"), "-"))
				if strings.HasPrefix(methodLine, "func (") {
					// Extract method name after receiver
					parts := strings.Split(methodLine, ")")
					if len(parts) >= 2 {
						methodPart := strings.TrimSpace(parts[1])
						methodName := strings.Split(methodPart, "(")[0]
						methodName = strings.TrimSpace(methodName)
						if methodName != "" {
							methods = append(methods, methodName)
						}
					}
				}
			}
		}
	}
	return methods
}

// detectChangePatterns identifies patterns in the changes
func (a *Analyzer) detectChangePatterns(change *parser.Change) []string {
	var patterns []string
	diff := change.Diff

	// Detect error handling additions
	if strings.Contains(diff, "+") && (strings.Contains(diff, "if err != nil") || strings.Contains(diff, "return err")) {
		patterns = append(patterns, "error-handling")
	}

	// Detect test additions
	if strings.HasSuffix(change.File, "_test.go") {
		if strings.Contains(diff, "+func Test") {
			patterns = append(patterns, "test-addition")
		}
	}

	// Detect import changes
	if strings.Contains(diff, "+import") || strings.Contains(diff, "-import") {
		patterns = append(patterns, "import-changes")
	}

	// Detect comment additions
	addedComments := 0
	removedComments := 0
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") && strings.Contains(line, "//") {
			addedComments++
		}
		if strings.HasPrefix(line, "-") && strings.Contains(line, "//") {
			removedComments++
		}
	}
	if addedComments > removedComments && addedComments >= 3 {
		patterns = append(patterns, "documentation")
	}

	// Detect refactoring (many removals and additions in same file)
	if change.Added > 10 && change.Removed > 10 {
		patterns = append(patterns, "refactoring")
	}

	// Detect configuration changes
	if strings.Contains(change.File, "config") || strings.HasSuffix(change.File, ".json") ||
		strings.HasSuffix(change.File, ".yaml") || strings.HasSuffix(change.File, ".yml") {
		patterns = append(patterns, "configuration")
	}

	// Detect API/endpoint changes
	if strings.Contains(diff, "http.") || strings.Contains(diff, "router.") ||
		strings.Contains(diff, "endpoint") || strings.Contains(diff, "handler") {
		patterns = append(patterns, "api-changes")
	}

	// Detect database changes
	if strings.Contains(diff, "sql") || strings.Contains(diff, "database") ||
		strings.Contains(diff, "query") || strings.Contains(diff, "gorm") {
		patterns = append(patterns, "database")
	}

	// Detect performance optimizations
	if strings.Contains(diff, "goroutine") || strings.Contains(diff, "sync.") ||
		strings.Contains(diff, "channel") || strings.Contains(diff, "concurrent") {
		patterns = append(patterns, "performance")
	}

	// Detect security additions
	if strings.Contains(diff, "auth") || strings.Contains(diff, "token") ||
		strings.Contains(diff, "security") || strings.Contains(diff, "crypto") {
		patterns = append(patterns, "security")
	}

	return patterns
}
>>>>>>> 1028df8 (fix(config): updated imports)
