package agent

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/carsor007/contextkeeper-agent/pkg/types"
)

// monitor implements Monitor interface for detecting AI tool outputs
type monitor struct {
	config      *types.Config
	subscribers []chan<- *types.AIOutput
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewMonitor creates a new terminal monitor
func NewMonitor(config *types.Config) (Monitor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &monitor{
		config:      config,
		subscribers: make([]chan<- *types.AIOutput, 0),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start begins monitoring for AI tool outputs
func (m *monitor) Start(ctx context.Context) error {
	log.Printf("Starting terminal monitor...")
	
	// Start monitoring different sources
	m.wg.Add(1)
	go m.monitorClaudeCode()
	
	m.wg.Add(1)
	go m.monitorShellHistory()
	
	m.wg.Add(1)
	go m.monitorProcesses()
	
	return nil
}

// Stop stops the monitor
func (m *monitor) Stop() error {
	log.Printf("Stopping terminal monitor...")
	m.cancel()
	m.wg.Wait()
	return nil
}

// Subscribe adds a channel to receive AI outputs
func (m *monitor) Subscribe(ch chan<- *types.AIOutput) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, ch)
}

// publish sends an AI output to all subscribers
func (m *monitor) publish(output *types.AIOutput) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, ch := range m.subscribers {
		select {
		case ch <- output:
		default:
			// Channel is full, skip this subscriber
			log.Printf("Warning: subscriber channel full, dropping output")
		}
	}
}

// monitorClaudeCode monitors for Claude Code (claude-ai CLI) outputs
func (m *monitor) monitorClaudeCode() {
	defer m.wg.Done()
	
	// This is a simplified implementation
	// In a real implementation, we'd hook into the Claude CLI or monitor its output
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Check for Claude Code sessions
			if output := m.detectClaudeCodeOutput(); output != nil {
				m.publish(output)
			}
		}
	}
}

// monitorShellHistory monitors shell history for AI tool usage
func (m *monitor) monitorShellHistory() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkShellHistory()
		}
	}
}

// monitorProcesses monitors running processes for AI tools
func (m *monitor) monitorProcesses() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkRunningProcesses()
		}
	}
}

// detectClaudeCodeOutput detects Claude Code CLI outputs
func (m *monitor) detectClaudeCodeOutput() *types.AIOutput {
	// Look for Claude Code temporary files or recent outputs
	// This is a placeholder implementation
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	
	// Check for Claude Code session files
	claudeDir := fmt.Sprintf("%s/.claude", homeDir)
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return nil
	}
	
	// This would be more sophisticated in a real implementation
	// For now, return nil as we need to integrate with actual Claude Code
	return nil
}

// checkShellHistory checks shell history for AI tool commands
func (m *monitor) checkShellHistory() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	// Check common shell history files
	historyFiles := []string{
		fmt.Sprintf("%s/.bash_history", homeDir),
		fmt.Sprintf("%s/.zsh_history", homeDir),
		fmt.Sprintf("%s/.fish_history", homeDir),
	}
	
	for _, histFile := range historyFiles {
		if err := m.parseHistoryFile(histFile); err != nil {
			// Silently continue - history file might not exist
			continue
		}
	}
}

// parseHistoryFile parses a shell history file for AI tool usage
func (m *monitor) parseHistoryFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// AI tool patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`claude.*`),
		regexp.MustCompile(`aider.*`),
		regexp.MustCompile(`cursor.*`),
		regexp.MustCompile(`copilot.*`),
		regexp.MustCompile(`chatgpt.*`),
		regexp.MustCompile(`gemini.*`),
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				// Found AI tool usage
				output := &types.AIOutput{
					Tool:        m.extractToolName(line),
					Content:     line,
					Metadata:    map[string]string{"source": "shell_history"},
					Timestamp:   time.Now(),
					ProjectPath: m.getCurrentDirectory(),
				}
				
				m.publish(output)
				break
			}
		}
	}
	
	return scanner.Err()
}

// checkRunningProcesses checks for running AI tool processes
func (m *monitor) checkRunningProcesses() {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if m.isAIToolProcess(line) {
			toolOutput := &types.AIOutput{
				Tool:        m.extractToolFromProcess(line),
				Content:     line,
				Metadata:    map[string]string{"source": "process_monitor"},
				Timestamp:   time.Now(),
				ProjectPath: m.getCurrentDirectory(),
			}
			
			m.publish(toolOutput)
		}
	}
}

// isAIToolProcess checks if a process line represents an AI tool
func (m *monitor) isAIToolProcess(processLine string) bool {
	aiTools := []string{
		"claude",
		"aider",
		"cursor",
		"copilot",
		"chatgpt",
		"gemini",
		"anthropic",
	}
	
	lowerLine := strings.ToLower(processLine)
	for _, tool := range aiTools {
		if strings.Contains(lowerLine, tool) {
			return true
		}
	}
	
	return false
}

// extractToolName extracts the AI tool name from a command
func (m *monitor) extractToolName(command string) string {
	command = strings.ToLower(command)
	
	if strings.Contains(command, "claude") {
		return "claude"
	}
	if strings.Contains(command, "aider") {
		return "aider"
	}
	if strings.Contains(command, "cursor") {
		return "cursor"
	}
	if strings.Contains(command, "copilot") {
		return "copilot"
	}
	if strings.Contains(command, "chatgpt") {
		return "chatgpt"
	}
	if strings.Contains(command, "gemini") {
		return "gemini"
	}
	
	return "unknown"
}

// extractToolFromProcess extracts tool name from process line
func (m *monitor) extractToolFromProcess(processLine string) string {
	return m.extractToolName(processLine)
}

// getCurrentDirectory gets the current working directory
func (m *monitor) getCurrentDirectory() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return ""
}