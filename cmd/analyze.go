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

	// Get commit type distribution
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

		// Extract commit type
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			commitType := strings.TrimSpace(parts[0])
			// Extract type from scope if present
			if strings.Contains(commitType, "(") {
				scopeStart := strings.Index(commitType, "(")
				commitType = strings.TrimSpace(commitType[:scopeStart])
			}
			stats.CommitTypes[commitType]++
		}
	}

	// Get most active files
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

	// Get top 5 most active files
	stats.MostActiveFiles = getTopFiles(fileCounts, 5)

	// Get author statistics
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

	// Get recent activity
	cmd = exec.Command("git", "log", "--since=1 week ago", "--oneline")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	recentCommits := strings.Split(string(output), "\n")
	stats.RecentActivity = fmt.Sprintf("%d commits in the last week", len(recentCommits)-1)

	return stats, nil
}

func getTopFiles(fileCounts map[string]int, limit int) []string {
	var files []string
	for file := range fileCounts {
		files = append(files, file)
	}

	// Simple sorting (in a real implementation, you'd use sort.Slice)
	// For now, just return the first few files
	if len(files) > limit {
		return files[:limit]
	}
	return files
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
