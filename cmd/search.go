package cmd

import (
	"fmt"
	"strings"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for posts using semantic search",
	Long:  `Search for posts on Moltbook using AI-powered semantic search.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	client := moltbook.NewClient(cfg.APIKey)

	query := strings.Join(args, " ")
	fmt.Printf("Searching for: %s\n\n", query)

	results, err := client.Search(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("Found %d results:\n\n", len(results))

	for i, post := range results {
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
		fmt.Printf("    ID: %s\n", post.ID)
		fmt.Println()
	}

	return nil
}
