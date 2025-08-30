package cmd

import (
	"fmt"
	"strings"

	"github.com/andev0x/gitmit/internal/llm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/andev0x/gitmit/internal/analyzer"
)

var smartCmd = &cobra.Command{
	Use:	"smart",
	Short:	"Smart commit with intelligent suggestions",
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

	// Get staged diff
	diff, err := gitAnalyzer.GetStagedDiff()
	if err != nil {
		color.Red("❌ Failed to get staged diff: %v", err)
		return err
	}

	// Get recent commits
	recentCommits, err := gitAnalyzer.GetRecentCommits(5)
	if err != nil {
		color.Red("❌ Failed to get recent commits: %v", err)
		return err
	}

	// Format analysis
	analysisString := formatAnalysis(changeAnalysis)

	// Propose commit message
	color.Cyan("🤖 Proposing commit message...")
	proposedMessage, err := llm.ProposeCommitWithAnalysis(cmd.Context(), "", diff, analysisString, recentCommits)
	if err != nil {
		color.Red("❌ Failed to propose commit message: %v", err)
		return err
	}

	color.Green("💡 Proposed Commit Message:")
	color.White(proposedMessage)

	return nil
}

func formatAnalysis(analysis *analyzer.ChangeAnalysis) string {
	var builder strings.Builder

	builder.WriteString("File Operations:\n")
	if len(analysis.Added) > 0 {
		builder.WriteString(fmt.Sprintf("  Added: %s\n", strings.Join(analysis.Added, ", ")))
	}
	if len(analysis.Modified) > 0 {
		builder.WriteString(fmt.Sprintf("  Modified: %s\n", strings.Join(analysis.Modified, ", ")))
	}
	if len(analysis.Deleted) > 0 {
		builder.WriteString(fmt.Sprintf("  Deleted: %s\n", strings.Join(analysis.Deleted, ", ")))
	}
	if len(analysis.Renamed) > 0 {
		builder.WriteString(fmt.Sprintf("  Renamed: %s\n", strings.Join(analysis.Renamed, ", ")))
	}

	builder.WriteString("\nFile Types:\n")
	for fileType, count := range analysis.FileTypes {
		builder.WriteString(fmt.Sprintf("  %s: %d\n", fileType, count))
	}

	if len(analysis.Scopes) > 0 {
		builder.WriteString("\nDetected Scopes:\n")
		builder.WriteString(fmt.Sprintf("  %s\n", strings.Join(analysis.Scopes, ", ")))
	}

	if len(analysis.DiffHints) > 0 {
		builder.WriteString("\nContext Hints:\n")
		for _, hint := range analysis.DiffHints {
			builder.WriteString(fmt.Sprintf("  - %s\n", hint))
		}
	}

	return builder.String()
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
