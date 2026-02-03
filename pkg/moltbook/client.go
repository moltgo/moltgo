package moltbook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://www.moltbook.com/api/v1"
)

// Client is the Moltbook API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Moltbook API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterRequest represents an agent registration request
type RegisterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RegisterResponse represents the response from registration
type RegisterResponse struct {
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
	Hint             string `json:"hint,omitempty"`
	Message          string `json:"message,omitempty"`
	Agent            *AgentRegistration `json:"agent,omitempty"`
	TweetTemplate    string `json:"tweet_template,omitempty"`
	// Legacy flat fields (for backward compatibility)
	APIKey           string `json:"api_key,omitempty"`
	AgentID          string `json:"agent_id,omitempty"`
	ClaimURL         string `json:"claim_url,omitempty"`
	VerificationCode string `json:"verification_code,omitempty"`
}

// AgentRegistration represents the agent data in registration response
type AgentRegistration struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	APIKey           string `json:"api_key"`
	ClaimURL         string `json:"claim_url"`
	VerificationCode string `json:"verification_code"`
	ProfileURL       string `json:"profile_url,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

// Post represents a Moltbook post
type Post struct {
	ID          string `json:"id"`
	Submolt     string `json:"submolt"`
	Title       string `json:"title"`
	Content     string `json:"content,omitempty"`
	URL         string `json:"url,omitempty"`
	Author      string `json:"author"`
	Score       int    `json:"score"`
	NumComments int    `json:"num_comments"`
	CreatedAt   string `json:"created_at"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Score     int    `json:"score"`
	CreatedAt string `json:"created_at"`
}

// Agent represents an agent profile
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// Register registers a new agent with Moltbook
func Register(name, description string) (*RegisterResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	reqData := RegisterRequest{
		Name:        name,
		Description: description,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", BaseURL+"/agents/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result RegisterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// If we can't parse the response, return the raw body
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("registration failed (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response body
	if !result.Success && result.Error != "" {
		errMsg := result.Error
		if result.Hint != "" {
			errMsg += " - " + result.Hint
		}
		return nil, fmt.Errorf("registration failed: %s", errMsg)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Normalize response - if agent field exists, copy values to top level for easier access
	if result.Agent != nil {
		if result.APIKey == "" {
			result.APIKey = result.Agent.APIKey
		}
		if result.AgentID == "" {
			result.AgentID = result.Agent.ID
		}
		if result.ClaimURL == "" {
			result.ClaimURL = result.Agent.ClaimURL
		}
		if result.VerificationCode == "" {
			result.VerificationCode = result.Agent.VerificationCode
		}
	}

	return &result, nil
}

// doRequest performs an authenticated API request
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := BaseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return nil, fmt.Errorf("rate limited (retry after %s seconds)", retryAfter)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetProfileResponse represents the response from getting agent profile
type GetProfileResponse struct {
	Success bool  `json:"success"`
	Agent   Agent `json:"agent"`
}

// GetProfile gets the authenticated agent's profile
func (c *Client) GetProfile() (*Agent, error) {
	data, err := c.doRequest("GET", "/agents/me", nil)
	if err != nil {
		return nil, err
	}

	var response GetProfileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &response.Agent, nil
}

// UpdateProfileRequest represents a request to update agent profile
type UpdateProfileRequest struct {
	Description string `json:"description,omitempty"`
}

// UpdateProfile updates the authenticated agent's profile
func (c *Client) UpdateProfile(req *UpdateProfileRequest) (*Agent, error) {
	data, err := c.doRequest("PATCH", "/agents/me", req)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &agent, nil
}

// BrowsePostsRequest contains parameters for browsing posts
type BrowsePostsRequest struct {
	Submolt string
	Limit   int
}

// BrowsePostsResponse represents the response from browsing posts
type BrowsePostsResponse struct {
	Posts []Post `json:"posts"`
}

// BrowsePosts retrieves recent posts
func (c *Client) BrowsePosts(req *BrowsePostsRequest) ([]Post, error) {
	endpoint := fmt.Sprintf("/posts?limit=%d", req.Limit)
	if req.Submolt != "" {
		endpoint += "&submolt=" + req.Submolt
	}

	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response BrowsePostsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse posts: %w", err)
	}

	return response.Posts, nil
}

// CreatePostRequest represents a request to create a post
type CreatePostRequest struct {
	Submolt string `json:"submolt"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

// CreatePost creates a new post
func (c *Client) CreatePost(req *CreatePostRequest) (*Post, error) {
	data, err := c.doRequest("POST", "/posts", req)
	if err != nil {
		return nil, err
	}

	var post Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, fmt.Errorf("failed to parse post: %w", err)
	}

	return &post, nil
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// CreateComment creates a comment on a post
func (c *Client) CreateComment(postID string, content string) (*Comment, error) {
	req := CreateCommentRequest{Content: content}
	endpoint := fmt.Sprintf("/posts/%s/comments", postID)

	data, err := c.doRequest("POST", endpoint, req)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment: %w", err)
	}

	return &comment, nil
}

// VoteRequest represents a vote request
type VoteRequest struct {
	TargetType string `json:"target_type"` // "post" or "comment"
	TargetID   string `json:"target_id"`
	Direction  string `json:"direction"` // "up" or "down"
}

// Vote votes on a post or comment
func (c *Client) Vote(targetType, targetID, direction string) error {
	req := VoteRequest{
		TargetType: targetType,
		TargetID:   targetID,
		Direction:  direction,
	}

	_, err := c.doRequest("POST", "/vote", req)
	return err
}

// SearchResponse represents search results
type SearchResponse struct {
	Results []Post `json:"results"`
}

// Search performs semantic search for posts
func (c *Client) Search(query string) ([]Post, error) {
	endpoint := fmt.Sprintf("/search?q=%s", query)

	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return response.Results, nil
}
