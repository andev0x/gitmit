package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gitmit/internal/analyzer"
	"gitmit/internal/config"
	"gitmit/internal/formatter"
	"gitmit/internal/history"
	"gitmit/internal/parser"
	"gitmit/internal/templater"
)

var (
	stagedFlag  bool
	summaryFlag bool
	autoFlag    bool
	dryRunFlag  bool
	debugFlag   bool

	proposeCmd = &cobra.Command{
		Use:   "propose",
		Short: "Propose a commit message from a git diff",
		RunE:  runPropose,
	}
)

func init() {
	rootCmd.AddCommand(proposeCmd)

	proposeCmd.Flags().BoolVar(&stagedFlag, "staged", true, "Only parse staged files (default: true)")
	proposeCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Print short output (summary only)")
	proposeCmd.Flags().BoolVar(&autoFlag, "auto", false, "Commit with the generated message")
	proposeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without committing")
	proposeCmd.Flags().BoolVar(&debugFlag, "debug", false, "Print debug info (analyzer output + chosen templates)")
}

func runPropose(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	history, err := history.LoadHistory()
	if err != nil {
		return err
	}

	gitParser := parser.NewGitParser()
	changes, err := gitParser.ParseStagedChanges()
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return fmt.Errorf("no staged changes")
	}

	analyzer := analyzer.NewAnalyzer(changes, cfg)
	commitMessage := analyzer.AnalyzeChanges(gitParser.TotalAdded, gitParser.TotalRemoved)
	if commitMessage == nil {
		return fmt.Errorf("could not analyze changes")
	}

	templater, err := templater.NewTemplater("templates.json", history)
	if err != nil {
		return err
	}

	if debugFlag {
		// Print analyzer output
		fmt.Printf("Analyzer result: %+v\n", commitMessage)
		// Print available templates/action/topic info from templater
		if act, tpls := templater.DebugInfo(commitMessage); tpls != nil {
			fmt.Printf("Resolved action key: %s\n", act)
			fmt.Printf("Candidate templates (first 10):\n")
			for i, t := range tpls {
				if i >= 10 {
					break
				}
				fmt.Printf("  - %s\n", t)
			}
		}
	}

	initialMessage, err := templater.GetMessage(commitMessage)
	if err != nil {
		return err
	}

	formatter := formatter.NewFormatter()
	finalMessage := formatter.FormatMessage(initialMessage, commitMessage.IsMajor)

	if summaryFlag {
		fmt.Println(finalMessage)
	} else {
		color.Green(finalMessage)
		fmt.Println("\nCopy the message above and use it to commit.")
	}

	if autoFlag && !dryRunFlag {
		commitCmd := exec.Command("git", "commit", "-m", finalMessage)
		commitCmd.Stdout = os.Stdout
		commitCmd.Stderr = os.Stderr
		err := commitCmd.Run()
		if err != nil {
			return fmt.Errorf("error committing changes: %w", err)
		}
		fmt.Println("Changes committed successfully.")
		history.AddEntry(finalMessage, initialMessage) // Pass actual template used
		if err := history.SaveHistory(); err != nil {
			return err
		}
	} else if dryRunFlag {
		fmt.Println("\n(Dry run: no changes committed)")
	}

	return nil
}
