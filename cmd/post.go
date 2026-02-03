package cmd

import (
	"fmt"
	"time"

	"github.com/moltgo/moltgo/pkg/config"
	"github.com/moltgo/moltgo/pkg/moltbook"
	"github.com/spf13/cobra"
)

var (
	postSubmolt string
	postTitle   string
	postContent string
	postURL     string
)

var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Create a new post on Moltbook",
	Long: `Create a new post on Moltbook. You must specify a submolt (community),
title, and either content or a URL.`,
	RunE: runPost,
}

func init() {
	rootCmd.AddCommand(postCmd)

	postCmd.Flags().StringVarP(&postSubmolt, "submolt", "s", "", "Submolt (community) to post in (required)")
	postCmd.Flags().StringVarP(&postTitle, "title", "t", "", "Post title (required)")
	postCmd.Flags().StringVarP(&postContent, "content", "c", "", "Post content (text)")
	postCmd.Flags().StringVarP(&postURL, "url", "u", "", "Post URL (link)")

	postCmd.MarkFlagRequired("submolt")
	postCmd.MarkFlagRequired("title")
}

func runPost(cmd *cobra.Command, args []string) error {
	if postContent == "" && postURL == "" {
		return fmt.Errorf("must provide either --content or --url")
	}

	cfg, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	// Load and check state for rate limiting
	state, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Check rate limit (1 post per 30 minutes)
	if state.LastPostTime != "" {
		lastPost, err := time.Parse(time.RFC3339, state.LastPostTime)
		if err == nil {
			timeSince := time.Since(lastPost)
			if timeSince < 30*time.Minute {
				waitTime := 30*time.Minute - timeSince
				return fmt.Errorf("rate limit: wait %d more minutes before posting", int(waitTime.Minutes())+1)
			}
		}
	}

	client := moltbook.NewClient(cfg.APIKey)

	req := &moltbook.CreatePostRequest{
		Submolt: postSubmolt,
		Title:   postTitle,
		Content: postContent,
		URL:     postURL,
	}

	fmt.Printf("Creating post in /%s...\n", postSubmolt)

	post, err := client.CreatePost(req)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	fmt.Println("Post created successfully!")
	fmt.Printf("  ID: %s\n", post.ID)
	fmt.Printf("  Title: %s\n", post.Title)
	fmt.Printf("  Submolt: /%s\n", post.Submolt)

	// Update state
	state.PostsCreated++
	state.LastPostTime = time.Now().Format(time.RFC3339)
	if err := config.SaveState(state); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	return nil
}
