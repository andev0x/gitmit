package prompt

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/andev0x/gitmit/internal/llm"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// InteractivePrompt handles user interaction for commit message customization
type InteractivePrompt struct {
	reader       *bufio.Reader
	openAIAPIKey string
}

// New creates a new InteractivePrompt instance
func New(openAIAPIKey string) *InteractivePrompt {
	return &InteractivePrompt{
		reader:       bufio.NewReader(os.Stdin),
		openAIAPIKey: openAIAPIKey,
	}
}

// PromptForOpenAIKey securely prompts the user for their OpenAI API key.
func (p *InteractivePrompt) PromptForOpenAIKey() (string, error) {
	color.Yellow("\n🔑 Please enter your OpenAI API Key (input will be hidden):")
	color.Yellow("    (This key will NOT be saved to disk or committed to Git.)")
	fmt.Print("> ")

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read OpenAI API key: %w", err)
	}

	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

// PromptForMessage prompts the user to accept, reject, or edit the suggested message
func (p *InteractivePrompt) PromptForMessage(suggestedMessage, diff string) (string, error) {

	for {
		fmt.Print("Accept this message? (y/n/e/r): ")

		input, err := p.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		choice := strings.ToLower(strings.TrimSpace(input))

		switch choice {
		case "y", "yes", "":
			return suggestedMessage, nil
		case "n", "no":
			return "", nil
		case "e", "edit":
			return p.promptForEdit(suggestedMessage)
		case "r", "regenerate":
			return p.promptForRegenerate(diff)
		default:
			color.Yellow("Please enter 'y' (yes), 'n' (no), 'e' (edit) or 'r' (regenerate)")
		}
	}
}

// promptForEdit allows the user to enter a custom commit message
func (p *InteractivePrompt) promptForEdit(originalMessage string) (string, error) {
	fmt.Println("\nEnter your custom commit message:")
	fmt.Printf("(Press Enter to keep original: %s)\n", originalMessage)
	fmt.Print("> ")

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read custom message: %w", err)
	}

	customMessage := strings.TrimSpace(input)
	if customMessage == "" {
		return originalMessage, nil
	}

	return customMessage, nil
}

func (p *InteractivePrompt) promptForRegenerate(diff string) (string, error) {
	color.Yellow("Regenerating commit message...")
	newMessage, err := llm.ProposeCommit(context.Background(), p.openAIAPIKey, diff)
	if err != nil {
		if strings.Contains(err.Error(), "OPENAI_API_KEY not set") {
			color.Red("Error: OpenAI API Key is not set. Please provide a valid key.")
			newKey, keyErr := p.PromptForOpenAIKey()
			if keyErr != nil {
				return "", keyErr
			}
			p.openAIAPIKey = newKey
			// Retry proposing the commit with the new key
			newMessage, err = llm.ProposeCommit(context.Background(), p.openAIAPIKey, diff)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	color.Green(newMessage)
	return p.PromptForMessage(newMessage, diff)
}
