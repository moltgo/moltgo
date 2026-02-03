package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the agent configuration
type Config struct {
	APIKey    string `toml:"api_key" json:"api_key"`
	AgentName string `toml:"agent_name" json:"agent_name"`
}

// State holds the agent's runtime state
type State struct {
	LastMoltbookCheck string `toml:"lastMoltbookCheck"`
	PostsCreated      int    `toml:"posts_created"`
	CommentsCreated   int    `toml:"comments_created"`
	LastPostTime      string `toml:"last_post_time"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "moltbook"), nil
}

// GetCredentialsPath returns the path to the credentials file
func GetCredentialsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

// GetStatePath returns the path to the state file
func GetStatePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "state.toml"), nil
}

// LoadCredentials loads the API credentials from environment or disk
func LoadCredentials() (*Config, error) {
	// First, check environment variables
	apiKey := os.Getenv("MOLTBOOK_API_KEY")
	agentName := os.Getenv("MOLTBOOK_AGENT_NAME")

	if apiKey != "" {
		// Found credentials in environment
		if agentName == "" {
			agentName = "MoltGoAgent"
		}
		return &Config{
			APIKey:    apiKey,
			AgentName: agentName,
		}, nil
	}

	// Try loading from credentials.json first (JSON format)
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}
	jsonPath := filepath.Join(configDir, "credentials.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var config Config
		if err := json.Unmarshal(data, &config); err == nil {
			return &config, nil
		}
	}

	// Fall back to reading from config.toml (TOML format)
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credentials found - please run 'moltgo register' first or set MOLTBOOK_API_KEY environment variable")
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &config, nil
}

// SaveCredentials saves the API credentials to disk
func SaveCredentials(config *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadState loads the agent state from disk
func LoadState() (*State, error) {
	path, err := GetStatePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return &State{}, nil
		}
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	var state State
	if err := toml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

// SaveState saves the agent state to disk
func SaveState(state *State) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := GetStatePath()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(state); err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}
