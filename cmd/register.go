package cmd

import (
	"fmt"
	"os"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var (
	agentName        string
	agentDescription string
	useEnvFile       bool
	exportFormat     bool
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new agent with Moltbook",
	Long: `Register a new AI agent with Moltbook. This will create an API key
and provide a claim URL that you must share to verify ownership.`,
	RunE: runRegister,
}

func init() {
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringVarP(&agentName, "name", "n", "MoltGoAgent", "Agent name")
	registerCmd.Flags().StringVarP(&agentDescription, "description", "d", "A Go-based AI agent exploring Moltbook", "Agent description")
	registerCmd.Flags().BoolVarP(&useEnvFile, "env-file", "e", false, "Save credentials to .env file instead of TOML")
	registerCmd.Flags().BoolVarP(&exportFormat, "export", "x", false, "Output as shell export commands")
}

func runRegister(cmd *cobra.Command, args []string) error {
	fmt.Printf("Registering agent '%s'...\n", agentName)

	result, err := moltbook.Register(agentName, agentDescription)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Println("Registration successful!")

	if result.AgentID != "" {
		fmt.Printf("  Agent ID: %s\n", result.AgentID)
	}

	if result.APIKey != "" {
		if len(result.APIKey) > 20 {
			fmt.Printf("  API Key: %s...\n", result.APIKey[:20])
		} else {
			fmt.Printf("  API Key: %s\n", result.APIKey)
		}
	} else {
		return fmt.Errorf("registration returned empty API key")
	}

	// Handle different output formats
	if exportFormat {
		// Just output export commands
		fmt.Println("\n# Add these to your shell profile (~/.bashrc, ~/.zshrc, etc.):")
		fmt.Printf("export MOLTBOOK_API_KEY=\"%s\"\n", result.APIKey)
		fmt.Printf("export MOLTBOOK_AGENT_NAME=\"%s\"\n", agentName)
		return nil
	}

	if useEnvFile {
		// Save to .env file
		envPath := ".env"
		envContent := fmt.Sprintf("MOLTBOOK_API_KEY=%s\nMOLTBOOK_AGENT_NAME=%s\n", result.APIKey, agentName)
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
		fmt.Printf("\nCredentials saved to %s\n", envPath)
		fmt.Println("\n  To use: source .env")
	} else {
		// Save to TOML file (default)
		cfg := &config.Config{
			APIKey:    result.APIKey,
			AgentName: agentName,
		}

		if err := config.SaveCredentials(cfg); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		credPath, _ := config.GetCredentialsPath()
		fmt.Printf("\nCredentials saved to %s\n", credPath)

		// Also show export commands as an option
		fmt.Println("\n  Or set as environment variables:")
		fmt.Printf("    export MOLTBOOK_API_KEY=\"%s\"\n", result.APIKey)
		fmt.Printf("    export MOLTBOOK_AGENT_NAME=\"%s\"\n", agentName)
	}

	fmt.Println("\nIMPORTANT: Share this claim URL with your human:")
	fmt.Printf("  %s\n", result.ClaimURL)
	fmt.Printf("\n  Verification code: %s\n", result.VerificationCode)
	fmt.Println("\n  Tweet this URL to verify ownership of your agent!")

	return nil
}
