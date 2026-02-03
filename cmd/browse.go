package cmd

import (
	"fmt"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var (
	browseSubmolt string
	browseLimit   int
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse recent posts on Moltbook",
	Long:  `Browse and view recent posts from Moltbook. Optionally filter by submolt (community).`,
	RunE:  runBrowse,
}

func init() {
	rootCmd.AddCommand(browseCmd)

	browseCmd.Flags().StringVarP(&browseSubmolt, "submolt", "s", "", "Filter by submolt (community)")
	browseCmd.Flags().IntVarP(&browseLimit, "limit", "l", 10, "Number of posts to retrieve")
}

func runBrowse(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	client := moltbook.NewClient(cfg.APIKey)

	req := &moltbook.BrowsePostsRequest{
		Submolt: browseSubmolt,
		Limit:   browseLimit,
	}

	if browseSubmolt != "" {
		fmt.Printf("Browsing posts from /%s...\n\n", browseSubmolt)
	} else {
		fmt.Println("Browsing recent posts...")
	}

	posts, err := client.BrowsePosts(req)
	if err != nil {
		return fmt.Errorf("failed to browse posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found.")
		return nil
	}

	for i, post := range posts {
		fmt.Printf("[%d] %s\n", i+1, post.Title)
		fmt.Printf("    by %s in /%s\n", post.Author, post.Submolt)
		fmt.Printf("    Score: %d | Comments: %d\n", post.Score, post.NumComments)
		if post.Content != "" {
			content := post.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("    %s\n", content)
		}
		if post.URL != "" {
			fmt.Printf("    URL: %s\n", post.URL)
		}
		fmt.Printf("    ID: %s | Posted: %s\n", post.ID, post.CreatedAt)
		fmt.Println()
	}

	fmt.Printf("Total posts retrieved: %d\n", len(posts))

	return nil
}
