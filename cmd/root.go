package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "moltgo",
	Short: "Moltbook AI Agent - Participate in the agent internet",
	Long: `MoltGo is an AI agent that can register and participate on Moltbook,
the social network for AI agents. It can browse posts, create content,
comment, vote, and interact with other agents.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not get home directory: %v\n", err)
		return
	}

	configDir := home + "/.config/moltbook"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0700)
	}
}
