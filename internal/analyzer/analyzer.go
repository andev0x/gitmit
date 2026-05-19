package analyzer

import (
	"bufio"
	"path/filepath"
	"regexp"
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
	Files             []string
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
func (a *Analyzer) AnalyzeChanges(totalAdded, totalRemoved int, branchName string) *CommitMessage {
	if len(a.changes) == 0 {
		return nil
	}

	commitMessage := &CommitMessage{
		TotalAdded:   totalAdded,
		TotalRemoved: totalRemoved,
	}

	var allFiles []string
	var allFileExtensions []string
	var allTopics []string
	var allPurposes []string
	var allItems []string
	var allFunctions []string
	var allStructs []string
	var allMethods []string
	var allPatterns []string

	for _, change := range a.changes {
		allFiles = append(allFiles, change.File)
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

	commitMessage.Files = uniqueStrings(allFiles)
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

	// Initialize a score tracker for the action (type)
	scoreMap := make(map[string]int)

	// Step 1: Scan the Branch status
	if branchName != "" {
		branchAction, branchScope := a.parseBranchName(branchName)
		if branchAction != "" {
			scoreMap[branchAction] += 3
		}
		if branchScope != "" {
			commitMessage.Scope = branchScope
		}
	}

	// Step 2: Add weights from diff stat ratio
	statAction := a.analyzeDiffStat(totalAdded, totalRemoved)
	if statAction != "" {
		scoreMap[statAction] += 2
	}

	// Step 3: Aggregate keyword scores
	keywordScores := a.calculateKeywordScores()
	for action, score := range keywordScores {
		scoreMap[action] += score
	}

	// Step 4: Add weights from multi-file patterns
	multiPatterns := a.detectMultiFilePatterns()
	for _, p := range multiPatterns {
		switch p {
		case "feature-addition":
			scoreMap["feat"] += 4
		case "bug-fix-cascade":
			scoreMap["fix"] += 4
		case "refactor-sweep":
			scoreMap["refactor"] += 3
		case "test-suite-update":
			scoreMap["test"] += 4
		}
	}

	// Step 5: Select the recommended type with the highest accumulated score
	bestAction := ""
	maxScore := -1
	for action, score := range scoreMap {
		if score > maxScore {
			maxScore = score
			bestAction = action
		}
	}

	if bestAction != "" {
		commitMessage.Action = bestAction
	} else {
		// Fallback to default action determination if no signals
		commitMessage.Action = a.determineAction(a.changes[0])
	}

	// Default analysis based on the first change if no specific fallback applies
	firstChange := a.changes[0]

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

	// NEW: Monitoring Dependency Changes (Dependency Watcher)
	newDeps := a.detectNewDependencies()
	if len(newDeps) > 0 {
		commitMessage.Action = "chore"
		commitMessage.Scope = "deps"
		commitMessage.Item = strings.Join(newDeps, ", ")
		commitMessage.Purpose = "update dependencies"
		return commitMessage // Priority return for dependency updates
	}

	// NEW: Learning from recent commit history (Commit History Consistency)
	if historyScope := a.analyzeHistoryScopes(); historyScope != "" {
		// Only override if scope is empty or "core"
		if commitMessage.Scope == "" || commitMessage.Scope == "core" {
			commitMessage.Scope = historyScope
		}
	}

	// Use commit history context to suggest consistent topics
	if commitMessage.Topic == "" || commitMessage.Topic == "core" {
		// Try to get topic from recent commit history
		if recentTopic := a.getRecentCommitTopic(); recentTopic != "" {
			commitMessage.Topic = recentTopic
		}
	}

	return commitMessage
}

// calculateKeywordScores analyzes git diff content and returns a map of scores for each action
func (a *Analyzer) calculateKeywordScores() map[string]int {
	actionScores := make(map[string]int)
	if len(a.config.Keywords) == 0 {
		return actionScores
	}

	// Concatenate all diffs
	var allDiffs strings.Builder
	for _, change := range a.changes {
		allDiffs.WriteString(change.Diff)
		allDiffs.WriteString("\n")
	}
	diffContent := strings.ToLower(allDiffs.String())

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

	return actionScores
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

	// Regex registry for functions
	patterns := map[string]*regexp.Regexp{
		"go":     regexp.MustCompile(`func\s+(?:\([^)]*\)\s+)?([A-Z][A-Za-z0-9]*)`),
		"ts":     regexp.MustCompile(`(?:function\s+([a-zA-Z0-9]*)|const\s+([a-zA-Z0-9]*)\s*=\s*(?:\([^)]*\)|[a-zA-Z0-9]*)\s*=>)`),
		"js":     regexp.MustCompile(`(?:function\s+([a-zA-Z0-9]*)|const\s+([a-zA-Z0-9]*)\s*=\s*(?:\([^)]*\)|[a-zA-Z0-9]*)\s*=>)`),
		"python": regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\(`),
		"java":   regexp.MustCompile(`(?:public|private|protected|static)\s+(?:[\w<>[\]]+\s+)+([a-zA-Z0-9_]+)\s*\(`),
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}

		cleanLine := strings.TrimPrefix(line, "+")

		for _, re := range patterns {
			matches := re.FindStringSubmatch(cleanLine)
			if len(matches) > 0 {
				// The first captured group (that is not empty) is the function name
				for i := 1; i < len(matches); i++ {
					if matches[i] != "" {
						functions = append(functions, matches[i])
						break
					}
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

	// Regex registry for structs/classes
	patterns := map[string]*regexp.Regexp{
		"go":     regexp.MustCompile(`type\s+([A-Z][A-Za-z0-9]*)\s+(?:struct|interface)`),
		"ts":     regexp.MustCompile(`class\s+([a-zA-Z0-9]*)`),
		"js":     regexp.MustCompile(`class\s+([a-zA-Z0-9]*)`),
		"python": regexp.MustCompile(`class\s+([a-zA-Z0-9_]+)\s*(?:\(|:)`),
		"java":   regexp.MustCompile(`(?:public|private|protected|abstract)?\s*class\s+([a-zA-Z0-9_]+)`),
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}

		cleanLine := strings.TrimPrefix(line, "+")

		for _, re := range patterns {
			matches := re.FindStringSubmatch(cleanLine)
			if len(matches) > 1 && matches[1] != "" {
				structs = append(structs, matches[1])
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

	// Structural Ratio Calculation
	ratio := float64(totalAdded) / float64(total)

	threshold := a.config.DiffStatThreshold
	if threshold == 0 {
		threshold = 0.5
	}

	// If deletions heavily dominate (Ratio < 0.2)
	if ratio < 0.2 {
		return "refactor"
	}

	// If additions heavily dominate (Ratio > 0.8)
	if ratio > 0.8 {
		// If many lines added, likely a feature
		if totalAdded > 30 {
			return "feat"
		}
	}

	// Balanced changes often indicate modifications or fixes
	if ratio >= 0.3 && ratio <= 0.7 {
		return "refactor"
	}

	return ""
}

// detectNewDependencies identifies newly added libraries in package management files
func (a *Analyzer) detectNewDependencies() []string {
	var newDeps []string
	depFiles := map[string]*regexp.Regexp{
		"go.mod":           regexp.MustCompile(`^\+\s+([^\s]+)\s+v`),
		"package.json":    regexp.MustCompile(`^\+\s+"([^"]+)":`),
		"requirements.txt": regexp.MustCompile(`^\+([a-zA-Z0-9\-_]+)==`),
		"Cargo.toml":      regexp.MustCompile(`^\+([a-zA-Z0-9\-_]+)\s+=`),
	}

	for _, change := range a.changes {
		fileName := filepath.Base(change.File)
		if re, ok := depFiles[fileName]; ok {
			scanner := bufio.NewScanner(strings.NewReader(change.Diff))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						newDeps = append(newDeps, matches[1])
					}
				}
			}
		}
	}
	return uniqueStrings(newDeps)
}

