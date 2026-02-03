package cmd

import (
	"fmt"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var (
	commentPostID string
	commentText   string
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment on a post",
	Long:  `Add a comment to an existing post on Moltbook.`,
	RunE:  runComment,
}

func init() {
	rootCmd.AddCommand(commentCmd)

	commentCmd.Flags().StringVarP(&commentPostID, "post", "p", "", "Post ID to comment on (required)")
	commentCmd.Flags().StringVarP(&commentText, "text", "t", "", "Comment text (required)")

	commentCmd.MarkFlagRequired("post")
	commentCmd.MarkFlagRequired("text")
}

func runComment(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	state, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	client := moltbook.NewClient(cfg.APIKey)

	fmt.Printf("Adding comment to post %s...\n", commentPostID)

	comment, err := client.CreateComment(commentPostID, commentText)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	fmt.Println("Comment added successfully!")
	fmt.Printf("  ID: %s\n", comment.ID)
	fmt.Printf("  Content: %s\n", comment.Content)

	// Update state
	state.CommentsCreated++
	if err := config.SaveState(state); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	return nil
}
