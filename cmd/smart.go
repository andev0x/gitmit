package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/andev0x/gitmit/internal/analyzer"
	"github.com/andev0x/gitmit/internal/generator"
)

var smartCmd = &cobra.Command{
	Use:   "smart",
	Short: "Smart commit with intelligent suggestions",
	Long: `Smart commit analyzes your changes and provides intelligent suggestions:
• Auto-detects commit type based on changes
• Suggests appropriate scopes
• Identifies breaking changes
• Provides context-aware descriptions`,
	RunE: runSmart,
}

func init() {
	rootCmd.AddCommand(smartCmd)
}

func runSmart(cmd *cobra.Command, args []string) error {
	color.Cyan("🧠 Smart Commit Analysis")
	fmt.Println()

	// Initialize components
	gitAnalyzer := analyzer.New()
	msgGenerator := generator.New("")

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
	changeAnalysis, err := gitAnalyzer.AnalyzeChanges(stagedChanges)
	if err != nil {
		color.Red("❌ Failed to analyze changes: %v", err)
		return err
	}

	// Display smart analysis
	displaySmartAnalysis(changeAnalysis)

	// Generate smart suggestions using templates
	suggestions := generateSmartSuggestions(changeAnalysis, msgGenerator)
	displaySmartSuggestions(suggestions)

	return nil
}

type SmartSuggestion struct {
	Type        string
	Scope       string
	Description string
	Confidence  int
	Reasoning   string
}

func generateSmartSuggestions(analysis *analyzer.ChangeAnalysis, msgGenerator *generator.MessageGenerator) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Analyze based on file operations
	if len(analysis.Added) > 0 && len(analysis.Modified) == 0 && len(analysis.Deleted) == 0 {
		suggestions = append(suggestions, SmartSuggestion{
			Type:        "feat",
			Scope:       getPrimaryScope(analysis.Scopes),
			Description: fmt.Sprintf("add %s", getFileDescription(analysis.Added)),
			Confidence:  90,
			Reasoning:   "Pure file additions typically indicate new features",
		})
	}

	// Analyze based on file types
	if analysis.FileTypes["md"] > 0 || analysis.FileTypes["txt"] > 0 {
		suggestions = append(suggestions, SmartSuggestion{
			Type:        "docs",
			Scope:       "docs",
			Description: "update documentation",
			Confidence:  95,
			Reasoning:   "Documentation files detected",
		})
	}

	if analysis.FileTypes["test"] > 0 || analysis.FileTypes["spec"] > 0 {
		suggestions = append(suggestions, SmartSuggestion{
			Type:        "test",
			Scope:       "test",
			Description: "add or update tests",
			Confidence:  90,
			Reasoning:   "Test files detected",
		})
	}

	// Analyze based on context hints
	for _, hint := range analysis.DiffHints {
		switch {
		case strings.Contains(hint, "fix") || strings.Contains(hint, "bug"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "fix",
				Scope:       getPrimaryScope(analysis.Scopes),
				Description: "fix bug or issue",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "performance") || strings.Contains(hint, "optimize"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "perf",
				Scope:       getPrimaryScope(analysis.Scopes),
				Description: "improve performance",
				Confidence:  80,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "security"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "security",
				Scope:       "security",
				Description: "improve security",
				Confidence:  90,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "config") || strings.Contains(hint, "settings"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "config",
				Scope:       "config",
				Description: "update configuration",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "deploy") || strings.Contains(hint, "docker"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "deploy",
				Scope:       "deploy",
				Description: "update deployment",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "revert") || strings.Contains(hint, "rollback"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "revert",
				Scope:       getPrimaryScope(analysis.Scopes),
				Description: "revert changes",
				Confidence:  90,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		case strings.Contains(hint, "wip") || strings.Contains(hint, "work in progress"):
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "wip",
				Scope:       getPrimaryScope(analysis.Scopes),
				Description: "work in progress",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("Context hint: %s", hint),
			})
		}
	}

	// Analyze based on scopes
	for _, scope := range analysis.Scopes {
		switch scope {
		case "ci", ".github":
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "ci",
				Scope:       scope,
				Description: "update CI/CD configuration",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("CI/CD scope detected: %s", scope),
			})
		case "build", "webpack", "vite":
			suggestions = append(suggestions, SmartSuggestion{
				Type:        "build",
				Scope:       scope,
				Description: "update build configuration",
				Confidence:  85,
				Reasoning:   fmt.Sprintf("Build scope detected: %s", scope),
			})
		}
	}

	// Default suggestion if no specific patterns detected
	if len(suggestions) == 0 {
		suggestions = append(suggestions, SmartSuggestion{
			Type:        "feat",
			Scope:       getPrimaryScope(analysis.Scopes),
			Description: "update code",
			Confidence:  60,
			Reasoning:   "Default suggestion for general code changes",
		})
	}

	return suggestions
}

