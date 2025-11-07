package templater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"gitmit/internal/analyzer"
)

// Templates holds the loaded commit message templates
type Templates map[string]map[string][]string

// Templater is responsible for selecting and formatting commit messages
type Templater struct {
	templates Templates
}

// NewTemplater creates a new Templater
func NewTemplater(templateFile string) (*Templater, error) {
	data, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("error reading template file: %w", err)
	}

	var templates Templates
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling template file: %w", err)
	}

	rand.Seed(time.Now().UnixNano())

	return &Templater{templates: templates}, nil
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

	// Select a random template
	template := topicTemplates[rand.Intn(len(topicTemplates))]

	// Replace placeholders
	replacer := strings.NewReplacer(
		"{topic}", msg.Topic,
		"{item}", msg.Item,
		"{purpose}", msg.Purpose,
		"{source}", "", // Placeholder, to be implemented
		"{target}", "", // Placeholder, to be implemented
	)
	formattedMsg := replacer.Replace(template)

	// Handle scope
	if msg.Scope != "" {
		formattedMsg = strings.Replace(formattedMsg, "("+msg.Topic+")", "("+msg.Scope+")", 1)
	}

	return formattedMsg, nil
}
