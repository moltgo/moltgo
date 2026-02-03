package cmd

import (
	"fmt"
	"time"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Perform periodic check-in with Moltbook",
	Long: `Perform a heartbeat check-in with Moltbook. This should be run every 4+ hours
to keep your agent active and engaged with the community.`,
	RunE: runHeartbeat,
}

func init() {
	rootCmd.AddCommand(heartbeatCmd)
}

func runHeartbeat(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	state, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	client := moltbook.NewClient(cfg.APIKey)

	now := time.Now()
	fmt.Printf("Heartbeat check at %s\n\n", now.Format("2006-01-02 15:04:05"))

	// Browse recent posts
	fmt.Println("Browsing recent posts...")
	posts, err := client.BrowsePosts(&moltbook.BrowsePostsRequest{
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to browse posts: %w", err)
	}

	if len(posts) > 0 {
		fmt.Printf("\nFound %d recent posts:\n\n", len(posts))
		for i, post := range posts {
			if i >= 3 { // Show only top 3
				break
			}
			fmt.Printf("  [%d] %s\n", i+1, post.Title)
			fmt.Printf("      by %s in /%s\n", post.Author, post.Submolt)
			fmt.Printf("      Score: %d | Comments: %d\n", post.Score, post.NumComments)
			fmt.Println()
		}
	} else {
		fmt.Println("  No posts found.")
	}

	// Update state
	state.LastMoltbookCheck = now.Format(time.RFC3339)
	if err := config.SaveState(state); err != nil {
		fmt.Printf("\nWarning: failed to save state: %v\n", err)
	}

	fmt.Println("Heartbeat complete")

	// Show next check time
	nextCheck := now.Add(4 * time.Hour)
	fmt.Printf("\nNext heartbeat recommended: %s\n", nextCheck.Format("2006-01-02 15:04:05"))

	return nil
}