func displaySmartSuggestions(suggestions []SmartSuggestion) {
	color.Green("💡 Smart Commit Suggestions:")
	fmt.Println()

	for i, suggestion := range suggestions {
		color.Cyan("Suggestion %d:", i+1)
		color.White("   Type: %s", suggestion.Type)
		if suggestion.Scope != "" {
			color.White("   Scope: %s", suggestion.Scope)
		}
		color.White("   Description: %s", suggestion.Description)
		color.White("   Confidence: %d%%", suggestion.Confidence)
		color.White("   Reasoning: %s", suggestion.Reasoning)
		fmt.Println()
	}

	// Show the best suggestion
	if len(suggestions) > 0 {
		best := suggestions[0]
		color.Green("🎯 Recommended Commit:")
		message := best.Type
		if best.Scope != "" {
			message += fmt.Sprintf("(%s)", best.Scope)
		}
		message += fmt.Sprintf(": %s", best.Description)
		color.White("   %s", message)
		fmt.Println()
	}
}

func getPrimaryScope(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	return scopes[0]
}

func getFileDescription(files []string) string {
	if len(files) == 1 {
		return files[0]
	}
	return fmt.Sprintf("%d files", len(files))
}

func displaySmartAnalysis(analysis *analyzer.ChangeAnalysis) {
	color.Green("📊 Smart Analysis Results:")
	fmt.Println()

	// File operations summary
	color.Cyan("📁 File Operations:")
	if len(analysis.Added) > 0 {
		color.Green("   ➕ Added: %d files", len(analysis.Added))
		for _, file := range analysis.Added {
			color.White("     - %s", file)
		}
	}
	if len(analysis.Modified) > 0 {
		color.Yellow("   📝 Modified: %d files", len(analysis.Modified))
		for _, file := range analysis.Modified {
			color.White("     - %s", file)
		}
	}
	if len(analysis.Deleted) > 0 {
		color.Red("   🗑️  Deleted: %d files", len(analysis.Deleted))
		for _, file := range analysis.Deleted {
			color.White("     - %s", file)
		}
	}
	if len(analysis.Renamed) > 0 {
		color.Blue("   🔄 Renamed: %d files", len(analysis.Renamed))
		for _, file := range analysis.Renamed {
			color.White("     - %s", file)
		}
	}
	fmt.Println()

	// File types analysis
	color.Cyan("🔍 File Type Analysis:")
	for fileType, count := range analysis.FileTypes {
		color.White("   %s: %d files", fileType, count)
	}
	fmt.Println()

	// Scope analysis
	if len(analysis.Scopes) > 0 {
		color.Cyan("🎯 Detected Scopes:")
		for _, scope := range analysis.Scopes {
			color.White("   - %s", scope)
		}
		fmt.Println()
	}

	// Context hints
	if len(analysis.DiffHints) > 0 {
		color.Cyan("💡 Context Hints:")
		for _, hint := range analysis.DiffHints {
			color.White("   - %s", hint)
		}
		fmt.Println()
	}
}
