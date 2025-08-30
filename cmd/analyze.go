package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze commit history and provide insights",
	Long: `Analyze your git commit history to provide insights about:
• Commit patterns and trends
• Most active files and directories
• Commit type distribution
• Development velocity
• Potential improvements`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	color.Cyan("📊 Analyzing commit history...")
	fmt.Println()

	// Get commit statistics
	stats, err := getCommitStats()
	if err != nil {
		color.Red("❌ Failed to get commit statistics: %v", err)
		return err
	}

	// Display insights
	displayCommitInsights(stats)

	return nil
}

type CommitStats struct {
	TotalCommits    int
	CommitTypes     map[string]int
	MostActiveFiles []string
	Authors         map[string]int
	RecentActivity  string
}

func getCommitStats() (*CommitStats, error) {
	stats := &CommitStats{
		CommitTypes: make(map[string]int),
		Authors:     make(map[string]int),
	}

	// Get total number of commits
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	fmt.Sscanf(string(output), "%d", &stats.TotalCommits)

	// Get commit type distribution with enhanced parsing
	cmd = exec.Command("git", "log", "--pretty=format:%s")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Enhanced commit type extraction
		commitType := extractCommitType(line)
		if commitType != "" {
			stats.CommitTypes[commitType]++
		}
	}

	// Get most active files with better analysis
	cmd = exec.Command("git", "log", "--name-only", "--pretty=format:")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	fileCounts := make(map[string]int)
	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "commit") {
			fileCounts[line]++
		}
	}

	// Get top 5 most active files with better sorting
	stats.MostActiveFiles = getTopFilesSorted(fileCounts, 5)

	// Get author statistics with enhanced parsing
	cmd = exec.Command("git", "shortlog", "-sn")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var count int
		var author string
		fmt.Sscanf(line, "%d\t%s", &count, &author)
		stats.Authors[author] = count
	}

	// Get recent activity with more detailed analysis
	cmd = exec.Command("git", "log", "--since=1 week ago", "--oneline")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	recentCommits := strings.Split(string(output), "\n")
	stats.RecentActivity = fmt.Sprintf("%d commits in the last week", len(recentCommits)-1)

	return stats, nil
}

// extractCommitType extracts commit type with enhanced pattern matching
func extractCommitType(commitMessage string) string {
	// Remove breaking change indicator
	message := strings.TrimSuffix(commitMessage, "!")

	// Split by colon to get type and scope
	parts := strings.Split(message, ":")
	if len(parts) == 0 {
		return ""
	}

	commitType := strings.TrimSpace(parts[0])

	// Extract type from scope if present (e.g., "feat(api)" -> "feat")
	if strings.Contains(commitType, "(") {
		scopeStart := strings.Index(commitType, "(")
		commitType = strings.TrimSpace(commitType[:scopeStart])
	}

	// Validate commit type
	validTypes := []string{
		"feat", "fix", "refactor", "chore", "test", "docs",
		"style", "perf", "ci", "build", "security", "config",
		"deploy", "revert", "wip", "hotfix", "patch", "release",
	}

	for _, validType := range validTypes {
		if strings.EqualFold(commitType, validType) {
			return strings.ToLower(commitType)
		}
	}

	// If not a conventional commit, try to infer from message content
	return inferCommitTypeFromMessage(commitMessage)
}

