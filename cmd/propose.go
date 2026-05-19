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
	"gitmit/internal/ai"
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
	branchName, _ := gitParser.GetCurrentBranch()
	commitMessage := analyzer.AnalyzeChanges(gitParser.TotalAdded, gitParser.TotalRemoved, branchName)
	if commitMessage == nil {
		return fmt.Errorf("could not analyze changes")
	}

	templater, err := templater.NewTemplater("templates.json", history)
	if err != nil {
		return err
	}

	f := formatter.NewFormatter()

	// Calculate Heuristic Suggestion (Always available)
	heuristicMsg, err := templater.GetMessage(commitMessage)
	if err != nil {
		return err
	}
	formattedHeuristic := f.FormatMessage(heuristicMsg, commitMessage.IsMajor)

	var aiMsg string
	var finalMessage string
	var usingAI bool

	// AI Engine Logic
	if cfg.Engine == "ollama" {
		prompt, err := ai.RenderPrompt(commitMessage, cfg.ProjectType, branchName)
		if err == nil {
			client := ai.NewOllamaClient(cfg.Ollama)
			aiResponse, err := client.Generate(prompt)
			if err == nil && ai.IsValidCommitMessage(aiResponse) {
				aiMsg = strings.TrimSpace(aiResponse)
				usingAI = true
				finalMessage = aiMsg
			}
		}
	}

	if !usingAI {
		finalMessage = formattedHeuristic
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

	if suggestionsFlag && !usingAI {
		// Show ranked suggestions only for Heuristic
		color.Blue("\n💡 Ranked Suggestions:")
		suggestions, _ := templater.GetSuggestions(commitMessage, maxSuggestions)
		for i, msg := range suggestions {
			fmt.Printf("%d. %s\n", i+1, f.FormatMessage(msg, commitMessage.IsMajor))
		}
		fmt.Println()
	}

	// Interactive Mode logic
	if !summaryFlag && !autoFlag && !dryRunFlag {
		usedSuggestions := map[string]bool{finalMessage: true}
		regenerationCount := 0
		const maxRegenerations = 10

		for {
			fmt.Println()
			if usingAI {
				color.Cyan("Generated via: Local AI Engine [%s]", cfg.Ollama.Model)
			} else {
				color.Blue("Generated via: Heuristic Engine [Matrix Scored]")
			}

			color.Green("\n💡 Suggested commit message:")
			fmt.Printf("%s\n\n", finalMessage)

			color.Blue("Actions:")
			fmt.Println("  y - Accept and commit")
			fmt.Println("  n - Reject and exit")
			fmt.Println("  e - Edit message manually")

			if usingAI {
				fmt.Println("  r - Regenerate an alternative AI suggestion")
				fmt.Println("  h - Fallback to classic Heuristic suggestion")
			} else {
				fmt.Println("  r - Regenerate different suggestion (Heuristic)")
				fmt.Println("  a - Upgrade suggestion with Local AI (Ollama)")
			}
			fmt.Printf("\nChoice [y/n/e/r/%s]: ", map[bool]string{true: "h", false: "a"}[usingAI])

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			choice := strings.TrimSpace(strings.ToLower(input))
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
					color.Green("\n✓ Updated commit message:")
				} else {
					color.Yellow("⚠ No changes made. Keeping current message.\n")
				}
				continue

			case "r":
				if regenerationCount >= maxRegenerations {
					color.Yellow("⚠ Maximum regeneration attempts reached.\n")
					continue
				}

				if usingAI {
					prompt, err := ai.RenderPrompt(commitMessage, cfg.ProjectType, branchName)
					if err == nil {
						client := ai.NewOllamaClient(cfg.Ollama)
						aiResponse, err := client.Generate(prompt)
						if err == nil && ai.IsValidCommitMessage(aiResponse) {
							finalMessage = strings.TrimSpace(aiResponse)
							regenerationCount++
						}
					}
				} else {
					newSuggestion, err := templater.GetAlternativeSuggestion(commitMessage, usedSuggestions)
					if err == nil && newSuggestion != "" {
						finalMessage = f.FormatMessage(newSuggestion, commitMessage.IsMajor)
						regenerationCount++
					}
				}
				usedSuggestions[finalMessage] = true
				continue

			case "a":
				if usingAI {
					continue
				}
				// Try to connect to Ollama
				prompt, err := ai.RenderPrompt(commitMessage, cfg.ProjectType, branchName)
				if err == nil {
					client := ai.NewOllamaClient(cfg.Ollama)
					aiResponse, err := client.Generate(prompt)
					if err == nil && ai.IsValidCommitMessage(aiResponse) {
						aiMsg = strings.TrimSpace(aiResponse)
						finalMessage = aiMsg
						usingAI = true
					} else {
						color.Red("\n⚠️  Ollama connection not detected on %s", cfg.Ollama.URL)
						fmt.Println("To enable Local AI generation, please ensure:")
						fmt.Println("  1. Ollama is running locally (`ollama serve`)")
						fmt.Printf("  2. The required model is pulled (`ollama pull %s`)\n", cfg.Ollama.Model)
						fmt.Println("  3. Your .gitmit.json sets \"engine\": \"ollama\"")
						fmt.Println("\nFalling back to interactive options...")
					}
				}
				continue

			case "h":
				if !usingAI {
					continue
				}
				usingAI = false
				finalMessage = formattedHeuristic
				continue

			default:
				color.Yellow("⚠ Invalid choice. Please select a valid option.\n")
				continue
			}
		}
	}

	// Handle non-interactive cases (summary, auto, dry-run)
	if summaryFlag {
		fmt.Println(finalMessage)
		return nil
	}

	color.Green("\n💡 Suggested commit message:")
	fmt.Printf("%s\n\n", finalMessage)



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
