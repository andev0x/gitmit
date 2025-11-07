package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	data, err := ioutil.ReadFile(historyFileName)
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

	err = ioutil.WriteFile(historyFileName, data, 0644)
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
