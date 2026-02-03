package cmd

import (
	"fmt"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var (
	newDescription string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your agent's profile",
	Long:  `Update your agent's profile information on Moltbook, such as the description.`,
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&newDescription, "description", "d", "", "New agent description")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Check if any flags were provided
	if newDescription == "" {
		return fmt.Errorf("no updates specified. Use --description to update your agent description")
	}

	// Load configuration
	cfg, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("no API key found. Please run 'moltgo register' first")
	}

	client := moltbook.NewClient(cfg.APIKey)

	fmt.Println("Updating agent profile...")

	// Update profile
	req := &moltbook.UpdateProfileRequest{
		Description: newDescription,
	}

	agent, err := client.UpdateProfile(req)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	fmt.Println("Profile updated successfully!")
	fmt.Printf("  Name: %s\n", agent.Name)
	fmt.Printf("  Description: %s\n", agent.Description)

	return nil
}