// inferCommitTypeFromMessage infers commit type from message content
func inferCommitTypeFromMessage(message string) string {
	lowerMessage := strings.ToLower(message)

	// Feature patterns
	if strings.Contains(lowerMessage, "add") || strings.Contains(lowerMessage, "new") ||
		strings.Contains(lowerMessage, "implement") || strings.Contains(lowerMessage, "create") ||
		strings.Contains(lowerMessage, "introduce") || strings.Contains(lowerMessage, "feature") {
		return "feat"
	}

	// Fix patterns
	if strings.Contains(lowerMessage, "fix") || strings.Contains(lowerMessage, "bug") ||
		strings.Contains(lowerMessage, "issue") || strings.Contains(lowerMessage, "problem") ||
		strings.Contains(lowerMessage, "error") || strings.Contains(lowerMessage, "resolve") ||
		strings.Contains(lowerMessage, "correct") || strings.Contains(lowerMessage, "patch") {
		return "fix"
	}

	// Documentation patterns
	if strings.Contains(lowerMessage, "doc") || strings.Contains(lowerMessage, "readme") ||
		strings.Contains(lowerMessage, "comment") || strings.Contains(lowerMessage, "guide") ||
		strings.Contains(lowerMessage, "tutorial") || strings.Contains(lowerMessage, "example") {
		return "docs"
	}

	// Test patterns
	if strings.Contains(lowerMessage, "test") || strings.Contains(lowerMessage, "spec") ||
		strings.Contains(lowerMessage, "coverage") || strings.Contains(lowerMessage, "assert") ||
		strings.Contains(lowerMessage, "mock") || strings.Contains(lowerMessage, "stub") {
		return "test"
	}

	// Performance patterns
	if strings.Contains(lowerMessage, "perf") || strings.Contains(lowerMessage, "optimize") ||
		strings.Contains(lowerMessage, "speed") || strings.Contains(lowerMessage, "fast") ||
		strings.Contains(lowerMessage, "efficient") || strings.Contains(lowerMessage, "cache") {
		return "perf"
	}

	// Refactor patterns
	if strings.Contains(lowerMessage, "refactor") || strings.Contains(lowerMessage, "restructure") ||
		strings.Contains(lowerMessage, "reorganize") || strings.Contains(lowerMessage, "clean") ||
		strings.Contains(lowerMessage, "simplify") || strings.Contains(lowerMessage, "improve") {
		return "refactor"
	}

	// Style patterns
	if strings.Contains(lowerMessage, "style") || strings.Contains(lowerMessage, "format") ||
		strings.Contains(lowerMessage, "lint") || strings.Contains(lowerMessage, "prettier") ||
		strings.Contains(lowerMessage, "indent") || strings.Contains(lowerMessage, "spacing") {
		return "style"
	}

	// Security patterns
	if strings.Contains(lowerMessage, "security") || strings.Contains(lowerMessage, "auth") ||
		strings.Contains(lowerMessage, "password") || strings.Contains(lowerMessage, "encrypt") ||
		strings.Contains(lowerMessage, "vulnerability") || strings.Contains(lowerMessage, "secure") {
		return "security"
	}

	// Configuration patterns
	if strings.Contains(lowerMessage, "config") || strings.Contains(lowerMessage, "setting") ||
		strings.Contains(lowerMessage, "env") || strings.Contains(lowerMessage, "configure") ||
		strings.Contains(lowerMessage, "setup") || strings.Contains(lowerMessage, "init") {
		return "config"
	}

	// Deployment patterns
	if strings.Contains(lowerMessage, "deploy") || strings.Contains(lowerMessage, "docker") ||
		strings.Contains(lowerMessage, "kubernetes") || strings.Contains(lowerMessage, "k8s") ||
		strings.Contains(lowerMessage, "helm") || strings.Contains(lowerMessage, "infrastructure") {
		return "deploy"
	}

	// CI/CD patterns
	if strings.Contains(lowerMessage, "ci") || strings.Contains(lowerMessage, "cd") ||
		strings.Contains(lowerMessage, "pipeline") || strings.Contains(lowerMessage, "workflow") ||
		strings.Contains(lowerMessage, "github") || strings.Contains(lowerMessage, "gitlab") ||
		strings.Contains(lowerMessage, "jenkins") || strings.Contains(lowerMessage, "travis") {
		return "ci"
	}

	// Build patterns
	if strings.Contains(lowerMessage, "build") || strings.Contains(lowerMessage, "webpack") ||
		strings.Contains(lowerMessage, "rollup") || strings.Contains(lowerMessage, "vite") ||
		strings.Contains(lowerMessage, "compile") || strings.Contains(lowerMessage, "bundle") {
		return "build"
	}

	// Revert patterns
	if strings.Contains(lowerMessage, "revert") || strings.Contains(lowerMessage, "rollback") ||
		strings.Contains(lowerMessage, "undo") || strings.Contains(lowerMessage, "restore") {
		return "revert"
	}

	// WIP patterns
	if strings.Contains(lowerMessage, "wip") || strings.Contains(lowerMessage, "work in progress") ||
		strings.Contains(lowerMessage, "draft") || strings.Contains(lowerMessage, "temporary") {
		return "wip"
	}

	// Default to chore for maintenance tasks
	return "chore"
}

// getTopFilesSorted returns top files sorted by frequency
func getTopFilesSorted(fileCounts map[string]int, limit int) []string {
	type fileCount struct {
		file  string
		count int
	}

	var files []fileCount
	for file, count := range fileCounts {
		files = append(files, fileCount{file: file, count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].count < files[j].count {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Return top files
	var result []string
	for i := 0; i < limit && i < len(files); i++ {
		result = append(result, files[i].file)
	}

	return result
}

func displayCommitInsights(stats *CommitStats) {
	color.Cyan("📈 Commit History Insights")
	fmt.Println()

	// Overall statistics
	color.Green("📊 Overall Statistics:")
	color.White("   Total commits: %d", stats.TotalCommits)
	color.White("   Recent activity: %s", stats.RecentActivity)
	fmt.Println()

	// Commit type distribution
	color.Green("🎯 Commit Type Distribution:")
	for commitType, count := range stats.CommitTypes {
		percentage := float64(count) / float64(stats.TotalCommits) * 100
		color.White("   %s: %d (%.1f%%)", commitType, count, percentage)
	}
	fmt.Println()

	// Most active files
	color.Green("📁 Most Active Files:")
	for i, file := range stats.MostActiveFiles {
		color.White("   %d. %s", i+1, file)
	}
	fmt.Println()

	// Top authors
	color.Green("👥 Top Contributors:")
	count := 0
	for author, commits := range stats.Authors {
		if count >= 5 {
			break
		}
		color.White("   %s: %d commits", author, commits)
		count++
	}
	fmt.Println()

	// Recommendations
	color.Green("💡 Recommendations:")
	if stats.CommitTypes["feat"] > stats.CommitTypes["fix"]*2 {
		color.Yellow("   ⚠️  Consider adding more bug fixes - feature commits are much higher than fixes")
	}
	if stats.CommitTypes["docs"] < stats.TotalCommits/10 {
		color.Yellow("   📚 Consider adding more documentation commits")
	}
	if stats.CommitTypes["test"] < stats.TotalCommits/5 {
		color.Yellow("   🧪 Consider adding more test commits")
	}
	if len(stats.Authors) == 1 {
		color.Yellow("   👥 Single contributor detected - consider code reviews for better quality")
	}
}
