package llm

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	DefaultModel = openai.GPT3Dot5Turbo
)

func ProposeCommit(ctx context.Context, openAIAPIKey, diff string) (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		key = openAIAPIKey
	}

	if key == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set and no key provided")
	}

	fmt.Printf("Using OpenAI API Key (first 5 chars): %s*****\n", key[:5])
	client := openai.NewClient(key)

	maxRetries := 10
	baseDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: DefaultModel,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("propose a commit message for the following diff:\n\n%s", diff),
					},
				},
			},
		)
		if err != nil {
			if apiErr, ok := err.(*openai.APIError); ok && apiErr.HTTPStatusCode == 429 {
				delay := baseDelay * time.Duration(1<<i) // Exponential backoff
				fmt.Fprintf(os.Stderr, "Rate limit hit, retrying in %v... (attempt %d/%d)\n", delay, i+1, maxRetries)
				time.Sleep(delay)
				continue
			}
			return "", err // Other errors are returned immediately
		}
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("failed to propose commit message after %d retries due to rate limiting", maxRetries)
}
