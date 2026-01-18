package history

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const historyFileName = ".commit_suggest_history.json"
const maxHistoryEntries = 10

// HistoryEntry represents a single entry in the commit history
type HistoryEntry struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Template  string    `json:"template,omitempty"` // Optional: store which template was used
}

// CommitHistory represents the list of past commit suggestions
type CommitHistory struct {
	Entries []HistoryEntry `json:"entries"`
}

// LoadHistory loads the commit history from .commit_suggest_history.json
func LoadHistory() (*CommitHistory, error) {
	data, err := os.ReadFile(historyFileName)
	if os.IsNotExist(err) {
		return &CommitHistory{Entries: []HistoryEntry{}}, nil // Return empty history if file doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("error reading commit history file %s: %w", historyFileName, err)
	}

	var history CommitHistory
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling commit history file %s: %w", historyFileName, err)
	}

	return &history, nil
}

// SaveHistory saves the commit history to .commit_suggest_history.json
func (h *CommitHistory) SaveHistory() error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling commit history: %w", err)
	}

	err = os.WriteFile(historyFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing commit history file %s: %w", historyFileName, err)
	}

	return nil
}

// AddEntry adds a new entry to the commit history, keeping only the latest maxHistoryEntries
func (h *CommitHistory) AddEntry(message, template string) {
	newEntry := HistoryEntry{
		Message:   message,
		Timestamp: time.Now(),
		Template:  template,
	}

	h.Entries = append([]HistoryEntry{newEntry}, h.Entries...)

	// Keep only the latest N entries
	if len(h.Entries) > maxHistoryEntries {
		h.Entries = h.Entries[:maxHistoryEntries]
	}
}

// Contains checks if the history contains a given message
func (h *CommitHistory) Contains(message string) bool {
	for _, entry := range h.Entries {
		if entry.Message == message {
			return true
		}
	}
	return false
}

// GetRecentCommitContext retrieves the most recent commit message from git history
// This helps maintain consistency by suggesting similar topics/scopes
func GetRecentCommitContext() (string, string, error) {
	// Get the last commit message on the current branch
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("error getting recent commit: %w", err)
	}

	commitMsg := strings.TrimSpace(out.String())
	if commitMsg == "" {
		return "", "", nil
	}

	// Extract topic/scope from conventional commit format: type(scope): message
	// Pattern: type(scope): message
	re := regexp.MustCompile(`^[a-z]+\(([^)]+)\):`)
	matches := re.FindStringSubmatch(commitMsg)
	if len(matches) > 1 {
		scope := matches[1]
		return commitMsg, scope, nil
	}

	return commitMsg, "", nil
}

// GetRecentCommits retrieves the last N commit messages from git history
func GetRecentCommits(count int) ([]string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--pretty=%B")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error getting recent commits: %w", err)
	}

	commits := []string{}
	lines := strings.Split(out.String(), "\n")
	currentCommit := ""
	for _, line := range lines {
		if line == "" {
			if currentCommit != "" {
				commits = append(commits, strings.TrimSpace(currentCommit))
				currentCommit = ""
			}
		} else {
			currentCommit += line + " "
		}
	}
	if currentCommit != "" {
		commits = append(commits, strings.TrimSpace(currentCommit))
	}

	return commits, nil
}
