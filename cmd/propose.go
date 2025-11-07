package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gitmit/internal/analyzer"
	"gitmit/internal/formatter"
	"gitmit/internal/parser"
	"gitmit/internal/templater"
)

var proposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Propose a commit message from a git diff",
	RunE:  runPropose,
}

func init() {
	rootCmd.AddCommand(proposeCmd)
}

func runPropose(cmd *cobra.Command, args []string) error {
	gitParser := parser.NewGitParser()
	changes, err := gitParser.ParseStagedChanges()
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return fmt.Errorf("no staged changes")
	}

	analyzer := analyzer.NewAnalyzer(changes)
	commitMessage := analyzer.AnalyzeChanges()
	if commitMessage == nil {
		return fmt.Errorf("could not analyze changes")
	}

	templater, err := templater.NewTemplater("templates.json")
	if err != nil {
		return err
	}

	initialMessage, err := templater.GetMessage(commitMessage)
	if err != nil {
		return err
	}

	formatter := formatter.NewFormatter()
	finalMessage := formatter.FormatMessage(initialMessage)

	color.Green(finalMessage)

	fmt.Println("\nCopy the message above and use it to commit.")

	return nil
}
