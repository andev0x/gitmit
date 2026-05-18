package parser

import (
	"bufio"
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
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe for git status: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting git status: %w", err)
	}

	var changes []*Change
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}

		// Porcelain format: XY filename
		stagedStatus := line[0:1]
		filename := strings.TrimSpace(line[3:])

		// Skip if not staged
		if stagedStatus == " " || stagedStatus == "?" {
			continue
		}

		action := stagedStatus
		change := &Change{
			File:          filename,
			Action:        action,
			FileExtension: getFileExtension(filename),
		}

		// Handle renames and copies
		if action == "R" || action == "C" {
			parts := strings.Split(filename, " -> ")
			if len(parts) == 2 {
				change.IsRename = action == "R"
				change.IsCopy = action == "C"
				change.Source = strings.TrimSpace(parts[0])
				change.Target = strings.TrimSpace(parts[1])
				change.File = change.Target
				change.FileExtension = getFileExtension(change.Target)
			}
		}

		// Get the diff for the file using streaming
		diffCmd := exec.Command("git", "diff", "--cached", "-U0", "--", change.File)
		diffStdout, err := diffCmd.StdoutPipe()
		if err == nil {
			if err := diffCmd.Start(); err == nil {
				diffScanner := bufio.NewScanner(diffStdout)
				var diffBuilder strings.Builder
				for diffScanner.Scan() {
					diffLine := diffScanner.Text()
					if strings.HasPrefix(diffLine, "+") && !strings.HasPrefix(diffLine, "+++") {
						change.Added++
					} else if strings.HasPrefix(diffLine, "-") && !strings.HasPrefix(diffLine, "---") {
						change.Removed++
					}
					diffBuilder.WriteString(diffLine)
					diffBuilder.WriteString("\n")
				}
				change.Diff = diffBuilder.String()
				diffCmd.Wait()
			}
		}

		p.TotalAdded += change.Added
		p.TotalRemoved += change.Removed

		if (change.Added + change.Removed) >= 500 {
			change.IsMajor = true
		}

		changes = append(changes, change)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error waiting for git status: %w", err)
	}

	return changes, nil
}

// GetCurrentBranch returns the name of the current git branch
func (p *GitParser) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdout pipe for rev-parse: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting rev-parse: %w", err)
	}

	var branch string
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		branch = strings.TrimSpace(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("error waiting for rev-parse: %w", err)
	}

	return branch, nil
}

// getFileExtension returns the file extension of a given file path
func getFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}
