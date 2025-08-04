package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// InteractivePrompt handles user interaction for commit message customization
type InteractivePrompt struct {
	reader *bufio.Reader
}

// New creates a new InteractivePrompt instance
func New() *InteractivePrompt {
	return &InteractivePrompt{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptForMessage prompts the user to accept, reject, or edit the suggested message
func (p *InteractivePrompt) PromptForMessage(suggestedMessage string) (string, error) {
	for {
		fmt.Print("Accept this message? (y/n/e to edit): ")
		
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
		default:
			color.Yellow("Please enter 'y' (yes), 'n' (no), or 'e' (edit)")
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