package analyzer

import (
	"bufio"
	"path/filepath"
	"strings"

	"gitmit/internal/config"
	"gitmit/internal/history"
	"gitmit/internal/parser"
)

// CommitMessage represents the analyzed commit message components
type CommitMessage struct {
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

	// Apply diff stat analysis to infer intent based on added vs deleted lines
	action := a.analyzeDiffStat(totalAdded, totalRemoved)
	if action != "" {
		commitMessage.Action = action
	} else {
		// Use keyword scoring algorithm to determine the best action
		action = a.determineActionByKeywordScoring()
		if action != "" {
			commitMessage.Action = action
		} else {
			// Fallback to default action determination
			commitMessage.Action = a.determineAction(firstChange)
		}
	}

	// Determine other components
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

	// Use commit history context to suggest consistent topics
	if commitMessage.Topic == "" || commitMessage.Topic == "core" {
		// Try to get topic from recent commit history
		if recentTopic := a.getRecentCommitTopic(); recentTopic != "" {
			commitMessage.Topic = recentTopic
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

// determineActionByKeywordScoring analyzes git diff content and scores keywords to determine the best action
// This implements the keyword scoring algorithm requirement
func (a *Analyzer) determineActionByKeywordScoring() string {
	if len(a.config.Keywords) == 0 {
		return "" // No keywords configured, fall back to default logic
	}

	// Concatenate all diffs
	var allDiffs strings.Builder
	for _, change := range a.changes {
		allDiffs.WriteString(change.Diff)
		allDiffs.WriteString("\n")
	}
	diffContent := strings.ToLower(allDiffs.String())

	// Score each action based on keyword matches
	actionScores := make(map[string]int)

	for action, keywords := range a.config.Keywords {
		score := 0
		for keyword, weight := range keywords {
			keywordLower := strings.ToLower(keyword)
			// Count occurrences and multiply by weight
			occurrences := strings.Count(diffContent, keywordLower)
			score += occurrences * weight
		}
		actionScores[action] = score
	}

	// Find the action with the highest score
	maxScore := 0
	bestAction := ""
	for action, score := range actionScores {
		if score > maxScore {
			maxScore = score
			bestAction = action
		}
	}

	// Only return the action if the score is significant (> 0)
	if maxScore > 0 {
		return bestAction
	}

	return ""
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
	if change.FileExtension == "md" {
		return "docs"
	}
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
	// Apply custom topic mappings from config first
	for pattern, topic := range a.config.TopicMappings {
		if strings.Contains(path, pattern) {
			return topic
		}
	}

	parts := strings.Split(filepath.Dir(path), string(filepath.Separator))
	if len(parts) > 0 {
		// Prioritize "internal" or "pkg" subdirectories
		for i, p := range parts {
			if (p == "internal" || p == "pkg") && i+1 < len(parts) {
				return parts[i+1]
			}
		}
		// Fallback to the most specific directory name that is not a generic name
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "." && parts[i] != "src" {
				return parts[i]
			}
		}
		// If no specific topic is found, return the top-level directory
		if parts[0] != "." {
			return parts[0]
		}
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
		"login":       "authentication",
		"auth":        "authentication",
		"user":        "user management",
		"validate":    "validation",
		"validation":  "validation",
		"query":       "database query",
		"database":    "database operations",
		"cache":       "caching",
		"caching":     "caching",
		"refactor":    "code restructuring",
		"logging":     "logging",
		"logger":      "logging",
		"docs":        "documentation",
		"readme":      "documentation",
		"middleware":  "middleware",
		"test":        "testing",
		"tests":       "testing",
		"config":      "configuration",
		"ci":          "ci/cd",
		"log":         "logging",
		"sql":         "database logic",
		"gorm":        "database logic",
		"feat":        "new feature",
		"bug":         "bug fix",
		"fix":         "bug fix",
		"hotfix":      "bug fix",
		"cleanup":     "code cleanup",
		"perf":        "performance improvement",
		"performance": "performance improvement",
		"security":    "security update",
		"dep":         "dependency update",
		"dependency":  "dependency update",
		"build":       "build system",
		"style":       "code style",
		"serialize":   "serialization",
		"deserialize": "deserialization",
		"json":        "data handling",
		"xml":         "data handling",
		"async":       "asynchronous operations",
		"await":       "asynchronous operations",
		"concurrent":  "concurrency",
		"parallel":    "parallel processing",
		"api":         "api endpoints",
		"endpoint":    "api endpoints",
		"route":       "routing",
		"ui":          "user interface",
		"frontend":    "user interface",
		"backend":     "backend logic",
		"server":      "server logic",
		"client":      "client logic",
		"docker":      "docker configuration",
		"kubernetes":  "kubernetes configuration",
		"k8s":         "kubernetes configuration",
		"aws":         "aws integration",
		"gcp":         "gcp integration",
		"azure":       "azure integration",
		"error":       "error handling",
		"exception":   "error handling",
	}

	for keyword, purpose := range keywords {
		if strings.Contains(strings.ToLower(diff), keyword) {
			return purpose
		}
	}
	return "general update"
}

func (a *Analyzer) applySmartFallback(msg *CommitMessage) *CommitMessage {
	// If a new file is created, suggest "feat"
	if len(a.changes) == 1 && a.changes[0].Action == "A" {
		return &CommitMessage{Action: "feat", Topic: a.determineTopic(a.changes[0].File), Item: a.determineItem(a.changes[0].File), Purpose: "initial implementation"}
	}

	// If a file is deleted, suggest "chore" or "refactor"
	if len(a.changes) == 1 && a.changes[0].Action == "D" {
		return &CommitMessage{Action: "chore", Topic: a.determineTopic(a.changes[0].File), Item: a.determineItem(a.changes[0].File), Purpose: "remove unused file"}
	}

	// If a test file is modified, suggest "test"
	if len(a.changes) == 1 && strings.HasSuffix(a.changes[0].File, "_test.go") {
		return &CommitMessage{Action: "test", Topic: a.determineTopic(a.changes[0].File), Item: a.determineItem(a.changes[0].File), Purpose: "update tests"}
	}

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

// detectFunctions extracts function names from diff using language-aware regex
func (a *Analyzer) detectFunctions(diff string) []string {
	var functions []string
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()

		// Only look at added lines
		if !strings.HasPrefix(line, "+") {
			continue
		}

		cleanLine := strings.TrimSpace(strings.TrimPrefix(line, "+"))

		// Go functions
		if strings.HasPrefix(cleanLine, "func ") {
			// Extract function name: func FunctionName( or func (receiver) MethodName(
			if strings.Contains(cleanLine, "(") {
				// Check for method receiver
				if cleanLine[5] == '(' {
					// Method: func (r Receiver) MethodName
					parts := strings.SplitN(cleanLine[5:], ")", 2)
					if len(parts) == 2 {
						methodPart := strings.TrimSpace(parts[1])
						if idx := strings.Index(methodPart, "("); idx > 0 {
							methodName := strings.TrimSpace(methodPart[:idx])
							if methodName != "" {
								functions = append(functions, methodName)
							}
						}
					}
				} else {
					// Regular function: func FunctionName(
					parts := strings.Fields(cleanLine)
					if len(parts) >= 2 {
						funcName := strings.Split(parts[1], "(")[0]
						if funcName != "" {
							functions = append(functions, funcName)
						}
					}
				}
			}
		}

		// JavaScript/TypeScript functions
		if strings.Contains(cleanLine, "function ") {
			// function functionName( or function(
			idx := strings.Index(cleanLine, "function ")
			if idx >= 0 {
				remaining := cleanLine[idx+9:]
				if parenIdx := strings.Index(remaining, "("); parenIdx > 0 {
					funcName := strings.TrimSpace(remaining[:parenIdx])
					if funcName != "" && funcName != "function" {
						functions = append(functions, funcName)
					}
				}
			}
		}

		// Arrow functions: const funcName = () =>
		if strings.Contains(cleanLine, "=>") && (strings.Contains(cleanLine, "const ") || strings.Contains(cleanLine, "let ") || strings.Contains(cleanLine, "var ")) {
			// Extract: const funcName = ...
			for _, prefix := range []string{"const ", "let ", "var "} {
				if strings.Contains(cleanLine, prefix) {
					idx := strings.Index(cleanLine, prefix)
					remaining := cleanLine[idx+len(prefix):]
					if eqIdx := strings.Index(remaining, "="); eqIdx > 0 {
						funcName := strings.TrimSpace(remaining[:eqIdx])
						if funcName != "" {
							functions = append(functions, funcName)
						}
					}
					break
				}
			}
		}

		// Python functions
		if strings.HasPrefix(cleanLine, "def ") || strings.HasPrefix(cleanLine, "async def ") {
			// Extract: def function_name( or async def function_name(
			var remaining string
			if strings.HasPrefix(cleanLine, "async def ") {
				remaining = cleanLine[10:]
			} else {
				remaining = cleanLine[4:]
			}
			if parenIdx := strings.Index(remaining, "("); parenIdx > 0 {
				funcName := strings.TrimSpace(remaining[:parenIdx])
				if funcName != "" {
					functions = append(functions, funcName)
				}
			}
		}

		// Java/C/C++ methods
		// Pattern: public/private/protected Type methodName(
		if strings.Contains(cleanLine, "(") {
			for _, modifier := range []string{"public ", "private ", "protected ", "static "} {
				if strings.Contains(cleanLine, modifier) {
					parts := strings.Fields(cleanLine)
					// Find the part before (
					for _, part := range parts {
						if strings.Contains(part, "(") {
							funcName := strings.Split(part, "(")[0]
							if funcName != "" && funcName != "if" && funcName != "for" && funcName != "while" && funcName != "switch" {
								functions = append(functions, funcName)
								break
							}
						}
					}
					break
				}
			}
		}
	}
	return uniqueStrings(functions)
}

// detectStructs extracts struct/class names from diff using language-aware regex
func (a *Analyzer) detectStructs(diff string) []string {
	var structs []string
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()

		// Only look at added lines
		if !strings.HasPrefix(line, "+") {
			continue
		}

		cleanLine := strings.TrimSpace(strings.TrimPrefix(line, "+"))

		// Go structs and interfaces
		if strings.HasPrefix(cleanLine, "type ") && (strings.Contains(cleanLine, "struct") || strings.Contains(cleanLine, "interface")) {
			parts := strings.Fields(cleanLine)
			if len(parts) >= 2 {
				structName := parts[1]
				if structName != "" {
					structs = append(structs, structName)
				}
			}
		}

		// JavaScript/TypeScript classes
		if strings.HasPrefix(cleanLine, "class ") || strings.HasPrefix(cleanLine, "export class ") {
			var remaining string
			if strings.HasPrefix(cleanLine, "export class ") {
				remaining = cleanLine[13:]
			} else {
				remaining = cleanLine[6:]
			}

			// Extract class name (before space, { or extends)
			className := remaining
			for _, delimiter := range []string{" ", "{", "extends"} {
				if idx := strings.Index(className, delimiter); idx > 0 {
					className = className[:idx]
					break
				}
			}
			className = strings.TrimSpace(className)
			if className != "" {
				structs = append(structs, className)
			}
		}

		// Python classes
		if strings.HasPrefix(cleanLine, "class ") {
			remaining := cleanLine[6:]
			// Extract class name (before ( or :)
			className := remaining
			for _, delimiter := range []string{"(", ":"} {
				if idx := strings.Index(className, delimiter); idx > 0 {
					className = className[:idx]
					break
				}
			}
			className = strings.TrimSpace(className)
			if className != "" {
				structs = append(structs, className)
			}
		}

		// Java classes
		if strings.Contains(cleanLine, "class ") {
			for _, modifier := range []string{"public class ", "private class ", "protected class ", "abstract class "} {
				if strings.Contains(cleanLine, modifier) {
					idx := strings.Index(cleanLine, modifier)
					remaining := cleanLine[idx+len(modifier):]
					// Extract class name (before space, { or extends/implements)
					className := remaining
					for _, delimiter := range []string{" ", "{", "extends", "implements"} {
						if idx := strings.Index(className, delimiter); idx > 0 {
							className = className[:idx]
							break
						}
					}
					className = strings.TrimSpace(className)
					if className != "" {
						structs = append(structs, className)
					}
					break
				}
			}
		}
	}
	return uniqueStrings(structs)
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

	// Enhanced pattern detection for better context

	// Detect interface implementations
	if strings.Contains(diff, "+func (") && strings.Contains(diff, "interface") {
		patterns = append(patterns, "interface-implementation")
	}

	// Detect validation logic
	if strings.Contains(diff, "validate") || strings.Contains(diff, "Validate") ||
		(strings.Contains(diff, "if ") && strings.Contains(diff, "return")) {
		patterns = append(patterns, "validation")
	}

	// Detect logging enhancements
	if strings.Contains(diff, "+") && (strings.Contains(diff, "log.") ||
		strings.Contains(diff, "logger.") || strings.Contains(diff, "Logger")) {
		patterns = append(patterns, "logging")
	}

	// Detect middleware changes
	if strings.Contains(change.File, "middleware") ||
		(strings.Contains(diff, "func ") && strings.Contains(diff, "next")) {
		patterns = append(patterns, "middleware")
	}

	// Detect dependency injection setup
	if strings.Contains(diff, "New") && strings.Contains(diff, "return &") {
		patterns = append(patterns, "dependency-injection")
	}

	// Detect CLI/command changes
	if strings.Contains(diff, "cobra.Command") || strings.Contains(diff, "flag.") {
		patterns = append(patterns, "cli")
	}

	// Detect type definitions
	if strings.Contains(diff, "+type ") {
		patterns = append(patterns, "type-definition")
	}

	// Detect constant definitions
	if strings.Contains(diff, "+const ") || strings.Contains(diff, "+var ") {
		patterns = append(patterns, "constant-definition")
	}

	return patterns
}

// getRecentCommitTopic retrieves the topic/scope from the most recent commit
// This helps maintain consistency in commit history
func (a *Analyzer) getRecentCommitTopic() string {
	_, scope, err := history.GetRecentCommitContext()
	if err != nil || scope == "" {
		return ""
	}
	return scope
}

// analyzeDiffStat analyzes the ratio of added vs deleted lines to infer intent
// This implements the Diff Stat Analysis requirement
func (a *Analyzer) analyzeDiffStat(totalAdded, totalRemoved int) string {
	if totalAdded == 0 && totalRemoved == 0 {
		return ""
	}

	total := totalAdded + totalRemoved
	if total == 0 {
		return ""
	}

	deletedRatio := float64(totalRemoved) / float64(total)
	addedRatio := float64(totalAdded) / float64(total)

	threshold := a.config.DiffStatThreshold
	if threshold == 0 {
		threshold = 0.5
	}

	// If deleted lines dominate, suggest cleanup or refactor
	if deletedRatio > threshold+0.2 { // More than 70% deletions
		return "refactor"
	}

	// If a large number of lines are added with minimal deletions, suggest feat
	if addedRatio > threshold+0.2 && totalAdded > 50 {
		// Check if it's a new file addition
		for _, change := range a.changes {
			if change.Action == "A" && change.Added > 30 {
				return "feat"
			}
		}
	}

	// Balanced changes often indicate modifications or fixes
	if deletedRatio > 0.3 && addedRatio > 0.3 {
		return "refactor"
	}

	return ""
}
