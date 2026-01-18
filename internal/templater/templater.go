package templater

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitmit/internal/analyzer"
	"gitmit/internal/history"
)

//go:embed templates.json
var embeddedTemplates embed.FS

// Templates holds the loaded commit message templates
type Templates map[string]map[string][]string

// Templater is responsible for selecting and formatting commit messages
type Templater struct {
	templates Templates
	history   *history.CommitHistory
}

// NewTemplater creates a new Templater
func NewTemplater(templateFile string, hist *history.CommitHistory) (*Templater, error) {
	var data []byte
	var err error

	// For offline use, try loading from multiple locations in order:
	// 1. Current working directory
	// 2. Executable's directory
	// 3. Embedded templates

	// Try current working directory first
	pwd, _ := os.Getwd()
	localPath := filepath.Join(pwd, templateFile)
	data, err = os.ReadFile(localPath)

	// If not found in current directory, try executable's directory
	if err != nil || len(data) == 0 {
		execPath, execErr := os.Executable()
		if execErr == nil {
			execDir := filepath.Dir(execPath)
			execLocalPath := filepath.Join(execDir, templateFile)
			data, err = os.ReadFile(execLocalPath)
		}
	}

	// Finally, try embedded templates
	if err != nil || len(data) == 0 {
		data, err = embeddedTemplates.ReadFile(templateFile)
		if err != nil {
			return nil, fmt.Errorf("error reading templates: tried current directory (%s), executable directory, and embedded templates", localPath)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("no valid templates found in any location")
		}
	}

	var templates Templates
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling template file: %w", err)
	}

	// Comprehensive template validation for offline use
	requiredActions := []string{"A", "M", "D", "R", "MISC"}
	missingActions := []string{}

	for _, action := range requiredActions {
		actionTemplates, ok := templates[action]
		if !ok {
			missingActions = append(missingActions, action)
			continue
		}

		// Validate that each action has _default templates
		if defaultTemplates, ok := actionTemplates["_default"]; !ok || len(defaultTemplates) == 0 {
			return nil, fmt.Errorf("template validation failed: action '%s' missing required '_default' templates", action)
		}

		// Validate that templates are properly formatted
		for topic, messages := range actionTemplates {
			if len(messages) == 0 {
				return nil, fmt.Errorf("template validation failed: action '%s', topic '%s' has no templates", action, topic)
			}

			// Check for valid placeholder format in each template
			for _, tmpl := range messages {
				if strings.Count(tmpl, "{") != strings.Count(tmpl, "}") {
					return nil, fmt.Errorf("template validation failed: mismatched placeholder braces in template: %s", tmpl)
				}
			}
		}
	}

	if len(missingActions) > 0 {
		return nil, fmt.Errorf("template validation failed: missing required actions: %v", missingActions)
	}

	// No need to seed in Go 1.20+ as it's automatically handled

	return &Templater{templates: templates, history: hist}, nil
}

