package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AIContextKeeper/contextkeeper-agent/pkg/types"
)

// client implements SyncClient interface for ContextKeeper.dev API
type client struct {
	httpClient *http.Client
	baseURL    string
	sessionID  string
	apiKey     string
}

// NewClient creates a new sync client
func NewClient(config *types.Config, sessionMgr SessionManager) (*client, error) {
	return &client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   config.ServerURL,
		sessionID: sessionMgr.GetOrCreateSession(),
		apiKey:    config.APIKey,
	}, nil
}

// SessionManager interface (minimal definition for this package)
type SessionManager interface {
	GetOrCreateSession() string
}

// SaveSession saves a session to ContextKeeper.dev
func (c *client) SaveSession(session *types.Session) error {
	// Set session identifier
	session.UserSessionID = c.sessionID
	
	// Convert session to JSON
	jsonData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	// Create request
	url := fmt.Sprintf("%s/api/summaries", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	c.setHeaders(req)
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Handle response
	switch resp.StatusCode {
	case 200, 201:
		return nil
	case 400:
		return fmt.Errorf("bad request: missing or invalid session ID")
	case 401:
		return fmt.Errorf("unauthorized: invalid or missing API key")
	default:
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// GetUsageInfo retrieves current usage information
func (c *client) GetUsageInfo(sessionID string) (*types.UsageInfo, error) {
	url := fmt.Sprintf("%s/api/summaries", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	c.setHeaders(req)
	
	// Add query parameter for counting
	q := req.URL.Query()
	q.Add("count", "true")
	req.URL.RawQuery = q.Encode()
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	
	// Parse response
	var result struct {
		Count int `json:"count"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &types.UsageInfo{
		Current:    result.Count,
		Limit:      -1,
		Percentage: 0,
		HasReached: false,
	}, nil
}

// TestConnection tests the connection to ContextKeeper.dev
func (c *client) TestConnection() error {
	url := fmt.Sprintf("%s/api/summaries", c.baseURL)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("connection test failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// setHeaders sets the required headers for API requests
func (c *client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-session-id", c.sessionID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
}

