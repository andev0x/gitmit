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
	actionUpper := strings.ToUpper(msg.Action)
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
		selectedTemplate = topicTemplates[rand.Intn(len(topicTemplates))]
	}

	formattedMsg := replacer.Replace(selectedTemplate)

	// Handle scope
	if msg.Scope != "" {
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

	return formattedMsg, nil
}