// GetMessage selects and formats a commit message
func (t *Templater) GetMessage(msg *analyzer.CommitMessage) (string, error) {
	// Check if this is a special file that needs dedicated handling
	specialGroup := resolveSpecialFile(msg)
	var actionKey string

	if specialGroup != "" {
		// Force use of special template group
		actionKey = specialGroup
	} else {
		// Map analyzer action names (feat, fix, refactor, chore, docs, test, etc.)
		// to the template groups used in templates.json (A, M, D, R, DOC, MISC)
		actionMap := map[string]string{
			"feat":     "A",
			"add":      "A",
			"fix":      "M",
			"bugfix":   "M",
			"refactor": "R",
			"chore":    "D",
			"test":     "M",
			"docs":     "DOC",
			"ci":       "M",
			"perf":     "M",
			"style":    "MISC",
			"build":    "MISC",
			"security": "SECURITY",
		}

		// Normalize and resolve action group
		actionLower := strings.ToLower(msg.Action)
		if key, ok := actionMap[actionLower]; ok {
			actionKey = key
		} else if len(msg.Action) == 1 {
			// Already a single-letter action like A/M/D/R
			actionKey = strings.ToUpper(msg.Action)
		} else {
			// fallback to MISC
			actionKey = "MISC"
		}
	}

	actionTemplates, ok := t.templates[actionKey]
	if !ok {
		// Try fallbacks: specific order prefers DOC then A then M then MISC
		fallbackActions := []string{"DOC", "A", "M", "R", "D", "MISC"}
		for _, fb := range fallbackActions {
			if templates, exists := t.templates[fb]; exists {
				actionTemplates = templates
				ok = true
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("no suitable templates found for action: %s (resolved key: %s)", msg.Action, actionKey)
		}
	}

	// Topic selection with improved matching and weighting
	normalizedTopic := strings.ToLower(strings.TrimSpace(msg.Topic))
	var topicTemplates []string

	// exact match
	if normalizedTopic != "" {
		if templates, exists := actionTemplates[normalizedTopic]; exists && len(templates) > 0 {
			topicTemplates = templates
		}
	}

	// fuzzy match if exact not found
	if len(topicTemplates) == 0 {
		for topic, templates := range actionTemplates {
			if topic == "_default" {
				continue
			}
			tname := strings.ToLower(topic)
			if normalizedTopic != "" && (strings.Contains(tname, normalizedTopic) || strings.Contains(normalizedTopic, tname)) {
				topicTemplates = templates
				break
			}
		}
	}

	// fall back to _default
	if len(topicTemplates) == 0 {
		if defaults, exists := actionTemplates["_default"]; exists && len(defaults) > 0 {
			topicTemplates = defaults
		} else {
			return "", fmt.Errorf("no suitable templates found for topic: %s (action: %s)", msg.Topic, actionKey)
		}
	}

	// Prepare placeholder values
	source := ""
	target := ""
	if len(msg.RenamedFiles) > 0 {
		source = msg.RenamedFiles[0].Source
		target = msg.RenamedFiles[0].Target
	}

	// Enhanced item selection based on detected structures
	item := msg.Item
	if len(msg.DetectedFunctions) > 0 {
		item = msg.DetectedFunctions[0]
	} else if len(msg.DetectedStructs) > 0 {
		item = msg.DetectedStructs[0]
	} else if len(msg.DetectedMethods) > 0 {
		item = msg.DetectedMethods[0]
	}

	// Scoring-based selection: prefer templates that use available context
	type scored struct {
		tmpl  string
		score float64
	}

	var candidates []scored

	for _, tmpl := range topicTemplates {
		score := 0.0

		// Core placeholder rewards
		if strings.Contains(tmpl, "{item}") && item != "" {
			score += 3.0
		}
		if strings.Contains(tmpl, "{purpose}") && msg.Purpose != "" && msg.Purpose != "general update" {
			score += 2.5
		}
		if strings.Contains(tmpl, "{source}") && source != "" {
			score += 3.0
		}
		if strings.Contains(tmpl, "{target}") && target != "" {
			score += 3.0
		}
		if strings.Contains(tmpl, "{topic}") && msg.Topic != "" {
			score += 1.5
		}

		// Context-aware bonuses

		// Pattern matching bonus
		for _, pattern := range msg.ChangePatterns {
			patternKeywords := map[string][]string{
				"error-handling":           {"fix", "error", "handle"},
				"test-addition":            {"test", "coverage"},
				"documentation":            {"docs", "document", "comment"},
				"api-changes":              {"api", "endpoint", "route"},
				"database":                 {"db", "database", "query", "schema"},
				"security":                 {"security", "auth", "token"},
				"performance":              {"perf", "optimize", "speed"},
				"validation":               {"validat", "check", "verify"},
				"logging":                  {"log", "trace", "debug"},
				"middleware":               {"middleware", "chain"},
				"interface-implementation": {"implement", "interface"},
				"cli":                      {"command", "flag", "cli"},
			}

			if keywords, exists := patternKeywords[pattern]; exists {
				for _, keyword := range keywords {
					if strings.Contains(strings.ToLower(tmpl), keyword) {
						score += 2.0
						break
					}
				}
			}
		}

		// File type context bonus
		for _, ext := range msg.FileExtensions {
			extKeywords := map[string][]string{
				"go":   {"func", "method", "type"},
				"json": {"config", "setting", "format"},
				"yaml": {"config", "setting"},
				"yml":  {"config", "setting"},
				"md":   {"docs", "document"},
				"sql":  {"database", "query"},
			}

			if keywords, exists := extKeywords[ext]; exists {
				for _, keyword := range keywords {
					if strings.Contains(strings.ToLower(tmpl), keyword) {
						score += 1.0
						break
					}
				}
			}
		}

		// Detected structure bonus
		if len(msg.DetectedFunctions) > 0 && strings.Contains(strings.ToLower(tmpl), "func") {
			score += 1.5
		}
		if len(msg.DetectedStructs) > 0 && strings.Contains(strings.ToLower(tmpl), "type") {
			score += 1.5
		}
		if len(msg.DetectedMethods) > 0 && strings.Contains(strings.ToLower(tmpl), "method") {
			score += 1.5
		}

		// Major change bonus
		if msg.IsMajor && (strings.Contains(strings.ToLower(tmpl), "restructure") ||
			strings.Contains(strings.ToLower(tmpl), "refactor") ||
			strings.Contains(strings.ToLower(tmpl), "major")) {
			score += 2.0
		}

		// Special case bonuses
		if msg.IsDocsOnly && strings.Contains(strings.ToLower(tmpl), "doc") {
			score += 2.5
		}
		if msg.IsConfigOnly && strings.Contains(strings.ToLower(tmpl), "config") {
			score += 2.5
		}
		if msg.IsDepsOnly && strings.Contains(strings.ToLower(tmpl), "dep") {
			score += 2.5
		}

		// Penalty for generic templates when specific context exists
		isGeneric := strings.Contains(strings.ToLower(tmpl), "general") ||
			strings.Contains(strings.ToLower(tmpl), "update")
		if isGeneric && (len(msg.ChangePatterns) > 0 || msg.Purpose != "general update") {
			score -= 1.0
		}

		// Bonus for templates that match project scope
		projectScope := inferProjectScope(msg)
		if projectScope != "" {
			templateLower := strings.ToLower(tmpl)
			scopeLower := strings.ToLower(projectScope)

			// Direct scope mention in template
			if strings.Contains(templateLower, scopeLower) {
				score += 1.5
			}

			// Topic placeholder with meaningful scope
			if strings.Contains(tmpl, "{topic}") && msg.Topic != "" && msg.Topic != "core" {
				score += 1.0
			}
		}

		// Small randomness for variety (0-0.5)
		score += rand.Float64() * 0.5

		candidates = append(candidates, scored{tmpl: tmpl, score: score})
	}

	// Sort candidates by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// Get best candidates (top scorers)
	bestScore := -1.0
	var bestCandidates []string
	for _, c := range candidates {
		if bestScore < 0 {
			bestScore = c.score
			bestCandidates = []string{c.tmpl}
		} else if c.score >= bestScore-0.5 { // Allow slight variance in "best"
			bestCandidates = append(bestCandidates, c.tmpl)
		} else {
			break // Scores are sorted, no need to continue
		}
	}

	// Prefer a template that is not in recent history
	replacerForCheck := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	var chosen string
	for _, tmpl := range bestCandidates {
		candidateMsg := replacerForCheck.Replace(tmpl)
		if !t.history.Contains(candidateMsg) {
			chosen = tmpl
			break
		}
	}

	// If all best candidates are in history, pick a random best candidate
	if chosen == "" {
		if len(bestCandidates) > 0 {
			chosen = bestCandidates[rand.Intn(len(bestCandidates))]
		} else {
			// final fallback: random from topicTemplates
			chosen = topicTemplates[rand.Intn(len(topicTemplates))]
		}
	}

	// Final replacement
	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	formattedMsg := replacer.Replace(chosen)

	// Infer and apply project scope for better context
	projectScope := inferProjectScope(msg)
	if projectScope != "" {
		// Try common scope patterns
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+projectScope+")", 1)
	}

	// If scope exists in message, prefer replacing the topic scope pattern when present
	if msg.Scope != "" {
		// try common patterns
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

	// Clean and normalize the final message
	formattedMsg = cleanFinalMessage(formattedMsg)

	return formattedMsg, nil
}

// GetSuggestions returns multiple commit message suggestions ranked by context matching
func (t *Templater) GetSuggestions(msg *analyzer.CommitMessage, maxSuggestions int) ([]string, error) {
	actionKey, candidates := t.DebugInfo(msg)
	if candidates == nil || len(candidates) == 0 {
		return nil, fmt.Errorf("no templates found for action: %s", actionKey)
	}

	// Score all candidates
	type scoredTemplate struct {
		template string
		score    float64
	}

	var scored []scoredTemplate

	// Prepare placeholder values
	source := ""
	target := ""
	if len(msg.RenamedFiles) > 0 {
		source = msg.RenamedFiles[0].Source
		target = msg.RenamedFiles[0].Target
	}

	for _, tmpl := range candidates {
		score := 0.0

		// Use the comprehensive scoring function
		score = t.scoreTemplate(tmpl, msg)

		// Core placeholder rewards (additional specific bonuses)
		if strings.Contains(tmpl, "{item}") && msg.Item != "" {
			score += 1.0
		}
		if strings.Contains(tmpl, "{purpose}") && msg.Purpose != "" && msg.Purpose != "general update" {
			score += 1.0
		}
		if strings.Contains(tmpl, "{source}") && source != "" {
			score += 1.5
		}
		if strings.Contains(tmpl, "{target}") && target != "" {
			score += 1.5
		}
		if strings.Contains(tmpl, "{topic}") && msg.Topic != "" {
			score += 0.5
		}

		// Small randomness for variety (0-1)
		score += rand.Float64()

		scored = append(scored, scoredTemplate{tmpl, score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Get top N suggestions
	suggestions := make([]string, 0, maxSuggestions)
	usedMessages := make(map[string]bool)

	// Enhanced item selection based on detected structures
	item := msg.Item
	if len(msg.DetectedFunctions) > 0 {
		item = msg.DetectedFunctions[0]
	} else if len(msg.DetectedStructs) > 0 {
		item = msg.DetectedStructs[0]
	} else if len(msg.DetectedMethods) > 0 {
		item = msg.DetectedMethods[0]
	}

	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	// Take top scored templates until we have enough unique messages
	for _, s := range scored {
		if len(suggestions) >= maxSuggestions {
			break
		}

		message := replacer.Replace(s.template)
		message = cleanFinalMessage(message) // Clean the message

		// Skip if we've seen this exact message or it's in history
		if usedMessages[message] || t.history.Contains(message) {
			continue
		}

		suggestions = append(suggestions, message)
		usedMessages[message] = true
	}

	// If we don't have enough suggestions, include some that might be in history
	if len(suggestions) < maxSuggestions {
		for _, s := range scored {
			if len(suggestions) >= maxSuggestions {
				break
			}

			message := replacer.Replace(s.template)
			message = cleanFinalMessage(message) // Clean the message
			if !usedMessages[message] {
				suggestions = append(suggestions, message)
				usedMessages[message] = true
			}
		}
	}

	return suggestions, nil
}

// DebugInfo returns the resolved action key and the candidate templates for a CommitMessage
func (t *Templater) DebugInfo(msg *analyzer.CommitMessage) (string, []string) {
	// same mapping as in GetMessage
	actionMap := map[string]string{
		"feat":     "A",
		"add":      "A",
		"fix":      "M",
		"bugfix":   "M",
		"refactor": "R",
		"chore":    "D",
		"test":     "M",
		"docs":     "DOC",
		"ci":       "M",
		"perf":     "M",
		"style":    "MISC",
		"build":    "MISC",
		"security": "SECURITY",
	}

	actionLower := strings.ToLower(msg.Action)
	var actionKey string
	if key, ok := actionMap[actionLower]; ok {
		actionKey = key
	} else if len(msg.Action) == 1 {
		actionKey = strings.ToUpper(msg.Action)
	} else {
		actionKey = "MISC"
	}

	actionTemplates, ok := t.templates[actionKey]
	if !ok {
		fallbackActions := []string{"DOC", "A", "M", "R", "D", "MISC"}
		for _, fb := range fallbackActions {
			if templates, exists := t.templates[fb]; exists {
				actionTemplates = templates
				ok = true
				break
			}
		}
		if !ok {
			return actionKey, nil
		}
	}

	normalizedTopic := strings.ToLower(strings.TrimSpace(msg.Topic))
	var topicTemplates []string
	if normalizedTopic != "" {
		if templates, exists := actionTemplates[normalizedTopic]; exists && len(templates) > 0 {
			topicTemplates = templates
		}
	}
	if len(topicTemplates) == 0 {
		for topic, templates := range actionTemplates {
			if topic == "_default" {
				continue
			}
			tname := strings.ToLower(topic)
			if normalizedTopic != "" && (strings.Contains(tname, normalizedTopic) || strings.Contains(normalizedTopic, tname)) {
				topicTemplates = templates
				break
			}
		}
	}
	if len(topicTemplates) == 0 {
		if defaults, exists := actionTemplates["_default"]; exists && len(defaults) > 0 {
			topicTemplates = defaults
		}
	}

	return actionKey, topicTemplates
}

// scoreTemplate scores a template based on how well it matches the commit message context
func (t *Templater) scoreTemplate(template string, msg *analyzer.CommitMessage) float64 {
	score := 0.0

	// Base score
	score += 1.0

	// PENALTY MECHANISM: Heavy penalty for templates requiring {item} but no data available
	if strings.Contains(template, "{item}") {
		// Check if we have any item data
		hasItem := msg.Item != ""
		hasDetectedStructures := len(msg.DetectedFunctions) > 0 ||
			len(msg.DetectedStructs) > 0 ||
			len(msg.DetectedMethods) > 0

		if !hasItem && !hasDetectedStructures {
			// Deduct 50 points - this template will never be selected
			score -= 50.0
		}
	}

	// Bonus for templates that match detected patterns
	for _, pattern := range msg.ChangePatterns {
		if strings.Contains(template, pattern) ||
			(pattern == "error-handling" && strings.Contains(template, "fix")) ||
			(pattern == "test-addition" && strings.Contains(template, "test")) ||
			(pattern == "documentation" && strings.Contains(template, "docs")) ||
			(pattern == "api-changes" && strings.Contains(template, "api")) ||
			(pattern == "database" && strings.Contains(template, "db")) ||
			(pattern == "security" && strings.Contains(template, "security")) ||
			(pattern == "performance" && strings.Contains(template, "perf")) {
			score += 2.0
		}
	}

	// Bonus for templates that use detected structures
	if len(msg.DetectedFunctions) > 0 && strings.Contains(template, "{item}") {
		score += 1.5
	}
	if len(msg.DetectedStructs) > 0 && strings.Contains(template, "{item}") {
		score += 1.5
	}

	// Bonus for templates with purpose placeholder when we have a good purpose
	if msg.Purpose != "general update" && strings.Contains(template, "{purpose}") {
		score += 1.0
	}

	// Bonus for templates that match file type context
	for _, ext := range msg.FileExtensions {
		if ext == "go" && strings.Contains(template, "func") {
			score += 0.5
		}
		if (ext == "json" || ext == "yaml" || ext == "yml") &&
			(strings.Contains(template, "config") || strings.Contains(template, "settings")) {
			score += 1.0
		}
		if ext == "md" && strings.Contains(template, "docs") {
			score += 1.5
		}
	}

	// Penalty for generic templates when we have specific information
	if strings.Contains(template, "general") && len(msg.ChangePatterns) > 0 {
		score -= 0.5
	}

	// Bonus for templates matching major changes
	if msg.IsMajor && (strings.Contains(template, "restructure") ||
		strings.Contains(template, "refactor") || strings.Contains(template, "major")) {
		score += 1.0
	}

	// Bonus for templates that match the project scope
	projectScope := inferProjectScope(msg)
	if projectScope != "" {
		templateLower := strings.ToLower(template)
		scopeLower := strings.ToLower(projectScope)

		// Direct scope mention in template
		if strings.Contains(templateLower, scopeLower) {
			score += 2.0
		}

		// Scope matches topic placeholder usage
		if strings.Contains(template, "{topic}") && msg.Topic != "" {
			score += 1.0
		}
	}

	return score
}

// GetAlternativeSuggestion generates a new commit message avoiding already used suggestions
// It uses intelligent variation algorithms including:
// - Context-aware scoring for relevance
// - Similarity detection to ensure diversity
// - History tracking to avoid repetition
// - Weighted randomization for variety
func (t *Templater) GetAlternativeSuggestion(msg *analyzer.CommitMessage, usedSuggestions map[string]bool) (string, error) {
	// Get all candidate templates using the same logic as GetSuggestions
	actionKey, candidates := t.DebugInfo(msg)
	if candidates == nil || len(candidates) == 0 {
		return "", fmt.Errorf("no templates found for action: %s", actionKey)
	}

	// Prepare placeholder values
	source := ""
	target := ""
	if len(msg.RenamedFiles) > 0 {
		source = msg.RenamedFiles[0].Source
		target = msg.RenamedFiles[0].Target
	}

	// Enhanced item selection based on detected structures
	item := msg.Item
	if len(msg.DetectedFunctions) > 0 {
		item = msg.DetectedFunctions[0]
	} else if len(msg.DetectedStructs) > 0 {
		item = msg.DetectedStructs[0]
	} else if len(msg.DetectedMethods) > 0 {
		item = msg.DetectedMethods[0]
	}

	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	// Score all candidates and sort by relevance with diversity bonus
	type scoredTemplate struct {
		template string
		message  string
		score    float64
	}

	var scored []scoredTemplate

	for _, tmpl := range candidates {
		message := replacer.Replace(tmpl)
		message = cleanFinalMessage(message) // Clean the message

		// Skip if already used
		if usedSuggestions[message] {
			continue
		}

		// Calculate base score using context matching
		score := t.scoreTemplate(tmpl, msg)

		// Add diversity bonus - prefer templates with different structure/wording
		diversityBonus := 0.0
		for usedMsg := range usedSuggestions {
			// Calculate simple similarity (Levenshtein-like heuristic)
			similarity := calculateSimilarity(message, usedMsg)
			if similarity < 0.5 {
				diversityBonus += 1.0
			} else if similarity < 0.7 {
				diversityBonus += 0.5
			}
		}
		score += diversityBonus

		// Small random factor for variety (0-1)
		score += rand.Float64()

		scored = append(scored, scoredTemplate{tmpl, message, score})
	}

	if len(scored) == 0 {
		// If all have been used, reset and try again with lower standards
		for _, tmpl := range candidates {
			message := replacer.Replace(tmpl)
			message = cleanFinalMessage(message) // Clean the message
			score := t.scoreTemplate(tmpl, msg) + rand.Float64()
			scored = append(scored, scoredTemplate{tmpl, message, score})
		}
	}

	if len(scored) == 0 {
		return "", fmt.Errorf("no alternative suggestions available")
	}

	// Sort by score descending and return the top one
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored[0].message, nil
}

// resolveSpecialFile detects special files like LICENSE, COPYING, .md docs, etc.
// Returns the special template group to use, or empty string if not a special file
func resolveSpecialFile(msg *analyzer.CommitMessage) string {
	// Check if all files are markdown documentation files
	if msg.IsDocsOnly {
		return "DOC"
	}

	// Check for .md file extensions - these are documentation
	for _, ext := range msg.FileExtensions {
		if ext == "md" {
			return "DOC"
		}
	}

	// Define special file patterns
	specialFiles := map[string]string{
		"license":      "LICENSE",
		"copying":      "LICENSE",
		"copyright":    "LICENSE",
		"readme":       "DOC",
		"changelog":    "DOC",
		"authors":      "DOC",
		"contributors": "DOC",
	}

	// Check topic and item for special file indicators
	topicLower := strings.ToLower(msg.Topic)
	itemLower := strings.ToLower(msg.Item)

	for pattern, group := range specialFiles {
		if strings.Contains(topicLower, pattern) || strings.Contains(itemLower, pattern) {
			return group
		}
	}

	return ""
}

// cleanFinalMessage post-processes the commit message to normalize format
func cleanFinalMessage(message string) string {
	// Normalize empty parentheses patterns like "feat():" to "feat:"
	message = strings.ReplaceAll(message, "():", ":")
	message = strings.ReplaceAll(message, "( ):", ":")
	message = strings.ReplaceAll(message, "(  ):", ":")

	// Remove excessive whitespace
	message = strings.TrimSpace(message)

	// Normalize multiple spaces to single space
	for strings.Contains(message, "  ") {
		message = strings.ReplaceAll(message, "  ", " ")
	}

	return message
}

// inferProjectScope extracts the project scope from file paths
// This helps match commit messages to the affected module/component
func inferProjectScope(msg *analyzer.CommitMessage) string {
	// If the analyzer already detected a scope, use it
	if msg.Scope != "" {
		return msg.Scope
	}

	// Use the topic as the scope if it's meaningful
	if msg.Topic != "" && msg.Topic != "core" && msg.Topic != "." {
		return msg.Topic
	}

	return ""
}

// calculateSimilarity returns a similarity score between 0.0 (completely different) and 1.0 (identical)
// Uses a hybrid approach combining:
// - Word-level Jaccard similarity (60% weight) - measures semantic overlap
// - Character-level position matching (40% weight) - measures structural similarity
// This ensures both content and structure are considered when determining diversity
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Normalize strings
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	// Word-level comparison
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	// Count common words
	wordSet1 := make(map[string]bool)
	for _, w := range words1 {
		wordSet1[w] = true
	}

	commonWords := 0
	for _, w := range words2 {
		if wordSet1[w] {
			commonWords++
		}
	}

	// Calculate Jaccard similarity
	totalUniqueWords := len(wordSet1)
	for _, w := range words2 {
		if !wordSet1[w] {
			totalUniqueWords++
		}
	}

	if totalUniqueWords == 0 {
		return 0.0
	}

	wordSimilarity := float64(commonWords) / float64(totalUniqueWords)

	// Character-level comparison (Levenshtein distance approximation)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	// Simple character overlap
	charMatches := 0
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] == s2[i] {
			charMatches++
		}
	}

	charSimilarity := float64(charMatches) / float64(maxLen)

	// Weighted combination (60% word-based, 40% character-based)
	return 0.6*wordSimilarity + 0.4*charSimilarity
}
