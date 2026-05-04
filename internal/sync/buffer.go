package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AIContextKeeper/contextkeeper-agent/pkg/types"
)

// buffer implements Buffer interface for local session buffering
type buffer struct {
	sessions    []*types.Session
	maxSize     int
	syncClient  SyncClient
	batchSize   int
	mu          sync.RWMutex
}

// SyncClient interface (minimal definition for this package)
type SyncClient interface {
	SaveSession(session *types.Session) error
	TestConnection() error
}

// NewBuffer creates a new session buffer
func NewBuffer(config *types.Config, syncClient SyncClient) (*buffer, error) {
	return &buffer{
		sessions:   make([]*types.Session, 0),
		maxSize:    config.MaxSessions,
		syncClient: syncClient,
		batchSize:  config.UploadBatch,
	}, nil
}

// Add adds a session to the buffer
func (b *buffer) Add(session *types.Session) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// Check if buffer is full
	if len(b.sessions) >= b.maxSize {
		// Try to flush some sessions to make room
		if err := b.flushOldest(b.batchSize); err != nil {
			return fmt.Errorf("buffer full and flush failed: %w", err)
		}
	}
	
	// Add session to buffer
	b.sessions = append(b.sessions, session)
	log.Printf("Added session to buffer: %s (buffer size: %d)", session.Title, len(b.sessions))
	
	return nil
}

// Flush uploads all buffered sessions to ContextKeeper.dev
func (b *buffer) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if len(b.sessions) == 0 {
		return nil
	}
	
	log.Printf("Flushing %d sessions to ContextKeeper.dev", len(b.sessions))
	
	// Test connection first
	if err := b.syncClient.TestConnection(); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	// Upload sessions in batches
	uploaded := 0
	failed := make([]*types.Session, 0)
	
	for i := 0; i < len(b.sessions); i += b.batchSize {
		end := i + b.batchSize
		if end > len(b.sessions) {
			end = len(b.sessions)
		}
		
		batch := b.sessions[i:end]
		
		for _, session := range batch {
			select {
			case <-ctx.Done():
				// Context cancelled, save remaining sessions for later
				failed = append(failed, b.sessions[i+uploaded:]...)
				b.sessions = failed
				return ctx.Err()
			default:
				if err := b.uploadSession(session); err != nil {
					log.Printf("Failed to upload session %s: %v", session.Title, err)
					failed = append(failed, session)
				} else {
					uploaded++
				}
			}
		}
		
		// Small delay between batches to avoid overwhelming the server
		if i+b.batchSize < len(b.sessions) {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// Keep only failed sessions in buffer
	b.sessions = failed
	
	if uploaded > 0 {
		log.Printf("Successfully uploaded %d sessions", uploaded)
	}
	
	if len(failed) > 0 {
		log.Printf("Failed to upload %d sessions, keeping in buffer", len(failed))
		return fmt.Errorf("failed to upload %d out of %d sessions", len(failed), uploaded+len(failed))
	}
	
	return nil
}

// Size returns the number of sessions in the buffer
func (b *buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.sessions)
}

// flushOldest flushes the oldest sessions from the buffer
func (b *buffer) flushOldest(count int) error {
	if len(b.sessions) == 0 {
		return nil
	}
	
	if count > len(b.sessions) {
		count = len(b.sessions)
	}
	
	// Upload oldest sessions
	uploaded := 0
	for i := 0; i < count; i++ {
		session := b.sessions[i]
		if err := b.uploadSession(session); err != nil {
			log.Printf("Failed to upload oldest session %s: %v", session.Title, err)
			break
		}
		uploaded++
	}
	
	// Remove uploaded sessions from buffer
	if uploaded > 0 {
		b.sessions = b.sessions[uploaded:]
		log.Printf("Flushed %d oldest sessions from buffer", uploaded)
	}
	
	if uploaded == 0 {
		return fmt.Errorf("failed to flush any sessions")
	}
	
	return nil
}

// uploadSession uploads a single session with error handling
func (b *buffer) uploadSession(session *types.Session) error {
	return b.syncClient.SaveSession(session)
}

// PermanentError represents an error that shouldn't be retried
type PermanentError struct {
	Message string
	Err     error
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// IsPermanentError checks if an error is permanent
func IsPermanentError(err error) bool {
	_, ok := err.(*PermanentError)
	return ok
}

// GetSessions returns a copy of all buffered sessions (for debugging/monitoring)
func (b *buffer) GetSessions() []*types.Session {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	sessions := make([]*types.Session, len(b.sessions))
	copy(sessions, b.sessions)
	return sessions
}

// Clear removes all sessions from the buffer
func (b *buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	count := len(b.sessions)
	b.sessions = b.sessions[:0]
	
	if count > 0 {
		log.Printf("Cleared %d sessions from buffer", count)
	}
}