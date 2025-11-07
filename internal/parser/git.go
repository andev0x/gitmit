package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Change represents a single file change
type Change struct {
	File          string
	Action        string
	Added         int
	Removed       int
	IsMajor       bool
	IsRename      bool
	IsCopy        bool
	Source        string
	Target        string
	Diff          string
	FileExtension string
}

// GitParser is responsible for parsing git diffs
type GitParser struct {
	TotalAdded   int
	TotalRemoved int
}

// NewGitParser creates a new GitParser
func NewGitParser() *GitParser {
	return &GitParser{}
}

// ParseStagedChanges parses the staged changes from git
func (p *GitParser) ParseStagedChanges() ([]*Change, error) {
	// Get the list of staged files and their status
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running git diff --cached --name-status: %w", err)
	}

	var changes []*Change
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		action := string(parts[0][0])
		file := parts[1]

		change := &Change{
			File:          file,
			Action:        action,
			FileExtension: getFileExtension(file),
		}

		// Handle renames and copies
		if action == "R" || action == "C" {
			if len(parts) < 3 {
				continue
			}
			change.IsRename = action == "R"
			change.IsCopy = action == "C"
			change.Source = parts[1]
			change.Target = parts[2]
			change.File = parts[2] // Use the new name as the file
			change.FileExtension = getFileExtension(parts[2])
		}

		// Get the diff for the file
		diffCmd := exec.Command("git", "diff", "--cached", "-U0", "--", change.File)
		var diffOut bytes.Buffer
		diffCmd.Stdout = &diffOut
		err := diffCmd.Run()
		if err != nil {
			return nil, fmt.Errorf("error running git diff for %s: %w", change.File, err)
		}
		change.Diff = diffOut.String()

		// Count added and removed lines
		diffScanner := bufio.NewScanner(strings.NewReader(change.Diff))
		for diffScanner.Scan() {
			diffLine := diffScanner.Text()
			if strings.HasPrefix(diffLine, "+") && !strings.HasPrefix(diffLine, "+++") {
				change.Added++
			} else if strings.HasPrefix(diffLine, "-") && !strings.HasPrefix(diffLine, "---") {
				change.Removed++
			}
		}
		p.TotalAdded += change.Added
		p.TotalRemoved += change.Removed

		// Detect large changes
		if (change.Added + change.Removed) >= 500 {
			change.IsMajor = true
		}

		changes = append(changes, change)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning git diff output: %w", err)
	}

	return changes, nil
}

// getFileExtension returns the file extension of a given file path
func getFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}
