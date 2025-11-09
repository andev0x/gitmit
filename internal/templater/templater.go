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
	}

	// Normalize and resolve action group
	actionLower := strings.ToLower(msg.Action)
	var actionKey string
	if key, ok := actionMap[actionLower]; ok {
		actionKey = key
	} else if len(msg.Action) == 1 {
		// Already a single-letter action like A/M/D/R
		actionKey = strings.ToUpper(msg.Action)
	} else {
		// fallback to MISC
		actionKey = "MISC"
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

	// Scoring-based selection: prefer templates that use available context
	type scored struct {
		tmpl  string
		score int
	}

	var candidates []scored

	for _, tmpl := range topicTemplates {
		score := 0
		// reward templates that include placeholders we can fill
		if strings.Contains(tmpl, "{item}") && msg.Item != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{purpose}") && msg.Purpose != "" && msg.Purpose != "general update" {
			score += 2
		}
		if strings.Contains(tmpl, "{source}") && source != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{target}") && target != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{topic}") && normalizedTopic != "" {
			score += 1
		}
		// small randomness to diversify choices
		score += rand.Intn(2)

		candidates = append(candidates, scored{tmpl: tmpl, score: score})
	}

	// sort candidates by score (simple selection of best score)
	bestScore := -1
	var bestCandidates []string
	for _, c := range candidates {
		if c.score > bestScore {
			bestScore = c.score
			bestCandidates = []string{c.tmpl}
		} else if c.score == bestScore {
			bestCandidates = append(bestCandidates, c.tmpl)
		}
	}

	// Prefer a template that is not in recent history
	replacerForCheck := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", msg.Item,
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
		"{item}", msg.Item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	formattedMsg := replacer.Replace(chosen)

	// If scope exists, prefer replacing the topic scope pattern when present
	if msg.Scope != "" {
		// try common patterns
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

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
		score    int
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
		score := 0

		// Core context matching
		if strings.Contains(tmpl, "{item}") && msg.Item != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{purpose}") && msg.Purpose != "" && msg.Purpose != "general update" {
			score += 2
		}
		if strings.Contains(tmpl, "{source}") && source != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{target}") && target != "" {
			score += 3
		}
		if strings.Contains(tmpl, "{topic}") && msg.Topic != "" {
			score += 1
		}

		// Additional heuristics
		if msg.IsDocsOnly && strings.Contains(strings.ToLower(tmpl), "doc") {
			score += 2
		}
		if msg.IsConfigOnly && strings.Contains(strings.ToLower(tmpl), "config") {
			score += 2
		}
		if msg.IsDepsOnly && strings.Contains(strings.ToLower(tmpl), "dep") {
			score += 2
		}

		// File type bonus
		for _, ext := range msg.FileExtensions {
			if strings.Contains(strings.ToLower(tmpl), ext) {
				score++
			}
		}

		// Small randomness for variety
		score += rand.Intn(2)

		scored = append(scored, scoredTemplate{tmpl, score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Get top N suggestions
	suggestions := make([]string, 0, maxSuggestions)
	usedMessages := make(map[string]bool)

	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", msg.Item,
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
