package types

import "time"

// Session represents an AI interaction session that will be sent to ContextKeeper
type Session struct {
	ID              string    `json:"id"`
	UserSessionID   string    `json:"user_session_id"`
	UserID          *string   `json:"user_id,omitempty"`
	Title           string    `json:"title"`
	Project         string    `json:"project"`
	Content         string    `json:"content"`
	ParsedSummary   *string   `json:"parsed_summary,omitempty"`
	CodeChanges     []string  `json:"code_changes,omitempty"`
	SessionContext  *string   `json:"session_context,omitempty"`
	Todos           []string  `json:"todos,omitempty"`
	Technologies    []string  `json:"technologies,omitempty"`
	Category        string    `json:"category"`
	Priority        string    `json:"priority"`
	CreatedAt       time.Time `json:"created_at"`
	Source          string    `json:"source"` // "vscode", "cli", "manual"
	Tool            string    `json:"tool"`   // "claude", "gemini", "chatgpt"
	ProjectPath     string    `json:"project_path"`
}

// SessionIdentifier holds session identification info
type SessionIdentifier struct {
	SessionID   string `json:"session_id"` // For anonymous: "session_1642512000000_abc123def"
	UserID      string `json:"user_id"`    // For authenticated: UUID from Supabase
	IsAnonymous bool   `json:"is_anonymous"`
}

// UsageInfo tracks usage limits for anonymous sessions
type UsageInfo struct {
	Current    int  `json:"current"`
	Limit      int  `json:"limit"`      // 50 for anonymous, unlimited for pro
	Percentage int  `json:"percentage"`
	HasReached bool `json:"has_reached"`
}

// Config holds agent configuration
type Config struct {
	SessionID    string `yaml:"session_id"`
	APIKey       string `yaml:"api_key,omitempty"`
	ServerURL    string `yaml:"server_url"`
	LocalPort    int    `yaml:"local_port"`
	LogLevel     string `yaml:"log_level"`
	EnableTLS    bool   `yaml:"enable_tls"`
	MaxSessions  int    `yaml:"max_sessions"`
	UploadBatch  int    `yaml:"upload_batch"`
}

// AIOutput represents parsed AI tool output
type AIOutput struct {
	Tool        string            `json:"tool"`
	Content     string            `json:"content"`
	Metadata    map[string]string `json:"metadata"`
	Timestamp   time.Time         `json:"timestamp"`
	ProjectPath string            `json:"project_path"`
}