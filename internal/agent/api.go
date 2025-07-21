package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/carsor007/contextkeeper-agent/pkg/types"
)

// APIServer provides HTTP API for VS Code extension integration
type APIServer struct {
	agent *Agent
	port  int
}

// NewAPIServer creates a new API server
func NewAPIServer(agent *Agent, port int) *APIServer {
	return &APIServer{
		agent: agent,
		port:  port,
	}
}

// Start starts the HTTP API server
func (s *APIServer) Start() error {
	mux := http.NewServeMux()
	
	// CORS middleware for VS Code extension
	mux.HandleFunc("/", s.corsMiddleware(s.handleRoot))
	mux.HandleFunc("/session", s.corsMiddleware(s.handleSession))
	mux.HandleFunc("/sessions", s.corsMiddleware(s.handleSessions))
	mux.HandleFunc("/health", s.corsMiddleware(s.handleHealth))
	mux.HandleFunc("/usage", s.corsMiddleware(s.handleUsage))
	
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting API server on %s", addr)
	
	return http.ListenAndServe(addr, mux)
}

// corsMiddleware adds CORS headers for VS Code extension
func (s *APIServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-session-id, x-source")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// handleRoot provides API information
func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":     "ContextKeeper Agent",
		"version":     "1.0.0",
		"endpoints": map[string]string{
			"GET /session":    "Get current session ID",
			"POST /sessions":  "Submit session data",
			"GET /health":     "Health check",
			"GET /usage":      "Usage information",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSession returns the current session ID
func (s *APIServer) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	sessionID := s.agent.sessionMgr.GetOrCreateSession()
	
	response := map[string]string{
		"session_id": sessionID,
		"source":     "agent",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSessions accepts session data from VS Code extension
func (s *APIServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var sessionData types.Session
	if err := json.NewDecoder(r.Body).Decode(&sessionData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Set the agent's session ID
	sessionData.UserSessionID = s.agent.sessionMgr.GetOrCreateSession()
	sessionData.Source = "vscode"
	
	// Process through the agent's pipeline
	if s.agent.sessionMgr.IsAnonymous() {
		// Free users: local storage only
		log.Printf("📝 Received VS Code session: %s (free user)", sessionData.Title)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "stored_locally",
			"message": "Upgrade to Pro for automatic sync",
		})
		return
	}
	
	// Paid users: sync to cloud
	if err := s.agent.buffer.Add(&sessionData); err != nil {
		log.Printf("Error adding VS Code session to buffer: %v", err)
		http.Error(w, "Failed to process session", http.StatusInternalServerError)
		return
	}
	
	log.Printf("✅ Received VS Code session: %s", sessionData.Title)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"session_id": sessionData.UserSessionID,
	})
}

// handleHealth provides health check
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"status":  "healthy",
		"running": s.agent.IsRunning(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUsage provides usage information
func (s *APIServer) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	usage, err := s.agent.GetUsageInfo()
	if err != nil {
		http.Error(w, "Failed to get usage info", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}