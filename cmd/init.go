package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gitmit/internal/config"
)

var (
	globalFlag bool

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a .gitmit.json configuration file",
		Long: `Generate a sample .gitmit.json configuration file with basic heuristic rules.

This allows you to customize gitmit's behavior without modifying source code.
You can create either a local config (in the current directory) or a global config (in your home directory).`,
		Example: `  gitmit init              # Create local .gitmit.json in current directory
  gitmit init --global    # Create global ~/.gitmit.json in home directory`,
		RunE: runInit,
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&globalFlag, "global", false, "Create global config in home directory (~/.gitmit.json)")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Detect project type automatically
	projectType := config.DetectProjectType()

	// Create sample configuration
	sampleConfig := config.Config{
		ProjectType:       projectType,
		DiffStatThreshold: 0.5,
		TopicMappings: map[string]string{
			"internal/api":      "api",
			"internal/database": "db",
			"internal/auth":     "auth",
			"internal/config":   "config",
			"cmd":               "cli",
			"pkg":               "core",
			"docs":              "docs",
		},
		KeywordMappings: map[string]string{
			"authentication": "auth",
			"database":       "db",
			"configuration":  "config",
		},
		Keywords: map[string]map[string]int{
			"feat": {
				"func":      3,
				"class":     2,
				"new":       2,
				"add":       2,
				"implement": 2,
			},
			"fix": {
				"bug":     3,
				"fix":     3,
				"error":   2,
				"issue":   2,
				"resolve": 2,
				"if err":  2,
				"try":     1,
				"catch":   1,
			},
			"refactor": {
				"refactor":    3,
				"restructure": 2,
				"rename":      2,
				"move":        2,
			},
			"test": {
				"test":   3,
				"Test":   3,
				"assert": 2,
				"expect": 2,
				"mock":   2,
			},
			"docs": {
				"docs":          3,
				"documentation": 3,
				"//":            1,
				"comment":       2,
			},
		},
		Templates: map[string]map[string]string{},
	}

	// Add language-specific keywords based on detected project type
	switch projectType {
	case "go":
		sampleConfig.Keywords["feat"]["type"] = 2
		sampleConfig.Keywords["feat"]["struct"] = 2
		sampleConfig.Keywords["feat"]["interface"] = 2
		sampleConfig.Keywords["fix"]["if err != nil"] = 3
		sampleConfig.Keywords["fix"]["panic"] = 2
	case "nodejs":
		sampleConfig.Keywords["feat"]["export"] = 2
		sampleConfig.Keywords["feat"]["const"] = 1
		sampleConfig.Keywords["fix"]["throw"] = 2
	case "python":
		sampleConfig.Keywords["feat"]["def"] = 3
		sampleConfig.Keywords["feat"]["async def"] = 3
		sampleConfig.Keywords["fix"]["except"] = 2
		sampleConfig.Keywords["fix"]["raise"] = 2
	}

	// Determine file path
	var configPath string
	if globalFlag {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}
		configPath = homeDir + "/.gitmit.json"
	} else {
		configPath = ".gitmit.json"
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		color.Yellow("⚠ Config file already exists: %s", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			color.Yellow("❌ Cancelled.")
			return nil
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(sampleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write to file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	color.Green("✅ Created config file: %s", configPath)
	color.Blue("\n📝 Detected project type: %s", projectType)
	fmt.Println("\nYou can now customize the configuration to fit your project's needs.")
	fmt.Println("\nConfiguration hierarchy:")
	fmt.Println("  1. Local (.gitmit.json) - project-specific settings")
	fmt.Println("  2. Global (~/.gitmit.json) - user-wide settings")
	fmt.Println("  3. Default (embedded) - built-in defaults")

	return nil
}
