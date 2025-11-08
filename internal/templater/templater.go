package templater

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
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
	// First, normalize the action
	actionUpper := strings.ToUpper(msg.Action)
	actionTemplates, ok := t.templates[actionUpper]
	if !ok {
		// For offline use, we have a more detailed fallback strategy
		fallbackActions := []string{"MISC", "A", "M"} // Priority order for fallbacks
		for _, fallback := range fallbackActions {
			if templates, exists := t.templates[fallback]; exists {
				actionTemplates = templates
				ok = true
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("no suitable templates found for action: %s (tried fallbacks: %v)", msg.Action, fallbackActions)
		}
	}

	// Topic selection with smart fallback for offline use
	var topicTemplates []string
	normalizedTopic := strings.ToLower(msg.Topic)

	// Try exact match first
	if templates, exists := actionTemplates[normalizedTopic]; exists && len(templates) > 0 {
		topicTemplates = templates
	} else {
		// Try fuzzy matching for similar topics
		for topic, templates := range actionTemplates {
			if strings.Contains(topic, normalizedTopic) || strings.Contains(normalizedTopic, topic) {
				topicTemplates = templates
				break
			}
		}

		// If no match found, fall back to _default
		if len(topicTemplates) == 0 {
			if defaults, exists := actionTemplates["_default"]; exists && len(defaults) > 0 {
				topicTemplates = defaults
			} else {
				return "", fmt.Errorf("no suitable templates found for topic: %s (action: %s)", msg.Topic, actionUpper)
			}
		}
	}

	// Prepare replacer for placeholders
	source := ""
	if len(msg.RenamedFiles) > 0 {
		source = msg.RenamedFiles[0].Source
	}
	target := ""
	if len(msg.RenamedFiles) > 0 {
		target = msg.RenamedFiles[0].Target
	}

	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", msg.Item,
		"{purpose}", msg.Purpose,
		"{source}", source,
		"{target}", target,
	)

	// Select a random template, avoiding recent duplicates
	var selectedTemplate string
	shuffledTemplates := make([]string, len(topicTemplates))
	copy(shuffledTemplates, topicTemplates)
	rand.Shuffle(len(shuffledTemplates), func(i, j int) {
		shuffledTemplates[i], shuffledTemplates[j] = shuffledTemplates[j], shuffledTemplates[i]
	})

	for _, tmpl := range shuffledTemplates {
		potentialMessage := replacer.Replace(tmpl)
		if !t.history.Contains(potentialMessage) {
			selectedTemplate = tmpl
			break
		}
	}

	if selectedTemplate == "" {
		// If all templates are recent duplicates, just pick a random one
		selectedTemplate = topicTemplates[rand.IntN(len(topicTemplates))]
	}

	formattedMsg := replacer.Replace(selectedTemplate)

	// Handle scope
	if msg.Scope != "" {
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

	return formattedMsg, nil
}
