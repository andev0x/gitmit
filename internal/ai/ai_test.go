package ai

import (
	"strings"
	"testing"

	"github.com/andev0x/gitmit/internal/analyzer"
)

func TestRenderPrompt(t *testing.T) {
	msg := &analyzer.CommitMessage{
		Action: "feat",
		Topic:  "auth",
		Files:  []string{"internal/auth/login.go", "internal/auth/logout.go"},
		DetectedFunctions: []string{"Login", "Logout"},
		TotalAdded: 50,
		TotalRemoved: 10,
	}

	prompt, err := RenderPrompt(msg, "go", "feature/auth-implementation")
	if err != nil {
		t.Fatalf("RenderPrompt failed: %v", err)
	}

	expectedParts := []string{
		"Project Type: go",
		"Active Branch Name: feature/auth-implementation",
		"Detected Intent/Type Bonus: feat",
		"internal/auth/login.go",
		"[func] Login",
		"Added/Deleted Line Ratio: 0.83",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("Prompt missing expected part: %s", part)
		}
	}
}

func TestIsValidCommitMessage(t *testing.T) {
	tests := []struct {
		msg      string
		expected bool
	}{
		{"feat(auth): add login functionality", true},
		{"fix: resolve memory leak", true},
		{"chore(deps): update dependencies", true},
		{"Invalid message", false},
		{"feat add something", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsValidCommitMessage(tt.msg); got != tt.expected {
			t.Errorf("IsValidCommitMessage(%q) = %v; want %v", tt.msg, got, tt.expected)
		}
	}
}
