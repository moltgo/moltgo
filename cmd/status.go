package cmd

import (
	"fmt"
	"time"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent status and statistics",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load credentials
	cfg, err := config.LoadCredentials()
	if err != nil {
		fmt.Println("Moltbook Agent Status")
		fmt.Println("  Status: Not registered")
		fmt.Println("\n  Run 'moltgo register' to get started!")
		return nil
	}

	// Load state
	state, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	fmt.Println("Moltbook Agent Status")
	fmt.Printf("  Name: %s\n", cfg.AgentName)
	fmt.Println("  Status: Registered")
	fmt.Printf("  API Key: %s...\n", cfg.APIKey[:20])

	// Fetch profile from API to get description
	client := moltbook.NewClient(cfg.APIKey)
	profile, err := client.GetProfile()
	if err == nil {
		if profile.ID != "" {
			fmt.Printf("  Agent ID: %s\n", profile.ID)
		}
		if profile.Description != "" {
			fmt.Printf("  Description: %s\n", profile.Description)
		}
	}

	fmt.Println("\n  Statistics:")
	fmt.Printf("    Posts created: %d\n", state.PostsCreated)
	fmt.Printf("    Comments created: %d\n", state.CommentsCreated)

	if state.LastMoltbookCheck != "" {
		lastCheck, err := time.Parse(time.RFC3339, state.LastMoltbookCheck)
		if err == nil {
			fmt.Printf("    Last check: %s\n", lastCheck.Format("2006-01-02 15:04:05"))
			timeSince := time.Since(lastCheck)
			fmt.Printf("    Time since last check: %s\n", formatDuration(timeSince))
		}
	}

	if state.LastPostTime != "" {
		lastPost, err := time.Parse(time.RFC3339, state.LastPostTime)
		if err == nil {
			fmt.Printf("    Last post: %s\n", lastPost.Format("2006-01-02 15:04:05"))
		}
	}

	credPath, _ := config.GetCredentialsPath()
	statePath, _ := config.GetStatePath()
	fmt.Printf("\n  Config files:\n")
	fmt.Printf("    Credentials: %s\n", credPath)
	fmt.Printf("    State: %s\n", statePath)

	return nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
