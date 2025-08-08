package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/andev0x/gitmit/internal/analyzer"
	"github.com/andev0x/gitmit/internal/generator"
	"github.com/andev0x/gitmit/internal/prompt"
)

var (
	version = "0.0.5"

	rootCmd = &cobra.Command{
		Use:   "gitmit",
		Short: "🧠 Smart Git Commit Message Generator",
		Long: `Gitmit is a lightweight CLI tool that analyzes your staged changes
and suggests professional commit messages following Conventional Commits format.

Features:
• Intelligent analysis of git status and diff
• Conventional Commits format (feat, fix, refactor, etc.)
• Interactive mode for customization
• Zero configuration required
• Lightning-fast local performance
• Complete offline operation`,
		Version: version,
		RunE:    runGitmit,
	}

	dryRun    bool
	verbose   bool
	useOpenAI bool
)

func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show suggested message without committing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed analysis")
	rootCmd.Flags().BoolVarP(&useOpenAI, "openai", "o", false, "Use OpenAI API for commit message generation")
}

func Execute() error {
	return rootCmd.Execute()
}

func runGitmit(cmd *cobra.Command, args []string) error {
	// Print header
	cyan := color.New(color.FgCyan, color.Bold)
	if _, err := cyan.Println("🧠 Gitmit - Smart Git Commit"); err != nil {
		return err
	}
	fmt.Println()

	// Initialize components
	gitAnalyzer := analyzer.New()

	var openAIAPIKey string
	if useOpenAI {
		// Temporarily create a prompt instance to get the key
		tempPrompt := prompt.New("") // Pass empty string for now
		key, err := tempPrompt.PromptForOpenAIKey()
		if err != nil {
			return err
		}
		openAIAPIKey = key
	}

	interactivePrompt := prompt.New(openAIAPIKey)
	msgGenerator := generator.New(openAIAPIKey)

	// Check if we're in a git repository
	if !gitAnalyzer.IsGitRepository() {
		color.Red("❌ Not a git repository")
		return fmt.Errorf("current directory is not a git repository")
	}

	// Get staged changes
	stagedChanges, err := gitAnalyzer.GetStagedChanges()
	if err != nil {
		color.Red("❌ Failed to get staged changes: %v", err)
		return err
	}

	if len(stagedChanges) == 0 {
		color.Yellow("⚠️  No staged changes found. Stage some files first with 'git add'")
		return nil
	}

	// Analyze changes
	if verbose {
		color.HiBlack("Analyzing staged changes...")
		fmt.Println()
	}

	changeAnalysis, err := gitAnalyzer.AnalyzeChanges(stagedChanges)
	if err != nil {
		color.Red("❌ Failed to analyze changes: %v", err)
		return err
	}

	// Display analysis if verbose
	if verbose {
		displayAnalysis(changeAnalysis)
	}

	// Generate suggested message
	suggestedMessage := msgGenerator.GenerateMessage(changeAnalysis)

	// Show suggested message
	color.Green("\n💡 Suggested commit message:")
	color.White("   %s\n", suggestedMessage)

	// Handle dry-run mode
	if dryRun {
		color.Cyan("🔍 Dry-run mode: No commit will be made")
		return nil
	}

	// Interactive prompt
	for {
		stagedDiff, err := gitAnalyzer.GetStagedDiff()
		if err != nil {
			color.Red("❌ Failed to get staged diff: %v", err)
			return err
		}
		message, err := interactivePrompt.PromptForMessage(suggestedMessage, stagedDiff)
		if err != nil {
			return err
		}
		if message == "__regenerate__" {
			// Locally regenerate the commit message
			suggestedMessage = msgGenerator.GenerateMessage(changeAnalysis)
			continue
		}
		finalMessage := message
		if finalMessage == "" {
			color.Yellow("🚫 Commit cancelled")
			return nil
		}
		// Commit with the final message
		if err := gitAnalyzer.Commit(finalMessage); err != nil {
			color.Red("❌ Failed to commit: %v", err)
			return err
		}
		color.Green("✅ Committed: %s", finalMessage)
		return nil
	}
}

func displayAnalysis(analysis *analyzer.ChangeAnalysis) {
	color.Cyan("📊 Change Analysis:")

	if len(analysis.Added) > 0 {
		color.Green("   ➕ Added: %v", analysis.Added)
	}

	if len(analysis.Modified) > 0 {
		color.Yellow("   📝 Modified: %v", analysis.Modified)
	}

	if len(analysis.Deleted) > 0 {
		color.Red("   🗑️  Deleted: %v", analysis.Deleted)
	}

	if len(analysis.Renamed) > 0 {
		color.Blue("   🔄 Renamed: %v", analysis.Renamed)
	}

	if len(analysis.DiffHints) > 0 {
		color.HiBlack("   🔍 Context: %v", analysis.DiffHints)
	}

	if len(analysis.Scopes) > 0 {
		color.Magenta("   🎯 Scopes: %v", analysis.Scopes)
	}
}
