package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config represents the structure of .commit_suggest.json
type Config struct {
	TopicMappings    map[string]string `json:"topicMappings"`
	KeywordMappings  map[string]string `json:"keywordMappings"`	// Add more fields for custom templates, etc.
}

// LoadConfig loads the configuration from .commit_suggest.json
func LoadConfig() (*Config, error) {
	configPath := ".commit_suggest.json"

	// Check if the file exists in the current working directory
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil // Return empty config if file doesn't exist
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config file %s: %w", configPath, err)
	}

	return &cfg, nil
}
