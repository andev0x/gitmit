package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/andev0x/gitmit/internal/config"
)

// OllamaRequest represents the request body for Ollama's /api/generate endpoint
type OllamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature,omitempty"`
}

// OllamaResponse represents the response body from Ollama
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// OllamaClient handles communication with the local Ollama daemon
type OllamaClient struct {
	config config.OllamaConfig
}

// NewOllamaClient creates a new OllamaClient
func NewOllamaClient(cfg config.OllamaConfig) *OllamaClient {
	return &OllamaClient{config: cfg}
}

// Generate sends a prompt to Ollama and returns the generated response
func (c *OllamaClient) Generate(prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:       c.config.Model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: c.config.Temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling ollama request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.config.URL)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ollama daemon unreachable at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("model '%s' not found. please run: ollama pull %s", c.config.Model, c.config.Model)
		}
		return "", fmt.Errorf("ollama returned status code: %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("error decoding ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}