// parseBranchName extracts type and scope from branch name
func (a *Analyzer) parseBranchName(branch string) (string, string) {
	// Patterns like feature/auth-login or bugfix/fix-memleak
	// Format: <type>/<scope>-<description> or <type>/<description>
	parts := strings.Split(branch, "/")
	if len(parts) < 2 {
		return "", ""
	}

	branchType := strings.ToLower(parts[0])
	description := parts[1]

	action := ""
	switch branchType {
	case "feature", "feat":
		action = "feat"
	case "bugfix", "fix":
		action = "fix"
	case "hotfix":
		action = "fix"
	case "refactor":
		action = "refactor"
	case "chore":
		action = "chore"
	case "docs":
		action = "docs"
	case "style":
		action = "style"
	case "perf":
		action = "perf"
	case "test":
		action = "test"
	case "ci":
		action = "ci"
	case "build":
		action = "build"
	}

	scope := ""
	// Try to extract scope from description: scope-description or scope_description
	descParts := regexp.MustCompile(`[-_]`).Split(description, 2)
	if len(descParts) > 1 {
		scope = descParts[0]
	} else if len(description) > 0 {
		// If it's just feature/auth, then auth is the scope
		scope = description
	}

	return action, scope
}

// analyzeHistoryScopes analyzes the last 5 commits for common scopes
func (a *Analyzer) analyzeHistoryScopes() string {
	commits, err := history.GetRecentCommits(5)
	if err != nil || len(commits) == 0 {
		return ""
	}
	return a.calculateHistoryScope(commits)
}

// calculateHistoryScope calculates the most frequent scope from a list of commit messages
func (a *Analyzer) calculateHistoryScope(commits []string) string {
	scopeCounts := make(map[string]int)
	re := regexp.MustCompile(`^[a-z]+\(([^)]+)\):`)

	for _, msg := range commits {
		matches := re.FindStringSubmatch(msg)
		if len(matches) > 1 {
			scope := matches[1]
			scopeCounts[scope]++
		}
	}

	totalCommits := len(commits)
	if totalCommits == 0 {
		return ""
	}

	for scope, count := range scopeCounts {
		// If a single scope appears in more than 50% of the commits
		if float64(count)/float64(totalCommits) > 0.5 {
			return scope
		}
	}

	return ""
}
