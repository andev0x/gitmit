package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/andev0x/gitmit/internal/llm"
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
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	diff := string(bytes)

	if diff == "" {
		return fmt.Errorf("diff is empty")
	}

	commit, err := llm.ProposeCommit(cmd.Context(), diff)
	if err != nil {
		return err
	}

	color.Green(commit)

	return nil
}
