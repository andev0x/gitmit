package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/andev0x/gitmit/internal/llm"
	"github.com/andev0x/gitmit/internal/prompt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var proposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Propose a commit message from a git diff",
	RunE:  runPropose,
}

func init() {
	rootCmd.AddCommand(proposeCmd)
}

func runPropose(cmd *cobra.Command, args []string) error {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	diff := string(bytes)

	if diff == "" {
		return fmt.Errorf("diff is empty")
	}

	var openAIAPIKey string
	if useOpenAI {
		// Temporarily create a prompt instance to get the key
		tempPrompt := prompt.New("") // Pass empty string for now
		key, err := tempPrompt.PromptForOpenAIKey()
		if err != nil {
			return err
		}
		openAIAPIKey = key
	}

	initialMessage, err := llm.ProposeCommit(cmd.Context(), openAIAPIKey, diff)
	if err != nil {
		return err
	}

	color.Green(initialMessage)

	p := prompt.New(openAIAPIKey)
	finalMessage, err := p.PromptForMessage(initialMessage, diff)
	if err != nil {
		return err
	}

	if finalMessage != "" {
		fmt.Println("\nFinal commit message:")
		color.Green(finalMessage)
	}

	return nil
}
