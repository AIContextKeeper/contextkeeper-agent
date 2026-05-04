package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AIContextKeeper/contextkeeper-agent/pkg/types"
)

// parser implements Parser interface for parsing AI tool outputs
type parser struct {
	claudeParser  *claudeParser
	geminiParser  *geminiParser
	genericParser *genericParser
}

// NewParser creates a new AI output parser
func NewParser() (*parser, error) {
	return &parser{
		claudeParser:  newClaudeParser(),
		geminiParser:  newGeminiParser(),
		genericParser: newGenericParser(),
	}, nil
}

// Parse parses an AI output into a session
func (p *parser) Parse(output *types.AIOutput) (*types.Session, error) {
	switch strings.ToLower(output.Tool) {
	case "claude":
		return p.claudeParser.parse(output)
	case "gemini":
		return p.geminiParser.parse(output)
	default:
		return p.genericParser.parse(output)
	}
}

// CanParse checks if the parser can handle a specific tool
func (p *parser) CanParse(tool string) bool {
	supportedTools := []string{"claude", "gemini", "chatgpt", "aider", "cursor", "copilot"}
	tool = strings.ToLower(tool)
	
	for _, supported := range supportedTools {
		if tool == supported {
			return true
		}
	}
	
	return true // Generic parser can handle any tool
}

// claudeParser handles Claude-specific parsing
type claudeParser struct {
	patterns map[string]*regexp.Regexp
}

func newClaudeParser() *claudeParser {
	return &claudeParser{
		patterns: map[string]*regexp.Regexp{
			"task":     regexp.MustCompile(`(?i)task:?\s*(.+)`),
			"question": regexp.MustCompile(`(?i)question:?\s*(.+)`),
			"request":  regexp.MustCompile(`(?i)request:?\s*(.+)`),
			"help":     regexp.MustCompile(`(?i)help.*with\s*(.+)`),
		},
	}
}

func (cp *claudeParser) parse(output *types.AIOutput) (*types.Session, error) {
	title := cp.extractTitle(output.Content)
	if title == "" {
		title = "Claude Session"
	}
	
	// Extract project name from path
	project := cp.extractProject(output.ProjectPath)
	
	// Categorize the session
	category := cp.categorizeContent(output.Content)
	
	// Extract technologies
	technologies := cp.extractTechnologies(output.Content)
	
	session := &types.Session{
		ID:            generateSessionID(),
		Title:         title,
		Project:       project,
		Content:       output.Content,
		Category:      category,
		Priority:      "medium",
		CreatedAt:     output.Timestamp,
		Source:        getSource(output.Metadata),
		Tool:          "claude",
		ProjectPath:   output.ProjectPath,
		Technologies:  technologies,
	}
	
	return session, nil
}

func (cp *claudeParser) extractTitle(content string) string {
	// Try to extract meaningful title from content
	for _, pattern := range cp.patterns {
		if matches := pattern.FindStringSubmatch(content); len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if len(title) > 100 {
				title = title[:100] + "..."
			}
			return title
		}
	}
	
	// Fallback: use first line or first 50 characters
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && len(lines[0]) > 0 {
		title := strings.TrimSpace(lines[0])
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		return title
	}
	
	return ""
}

func (cp *claudeParser) extractProject(projectPath string) string {
	if projectPath == "" {
		return "Unknown"
	}
	
	// Extract last directory name as project name
	parts := strings.Split(projectPath, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	
	return "Unknown"
}

func (cp *claudeParser) categorizeContent(content string) string {
	content = strings.ToLower(content)
	
	if strings.Contains(content, "bug") || strings.Contains(content, "error") || strings.Contains(content, "fix") {
		return "bug_fix"
	}
	if strings.Contains(content, "feature") || strings.Contains(content, "add") || strings.Contains(content, "implement") {
		return "feature"
	}
	if strings.Contains(content, "refactor") || strings.Contains(content, "optimize") || strings.Contains(content, "improve") {
		return "refactor"
	}
	if strings.Contains(content, "test") {
		return "testing"
	}
	if strings.Contains(content, "doc") || strings.Contains(content, "readme") {
		return "documentation"
	}
	
	return "other"
}

func (cp *claudeParser) extractTechnologies(content string) []string {
	technologies := []string{}
	content = strings.ToLower(content)
	
	techPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`\b(go|golang)\b`),
		"javascript": regexp.MustCompile(`\b(javascript|js|node\.?js)\b`),
		"typescript": regexp.MustCompile(`\b(typescript|ts)\b`),
		"python":     regexp.MustCompile(`\b(python|py)\b`),
		"react":      regexp.MustCompile(`\breact\b`),
		"nextjs":     regexp.MustCompile(`\b(next\.?js|nextjs)\b`),
		"docker":     regexp.MustCompile(`\bdocker\b`),
		"kubernetes": regexp.MustCompile(`\b(kubernetes|k8s)\b`),
		"aws":        regexp.MustCompile(`\b(aws|amazon)\b`),
		"git":        regexp.MustCompile(`\bgit\b`),
	}
	
	for tech, pattern := range techPatterns {
		if pattern.MatchString(content) {
			technologies = append(technologies, tech)
		}
	}
	
	return technologies
}

// geminiParser handles Gemini-specific parsing
type geminiParser struct{}

func newGeminiParser() *geminiParser {
	return &geminiParser{}
}

func (gp *geminiParser) parse(output *types.AIOutput) (*types.Session, error) {
	// For now, use similar logic to Claude parser
	// This can be specialized for Gemini-specific patterns later
	cp := newClaudeParser()
	session, err := cp.parse(output)
	if err != nil {
		return nil, err
	}
	
	session.Tool = "gemini"
	if session.Title == "Claude Session" {
		session.Title = "Gemini Session"
	}
	
	return session, nil
}

// genericParser handles generic AI tool outputs
type genericParser struct{}

func newGenericParser() *genericParser {
	return &genericParser{}
}

func (gp *genericParser) parse(output *types.AIOutput) (*types.Session, error) {
	// Basic parsing for unknown tools
	title := "AI Session"
	if len(output.Content) > 0 {
		lines := strings.Split(output.Content, "\n")
		if len(lines) > 0 && len(lines[0]) > 0 {
			title = strings.TrimSpace(lines[0])
			if len(title) > 50 {
				title = title[:50] + "..."
			}
		}
	}
	
	project := "Unknown"
	if output.ProjectPath != "" {
		parts := strings.Split(output.ProjectPath, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				project = parts[i]
				break
			}
		}
	}
	
	session := &types.Session{
		ID:          generateSessionID(),
		Title:       title,
		Project:     project,
		Content:     output.Content,
		Category:    "other",
		Priority:    "low",
		CreatedAt:   output.Timestamp,
		Source:      getSource(output.Metadata),
		Tool:        output.Tool,
		ProjectPath: output.ProjectPath,
	}
	
	return session, nil
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func getSource(metadata map[string]string) string {
	if source, ok := metadata["source"]; ok {
		return source
	}
	return "cli"
}