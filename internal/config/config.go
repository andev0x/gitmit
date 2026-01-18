package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the structure of .gitmit.json
type Config struct {
	TopicMappings     map[string]string            `json:"topicMappings"`
	KeywordMappings   map[string]string            `json:"keywordMappings"`
	ProjectType       string                       `json:"projectType"`       // go, nodejs, python, etc.
	Keywords          map[string]map[string]int    `json:"keywords"`          // action -> keyword -> score
	Templates         map[string]map[string]string `json:"templates"`         // Custom templates
	DiffStatThreshold float64                      `json:"diffStatThreshold"` // Threshold for add/delete ratio
}

// LoadConfig loads the configuration with hierarchy: Local (.gitmit.json) → Global (~/.gitmit.json) → Default (embedded)
func LoadConfig() (*Config, error) {
	// Initialize with default empty config
	cfg := &Config{
		TopicMappings:     make(map[string]string),
		KeywordMappings:   make(map[string]string),
		Keywords:          make(map[string]map[string]int),
		Templates:         make(map[string]map[string]string),
		DiffStatThreshold: 0.5,
	}

	// 1. Try to load embedded default config (optional)
	// For now, we'll use the hardcoded defaults above

	// 2. Try to load global config from ~/.gitmit.json
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(homeDir, ".gitmit.json")
		if err := mergeConfigFromFile(cfg, globalConfigPath); err == nil {
			// Successfully loaded global config
		}
	}

	// 3. Try to load local config from .gitmit.json in current working directory
	localConfigPath := ".gitmit.json"
	if err := mergeConfigFromFile(cfg, localConfigPath); err == nil {
		// Successfully loaded local config
	}

	// Also support legacy .commit_suggest.json for backward compatibility
	legacyConfigPath := ".commit_suggest.json"
	if err := mergeConfigFromFile(cfg, legacyConfigPath); err == nil {
		// Successfully loaded legacy config
	}

	// Auto-detect project type if not specified
	if cfg.ProjectType == "" {
		cfg.ProjectType = DetectProjectType()
	}

	// Load language-specific defaults based on project type
	loadLanguageDefaults(cfg)

	return cfg, nil
}

// DetectProjectType automatically detects the project type by checking for characteristic files
func DetectProjectType() string {
	// Check for Go project
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}

	// Check for Node.js project
	if _, err := os.Stat("package.json"); err == nil {
		return "nodejs"
	}

	// Check for Python project
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}
	if _, err := os.Stat("setup.py"); err == nil {
		return "python"
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python"
	}

	// Check for Java project
	if _, err := os.Stat("pom.xml"); err == nil {
		return "java"
	}
	if _, err := os.Stat("build.gradle"); err == nil {
		return "java"
	}

	// Check for Ruby project
	if _, err := os.Stat("Gemfile"); err == nil {
		return "ruby"
	}

	// Check for Rust project
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "rust"
	}

	// Check for PHP project
	if _, err := os.Stat("composer.json"); err == nil {
		return "php"
	}

	return "generic"
}

// loadLanguageDefaults loads language-specific keyword mappings and scoring
func loadLanguageDefaults(cfg *Config) {
	switch cfg.ProjectType {
	case "go":
		// Go-specific keywords
		if cfg.Keywords["feat"] == nil {
			cfg.Keywords["feat"] = make(map[string]int)
		}
		cfg.Keywords["feat"]["func"] = 3
		cfg.Keywords["feat"]["type"] = 2
		cfg.Keywords["feat"]["struct"] = 2
		cfg.Keywords["feat"]["interface"] = 2

		if cfg.Keywords["fix"] == nil {
			cfg.Keywords["fix"] = make(map[string]int)
		}
		cfg.Keywords["fix"]["if err != nil"] = 3
		cfg.Keywords["fix"]["error"] = 2
		cfg.Keywords["fix"]["panic"] = 2

	case "nodejs":
		// Node.js-specific keywords
		if cfg.Keywords["feat"] == nil {
			cfg.Keywords["feat"] = make(map[string]int)
		}
		cfg.Keywords["feat"]["function"] = 3
		cfg.Keywords["feat"]["class"] = 2
		cfg.Keywords["feat"]["const"] = 1
		cfg.Keywords["feat"]["export"] = 2

		if cfg.Keywords["fix"] == nil {
			cfg.Keywords["fix"] = make(map[string]int)
		}
		cfg.Keywords["fix"]["try"] = 2
		cfg.Keywords["fix"]["catch"] = 2
		cfg.Keywords["fix"]["throw"] = 2

	case "python":
		// Python-specific keywords
		if cfg.Keywords["feat"] == nil {
			cfg.Keywords["feat"] = make(map[string]int)
		}
		cfg.Keywords["feat"]["def"] = 3
		cfg.Keywords["feat"]["class"] = 2
		cfg.Keywords["feat"]["async def"] = 3

		if cfg.Keywords["fix"] == nil {
			cfg.Keywords["fix"] = make(map[string]int)
		}
		cfg.Keywords["fix"]["try"] = 2
		cfg.Keywords["fix"]["except"] = 2
		cfg.Keywords["fix"]["raise"] = 2
	}
}

// mergeConfigFromFile loads a config file and merges it into the existing config
func mergeConfigFromFile(cfg *Config, path string) error {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %w", path, err)
	}

	var fileCfg Config
	err = json.Unmarshal(data, &fileCfg)
	if err != nil {
		return fmt.Errorf("error unmarshaling config file %s: %w", path, err)
	}

	// Merge the loaded config into the existing config
	// Topic mappings
	if fileCfg.TopicMappings != nil {
		for k, v := range fileCfg.TopicMappings {
			cfg.TopicMappings[k] = v
		}
	}

	// Keyword mappings
	if fileCfg.KeywordMappings != nil {
		for k, v := range fileCfg.KeywordMappings {
			cfg.KeywordMappings[k] = v
		}
	}

	// Project type (override if specified)
	if fileCfg.ProjectType != "" {
		cfg.ProjectType = fileCfg.ProjectType
	}

	// Keywords
	if fileCfg.Keywords != nil {
		for action, keywords := range fileCfg.Keywords {
			if cfg.Keywords[action] == nil {
				cfg.Keywords[action] = make(map[string]int)
			}
			for keyword, score := range keywords {
				cfg.Keywords[action][keyword] = score
			}
		}
	}

	// Templates
	if fileCfg.Templates != nil {
		for k, v := range fileCfg.Templates {
			cfg.Templates[k] = v
		}
	}

	// Diff stat threshold
	if fileCfg.DiffStatThreshold > 0 {
		cfg.DiffStatThreshold = fileCfg.DiffStatThreshold
	}

	return nil
}
