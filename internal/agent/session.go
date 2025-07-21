package agent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/carsor007/contextkeeper-agent/pkg/types"
	"gopkg.in/yaml.v3"
)

// sessionManager implements SessionManager interface
type sessionManager struct {
	config     *types.Config
	configPath string
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *types.Config) (SessionManager, error) {
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".contextkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configPath := filepath.Join(configDir, "config.yaml")
	
	sm := &sessionManager{
		config:     config,
		configPath: configPath,
	}
	
	// Load existing config or create new one
	if err := sm.loadOrCreateConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	return sm, nil
}

// GetOrCreateSession returns the current session ID, creating one if needed
func (sm *sessionManager) GetOrCreateSession() string {
	if sm.config.SessionID == "" {
		sm.config.SessionID = sm.generateSessionID()
		sm.saveConfig()
	}
	return sm.config.SessionID
}

// GetUsageInfo returns current usage information
func (sm *sessionManager) GetUsageInfo() (*types.UsageInfo, error) {
	// For now, return mock data. In a real implementation,
	// this would query the ContextKeeper API or local cache
	return &types.UsageInfo{
		Current:    0,  // This would be fetched from API/cache
		Limit:      50, // Anonymous users have 50-save limit
		Percentage: 0,
		HasReached: false,
	}, nil
}

// IsAnonymous returns true if the user is anonymous (no API key)
func (sm *sessionManager) IsAnonymous() bool {
	return sm.config.APIKey == ""
}

// SetAPIKey sets the API key for authenticated users
func (sm *sessionManager) SetAPIKey(apiKey string) error {
	sm.config.APIKey = apiKey
	return sm.saveConfig()
}

// generateSessionID creates a new session ID like ContextKeeper.dev
// Format: "session_" + timestamp + "_" + random
func (sm *sessionManager) generateSessionID() string {
	timestamp := time.Now().UnixMilli()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)
	
	return fmt.Sprintf("session_%d_%s", timestamp, randomStr)
}

// loadOrCreateConfig loads existing config or creates a new one
func (sm *sessionManager) loadOrCreateConfig() error {
	// Try to load existing config
	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			sm.setDefaultConfig()
			return sm.saveConfig()
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse existing config
	if err := yaml.Unmarshal(data, sm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Ensure we have required fields
	if sm.config.ServerURL == "" {
		sm.config.ServerURL = "https://contextkeeper.dev"
	}
	if sm.config.LocalPort == 0 {
		sm.config.LocalPort = 8080
	}
	if sm.config.LogLevel == "" {
		sm.config.LogLevel = "info"
	}
	if sm.config.MaxSessions == 0 {
		sm.config.MaxSessions = 100
	}
	if sm.config.UploadBatch == 0 {
		sm.config.UploadBatch = 5
	}
	
	return nil
}

// setDefaultConfig sets default configuration values
func (sm *sessionManager) setDefaultConfig() {
	sm.config.ServerURL = "https://contextkeeper.dev"
	sm.config.LocalPort = 8080
	sm.config.LogLevel = "info"
	sm.config.EnableTLS = true
	sm.config.MaxSessions = 100
	sm.config.UploadBatch = 5
}

// saveConfig saves the current config to disk
func (sm *sessionManager) saveConfig() error {
	data, err := yaml.Marshal(sm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(sm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}