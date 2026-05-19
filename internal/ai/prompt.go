package ai

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/andev0x/gitmit/assets"
	"github.com/andev0x/gitmit/internal/analyzer"
)

// PromptContext represents the data structure passed to the prompt template
type PromptContext struct {
	ProjectType     string
	CurrentBranch   string
	RecommendedType string
	Files           []string
	CodeSymbols     []string
	DependencyAlert string
	DiffSummary     DiffSummary
}

// DiffSummary contains ratio of changes
type DiffSummary struct {
	Ratio float64
}

// RenderPrompt generates the prompt string using the provided context
func RenderPrompt(msg *analyzer.CommitMessage, projectType, branchName string) (string, error) {
	promptTemplate, err := assets.GetPrompt()
	if err != nil {
		return "", fmt.Errorf("error loading prompt template: %w", err)
	}

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing prompt template: %w", err)
	}

	var codeSymbols []string
	for _, f := range msg.DetectedFunctions {
		codeSymbols = append(codeSymbols, fmt.Sprintf("[func] %s", f))
	}
	for _, s := range msg.DetectedStructs {
		codeSymbols = append(codeSymbols, fmt.Sprintf("[struct] %s", s))
	}
	for _, m := range msg.DetectedMethods {
		codeSymbols = append(codeSymbols, fmt.Sprintf("[method] %s", m))
	}

	depAlert := "None"
	if msg.IsDepsOnly {
		depAlert = "Dependency changes detected in package manager files"
	}

	ratio := 0.0
	total := msg.TotalAdded + msg.TotalRemoved
	if total > 0 {
		ratio = float64(msg.TotalAdded) / float64(total)
	}

	ctx := PromptContext{
		ProjectType:     projectType,
		CurrentBranch:   branchName,
		RecommendedType: msg.Action,
		Files:           msg.Files,
		CodeSymbols:     codeSymbols,
		DependencyAlert: depAlert,
		DiffSummary: DiffSummary{
			Ratio: ratio,
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("error executing prompt template: %w", err)
	}

	return buf.String(), nil
}

// IsValidCommitMessage checks if the AI output follows the Conventional Commits format
func IsValidCommitMessage(msg string) bool {
	// Simple regex check for <type>(<scope>): <description> or <type>: <description>
	// Conventional commits regex: ^([a-z]+)(\([a-z0-9/,-]+\))?!?: .+$
	// We'll use a slightly more relaxed one as requested in the blueprint
	
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return false
	}

	// Basic check for type and colon
	types := []string{"feat", "fix", "refactor", "chore", "test", "docs", "style", "perf", "ci", "build", "security"}
	
	hasType := false
	for _, t := range types {
		if strings.HasPrefix(msg, t) {
			hasType = true
			break
		}
	}
	
	if !hasType {
		return false
	}

	if !strings.Contains(msg, ": ") {
		return false
	}

	return true
}
