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

// ParseStagedChanges parses the staged changes from git using git status --porcelain
func (p *GitParser) ParseStagedChanges() ([]*Change, error) {
	// Use git status --porcelain for more accurate file state detection
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running git status --porcelain: %w", err)
	}

	var changes []*Change
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}

		// Porcelain format: XY filename
		// X = staged status, Y = unstaged status
		stagedStatus := line[0:1]
		// unstagedStatus := line[1:2]
		filename := strings.TrimSpace(line[3:])

		// Skip if not staged (staged status is space)
		if stagedStatus == " " || stagedStatus == "?" {
			continue
		}

		// Map porcelain status to action
		action := stagedStatus
		switch stagedStatus {
		case "M":
			action = "M" // Modified
		case "A":
			action = "A" // Added
		case "D":
			action = "D" // Deleted
		case "R":
			action = "R" // Renamed
		case "C":
			action = "C" // Copied
		}

		change := &Change{
			File:          filename,
			Action:        action,
			FileExtension: getFileExtension(filename),
		}

		// Handle renames and copies (format: "R  oldname -> newname")
		if action == "R" || action == "C" {
			parts := strings.Split(filename, " -> ")
			if len(parts) == 2 {
				change.IsRename = action == "R"
				change.IsCopy = action == "C"
				change.Source = strings.TrimSpace(parts[0])
				change.Target = strings.TrimSpace(parts[1])
				change.File = change.Target // Use the new name as the file
				change.FileExtension = getFileExtension(change.Target)
			}
		}

		// Get the diff for the file
		diffCmd := exec.Command("git", "diff", "--cached", "-U0", "--", change.File)
		var diffOut bytes.Buffer
		diffCmd.Stdout = &diffOut
		err := diffCmd.Run()
		if err != nil && action != "D" {
			// For deleted files, diff may fail, which is expected
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
		return nil, fmt.Errorf("error scanning git status output: %w", err)
	}

	return changes, nil
}

// getFileExtension returns the file extension of a given file path
func getFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}
