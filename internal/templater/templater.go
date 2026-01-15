package templater

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Try loading from the executable's directory first
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		localTemplatePath := filepath.Join(execDir, templateFile)
		data, err = ioutil.ReadFile(localTemplatePath)
	}

	// If local file not found or error, use embedded templates
	if err != nil || len(data) == 0 {
		data, err = embeddedTemplates.ReadFile(templateFile)
		if err != nil {
			return nil, fmt.Errorf("error reading embedded template file: %w", err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("no templates found, neither local nor embedded")
		}
	}

	var templates Templates
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling template file: %w", err)
	}

	// Basic validation: ensure essential actions and _default exist
	if _, ok := templates["A"]; !ok {
		return nil, fmt.Errorf("template validation failed: missing 'A' action templates")
	}
	if _, ok := templates["M"]; !ok {
		return nil, fmt.Errorf("template validation failed: missing 'M' action templates")
	}
	if _, ok := templates["D"]; !ok {
		return nil, fmt.Errorf("template validation failed: missing 'D' action templates")
	}
	// Add more validation as needed

	rand.Seed(time.Now().UnixNano())

	return &Templater{templates: templates, history: hist}, nil
}

// GetMessage selects and formats a commit message
func (t *Templater) GetMessage(msg *analyzer.CommitMessage) (string, error) {
	// Map action to uppercase for template lookup
	actionUpper := strings.ToUpper(msg.Action)

	// Handle special actions that might not match template keys
	switch actionUpper {
	case "FEAT":
		actionUpper = "A" // Map to Add templates
	case "FIX", "REFACTOR", "PERF":
		actionUpper = "M" // Map to Modify templates
	case "CHORE":
		if msg.Action == "chore" {
			// Check if it's a deletion
			deletions := 0
			for _, change := range msg.RenamedFiles {
				if change.Action == "D" {
					deletions++
				}
			}
			if deletions > 0 {
				actionUpper = "D"
			} else {
				actionUpper = "MISC"
			}
		}
	case "SECURITY":
		actionUpper = "SECURITY"
	case "STYLE":
		actionUpper = "STYLE"
	case "TEST":
		actionUpper = "TEST"
	case "DOCS":
		actionUpper = "DOC"
	}

	actionTemplates, ok := t.templates[actionUpper]
	if !ok {
		// Fallback to MISC for unknown actions
		actionTemplates, ok = t.templates["MISC"]
		if !ok {
			return "", fmt.Errorf("no templates found for action: %s or MISC", msg.Action)
		}
	}

	topicTemplates, ok := actionTemplates[msg.Topic]
	if !ok || len(topicTemplates) == 0 {
		topicTemplates, ok = actionTemplates["_default"]
		if !ok || len(topicTemplates) == 0 {
			return "", fmt.Errorf("no templates found for topic: %s and no _default", msg.Topic)
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

	// Score and select the best template
	type scoredTemplate struct {
		template string
		score    float64
	}

	var scoredTemplates []scoredTemplate
	for _, tmpl := range topicTemplates {
		score := t.scoreTemplate(tmpl, msg)
		scoredTemplates = append(scoredTemplates, scoredTemplate{template: tmpl, score: score})
	}

	// Sort templates by score (highest first)
	for i := 0; i < len(scoredTemplates); i++ {
		for j := i + 1; j < len(scoredTemplates); j++ {
			if scoredTemplates[j].score > scoredTemplates[i].score {
				scoredTemplates[i], scoredTemplates[j] = scoredTemplates[j], scoredTemplates[i]
			}
		}
	}

	// Select the best template that's not in recent history
	var selectedTemplate string
	for _, st := range scoredTemplates {
		potentialMessage := replacer.Replace(st.template)
		if !t.history.Contains(potentialMessage) {
			selectedTemplate = st.template
			break
		}
	}

	if selectedTemplate == "" {
		// If all templates are recent duplicates, pick the highest scored one
		selectedTemplate = scoredTemplates[0].template
	}

	formattedMsg := replacer.Replace(selectedTemplate)

	// Handle scope
	if msg.Scope != "" {
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

	return formattedMsg, nil
}

// scoreTemplate scores a template based on how well it matches the commit message context
func (t *Templater) scoreTemplate(template string, msg *analyzer.CommitMessage) float64 {
	score := 0.0

	// Base score
	score += 1.0

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

	return score
}
