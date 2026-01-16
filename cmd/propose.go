package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	stagedFlag     bool
	summaryFlag    bool
	autoFlag       bool
	dryRunFlag     bool
	debugFlag      bool
	contextFlag    bool
	maxSuggestions int

	proposeCmd = &cobra.Command{
		Use:   "propose",
		Short: "Propose commit messages from git diff",
		Long: `Analyze staged changes and suggest commit messages based on the context.

When using --interactive (-i) or --suggestions (-s), multiple suggestions will be shown
ranked by how well they match the context (file types, changes, purposes).

The --context flag shows what was analyzed to help understand the suggestions.`,
		Example: `  gitmit propose              # Get best suggestion
  gitmit propose -i          # Choose from multiple suggestions
  gitmit propose -s          # Show ranked suggestions
  gitmit propose --context   # Show what was analyzed
  gitmit propose --auto      # Auto-commit with best suggestion`,
		RunE: runPropose,
	}
)

func init() {
	rootCmd.AddCommand(proposeCmd)

	proposeCmd.Flags().BoolVar(&stagedFlag, "staged", true, "Only parse staged files (default: true)")
	proposeCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Print short output (summary only)")
	proposeCmd.Flags().BoolVar(&autoFlag, "auto", false, "Auto-commit with the generated message")
	proposeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without committing")
	proposeCmd.Flags().BoolVar(&debugFlag, "debug", false, "Print debug info (analyzer output + chosen templates)")
	proposeCmd.Flags().BoolVar(&contextFlag, "context", false, "Show what was analyzed to generate suggestions")
	proposeCmd.Flags().IntVar(&maxSuggestions, "max-suggestions", 5, "Maximum number of suggestions to show")
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
		return fmt.Errorf("⚠️ no staged changes")
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

	// Show analysis context if requested
	if contextFlag || debugFlag {
		color.Blue("\n📊 Analysis Context:")
		fmt.Printf("Action: %s\n", commitMessage.Action)
		fmt.Printf("Topic:  %s\n", commitMessage.Topic)
		if commitMessage.Item != "" {
			fmt.Printf("Item:   %s\n", commitMessage.Item)
		}
		if commitMessage.Purpose != "" {
			fmt.Printf("Purpose: %s\n", commitMessage.Purpose)
		}
		if commitMessage.Scope != "" {
			fmt.Printf("Scope:  %s\n", commitMessage.Scope)
		}
		fmt.Printf("Files:  +%d -%d\n", commitMessage.TotalAdded, commitMessage.TotalRemoved)
		if len(commitMessage.FileExtensions) > 0 {
			fmt.Printf("Types:  %v\n", commitMessage.FileExtensions)
		}
		fmt.Println()
	}

	if debugFlag {
		// Print more detailed debug info
		fmt.Printf("Full analyzer output: %+v\n", commitMessage)
		if act, tpls := templater.DebugInfo(commitMessage); tpls != nil {
			fmt.Printf("Template group: %s\n", act)
			fmt.Printf("Candidate templates:\n")
			for i, t := range tpls {
				if i >= 10 {
					break
				}
				fmt.Printf("  - %s\n", t)
			}
		}
	}

	// Get multiple suggestions if interactive/suggestions mode
	var suggestions []string
	if interactiveFlag || suggestionsFlag {
		suggestions, err = templater.GetSuggestions(commitMessage, maxSuggestions)
		if err != nil {
			return err
		}
	} else {
		// Just get best message
		msg, err := templater.GetMessage(commitMessage)
		if err != nil {
			return err
		}
		suggestions = []string{msg}
	}

	formatter := formatter.NewFormatter()

	if len(suggestions) == 0 {
		return fmt.Errorf("no suitable commit messages found")
	}

	// Format all suggestions
	formattedSuggestions := make([]string, len(suggestions))
	for i, msg := range suggestions {
		formattedSuggestions[i] = formatter.FormatMessage(msg, commitMessage.IsMajor)
	}

	// Default to first/best suggestion
	finalMessage := formattedSuggestions[0]

	if suggestionsFlag {
		// Show all suggestions with ranking
		color.Blue("\n💡 Ranked Suggestions:")
		for i, msg := range formattedSuggestions {
			if i == 0 {
				color.Green("1. %s (recommended)\n", msg)
			} else {
				fmt.Printf("%d. %s\n", i+1, msg)
			}
		}
		fmt.Println()
	}

	if interactiveFlag && len(formattedSuggestions) > 1 {
		// TODO: Add interactive selection using a proper terminal UI library
		// For now, just show numbered options and read input
		color.Blue("\n📝 Choose a commit message:")
		for i, msg := range formattedSuggestions {
			fmt.Printf("%d. %s\n", i+1, msg)
		}
		fmt.Printf("\nEnter number (1-%d) [1]: ", len(formattedSuggestions))

		var choice string
		fmt.Scanln(&choice)

		if choice != "" {
			var num int
			if _, err := fmt.Sscanf(choice, "%d", &num); err == nil && num > 0 && num <= len(formattedSuggestions) {
				finalMessage = formattedSuggestions[num-1]
			}
		}
		fmt.Println()

	}

	// If not in summary mode, show the suggestion and prompt for action
	if !summaryFlag {
		color.Green("\n💡 Suggested commit message:")
		fmt.Printf("%s\n\n", finalMessage)

		if !autoFlag && !dryRunFlag {
			// Track used suggestions to avoid repetition with 'r' option
			usedSuggestions := map[string]bool{finalMessage: true}
			regenerationCount := 0
			const maxRegenerations = 10

			for {
				color.Blue("Actions:")
				fmt.Println("  y - Accept and commit")
				fmt.Println("  n - Reject and exit")
				fmt.Println("  e - Edit message manually")
				fmt.Println("  r - Regenerate different suggestion")
				fmt.Printf("\nChoice [y/n/e/r]: ")

				reader := bufio.NewReader(os.Stdin)
				choice, _ := reader.ReadString('\n')
				choice = strings.TrimSpace(strings.ToLower(choice))
				fmt.Println()

				switch choice {
				case "y", "":
					// Commit the message
					commitCmd := exec.Command("git", "commit", "-m", finalMessage)
					commitCmd.Stdout = os.Stdout
					commitCmd.Stderr = os.Stderr
					err := commitCmd.Run()
					if err != nil {
						return fmt.Errorf("error committing changes: %w", err)
					}
					color.Green("✅ Changes committed successfully.")
					history.AddEntry(finalMessage, "") // Save to history
					if err := history.SaveHistory(); err != nil {
						return err
					}
					return nil

				case "n":
					color.Yellow("❌ Commit cancelled.")
					return nil

				case "e":
					color.Blue("📝 Edit the commit message:")
					fmt.Printf("Current: %s\n", finalMessage)
					fmt.Print("New message: ")

					editedMessage, _ := reader.ReadString('\n')
					editedMessage = strings.TrimSpace(editedMessage)

					if editedMessage != "" {
						finalMessage = editedMessage
						usedSuggestions[finalMessage] = true
						// Show the edited message and prompt again
						color.Green("\n✓ Updated commit message:")
						fmt.Printf("%s\n\n", finalMessage)
						continue
					} else {
						color.Yellow("⚠ No changes made. Keeping current message.\n\n")
						continue
					}

				case "r":
					if regenerationCount >= maxRegenerations {
						color.Yellow("⚠ Maximum regeneration attempts reached.\n\n")
						continue
					}

					// Generate a new alternative suggestion
					newSuggestion, err := templater.GetAlternativeSuggestion(commitMessage, usedSuggestions)
					if err != nil || newSuggestion == "" {
						color.Yellow("⚠ Could not generate alternative suggestion. Try editing instead.\n\n")
						continue
					}

					regenerationCount++
					finalMessage = formatter.FormatMessage(newSuggestion, commitMessage.IsMajor)
					usedSuggestions[finalMessage] = true

					color.Green("\n💡 Alternative suggestion #%d:", regenerationCount)
					fmt.Printf("%s\n\n", finalMessage)
					continue

				default:
					color.Yellow("⚠ Invalid choice. Please select y, n, e, or r.\n\n")
					continue
				}
			}
		}
	} else {
		fmt.Println(finalMessage)
	}

	// Handle auto-commit and dry-run cases
	if autoFlag && !dryRunFlag {
		commitCmd := exec.Command("git", "commit", "-m", finalMessage)
		commitCmd.Stdout = os.Stdout
		commitCmd.Stderr = os.Stderr
		err := commitCmd.Run()
		if err != nil {
			return fmt.Errorf("error committing changes: %w", err)
		}
		color.Green("✅ Changes committed successfully.")
		history.AddEntry(finalMessage, "") // Save to history
		if err := history.SaveHistory(); err != nil {
			return err
		}
	} else if dryRunFlag {
		fmt.Println("\n(Dry run: no changes committed)")
	}

	return nil
}
