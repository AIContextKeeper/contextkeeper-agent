package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AIContextKeeper/contextkeeper-agent/internal/parser"
	agentsync "github.com/AIContextKeeper/contextkeeper-agent/internal/sync"
	"github.com/AIContextKeeper/contextkeeper-agent/pkg/types"
)

// Agent represents the main ContextKeeper agent
type Agent struct {
	config        *types.Config
	sessionMgr    SessionManager
	monitor       Monitor
	syncClient    SyncClient
	parser        Parser
	buffer        Buffer
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	running       bool
	mu            sync.RWMutex
}

// SessionManager interface for managing user sessions
type SessionManager interface {
	GetOrCreateSession() string
	GetUsageInfo() (*types.UsageInfo, error)
	IsAnonymous() bool
	SetAPIKey(apiKey string) error
}

// Monitor interface for monitoring terminal/CLI activity
type Monitor interface {
	Start(ctx context.Context) error
	Stop() error
	Subscribe(ch chan<- *types.AIOutput)
}

// SyncClient interface for syncing with ContextKeeper.dev
type SyncClient interface {
	SaveSession(session *types.Session) error
	GetUsageInfo(sessionID string) (*types.UsageInfo, error)
	TestConnection() error
}

// Parser interface for parsing AI tool outputs
type Parser interface {
	Parse(output *types.AIOutput) (*types.Session, error)
	CanParse(tool string) bool
}

// Buffer interface for local buffering and batch uploads
type Buffer interface {
	Add(session *types.Session) error
	Flush(ctx context.Context) error
	Size() int
}

// New creates a new ContextKeeper agent
func New(config *types.Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	a := &Agent{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := a.initializeComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return a, nil
}

// Start begins the agent's monitoring and processing
func (a *Agent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.running {
		return fmt.Errorf("agent is already running")
	}
	
	log.Printf("Starting ContextKeeper agent...")

	// Start API server for VS Code integration
	apiServer := NewAPIServer(a, a.config.LocalPort)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()
	
	// Start monitoring
	outputCh := make(chan *types.AIOutput, 100)
	a.monitor.Subscribe(outputCh)
	
	if err := a.monitor.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}
	
	// Start processing goroutine
	a.wg.Add(1)
	go a.processOutputs(outputCh)
	
	// Start periodic flush
	a.wg.Add(1)
	go a.periodicFlush()
	
	a.running = true
	log.Printf("ContextKeeper agent started successfully")
	
	return nil
}

// Stop gracefully stops the agent
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.running {
		return nil
	}
	
	log.Printf("Stopping ContextKeeper agent...")
	
	// Cancel context to stop all goroutines
	a.cancel()
	
	// Stop monitor
	if err := a.monitor.Stop(); err != nil {
		log.Printf("Error stopping monitor: %v", err)
	}
	
	// Wait for goroutines to finish
	a.wg.Wait()
	
	// Flush remaining sessions
	if err := a.buffer.Flush(context.Background()); err != nil {
		log.Printf("Error flushing buffer on shutdown: %v", err)
	}
	
	a.running = false
	log.Printf("ContextKeeper agent stopped")
	
	return nil
}

// IsRunning returns whether the agent is currently running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetUsageInfo returns current usage information
func (a *Agent) GetUsageInfo() (*types.UsageInfo, error) {
	return a.sessionMgr.GetUsageInfo()
}

// initializeComponents initializes all agent components
func (a *Agent) initializeComponents() error {
	// Initialize session manager
	sessionMgr, err := NewSessionManager(a.config)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	a.sessionMgr = sessionMgr
	
	// Initialize sync client
	syncClient, err := agentsync.NewClient(a.config, a.sessionMgr)
	if err != nil {
		return fmt.Errorf("failed to create sync client: %w", err)
	}
	a.syncClient = syncClient

	// Initialize parser
	p, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}
	a.parser = p

	// Initialize buffer
	buf, err := agentsync.NewBuffer(a.config, a.syncClient)
	if err != nil {
		return fmt.Errorf("failed to create buffer: %w", err)
	}
	a.buffer = buf
	
	// Initialize monitor
	monitor, err := NewMonitor(a.config)
	if err != nil {
		return fmt.Errorf("failed to create monitor: %w", err)
	}
	a.monitor = monitor
	
	return nil
}

// processOutputs processes AI outputs from the monitor
func (a *Agent) processOutputs(outputCh <-chan *types.AIOutput) {
	defer a.wg.Done()
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case output := <-outputCh:
			if err := a.processOutput(output); err != nil {
				log.Printf("Error processing output: %v", err)
			}
		}
	}
}

// processOutput processes a single AI output
func (a *Agent) processOutput(output *types.AIOutput) error {
	// Parse the output into a session
	session, err := a.parser.Parse(output)
	if err != nil {
		return fmt.Errorf("failed to parse output: %w", err)
	}
	
	// Check if user has premium features for auto-sync
	if a.sessionMgr.IsAnonymous() {
		// Free users: local storage only, show upgrade hint
		log.Printf("📝 Captured %s session: %s", session.Tool, session.Title)
		log.Printf("💡 Want automatic sync? Upgrade to Pro ($29/month) at contextkeeper.dev/pricing")
		// TODO: Store locally for manual export
		return nil
	}
	
	// Paid users: auto-sync to cloud
	if err := a.buffer.Add(session); err != nil {
		return fmt.Errorf("failed to add session to buffer: %w", err)
	}
	
	log.Printf("✅ Processed %s session: %s", session.Tool, session.Title)
	return nil
}

// periodicFlush periodically flushes the buffer
func (a *Agent) periodicFlush() {
	defer a.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // Flush every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if a.buffer.Size() > 0 {
				if err := a.buffer.Flush(a.ctx); err != nil {
					log.Printf("Error during periodic flush: %v", err)
				}
			}
		}
	}
}