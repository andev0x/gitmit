package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.1"
	// Global flags
	interactiveFlag bool
	suggestionsFlag bool

	rootCmd = &cobra.Command{
		Use:   "gitmit",
		Short: "🧠 Smart Git Commit Message Generator",
		Long: `Gitmit is a lightweight CLI tool that analyzes your staged changes
and suggests professional commit messages following Conventional Commits format.

Examples:
  gitmit                    # Analyze changes and suggest commit message
  gitmit propose           # Same as above
  gitmit propose -i       # Interactive mode with multiple suggestions
  gitmit propose -s       # Show multiple suggestions
  gitmit propose --auto   # Auto-commit with best suggestion`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Add global validation or setup here
			if suggestionsFlag {
				interactiveFlag = true // -s implies -i
			}
		},
	}
)

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactive mode with multiple suggestions")
	rootCmd.PersistentFlags().BoolVarP(&suggestionsFlag, "suggestions", "s", false, "Show multiple ranked suggestions")
}

func Execute() error {
	// ✅ Added: if no subcommand provided, fallback to "propose"
	if len(os.Args) == 1 {
		return proposeCmd.RunE(rootCmd, nil)
	}
	return rootCmd.Execute()
}
