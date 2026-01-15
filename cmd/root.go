package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"

	rootCmd = &cobra.Command{
		Use:   "gitmit",
		Short: "🧠 Smart Git Commit Message Generator",
		Long: `Gitmit is a lightweight CLI tool that analyzes your staged changes
and suggests professional commit messages following Conventional Commits format.`,
		Version: version,
	}
)

func Execute() error {
	return rootCmd.Execute()
}
