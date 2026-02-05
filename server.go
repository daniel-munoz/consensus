package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/daniel-munoz/consensus/output"
	"github.com/google/uuid"
)

//go:embed ui
var uiFiles embed.FS

// ConsensusRequest represents the JSON request body for creating a consensus session
type ConsensusRequest struct {
	Prompt            string   `json:"prompt"`
	UseMasterPrompt   bool     `json:"useMasterPrompt"`
	MasterProvider    string   `json:"masterProvider"`
	ResponseProviders []string `json:"responseProviders"`
	EmailRecipients   []string `json:"emailRecipients"`
}

// SessionResponse represents the JSON response for session status
type SessionResponse struct {
	SessionID    string                      `json:"sessionId"`
	Status       string                      `json:"status"` // pending, processing, complete, error
	MasterPrompt *ProviderResponse           `json:"masterPrompt,omitempty"`
	Responses    map[string]*ProviderResponse `json:"responses"`
	Error        string                      `json:"error,omitempty"`
}

// ProviderResponse represents the status and content from a single provider
type ProviderResponse struct {
	Status  string `json:"status"` // pending, complete, error
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Session holds the state of a consensus session
type Session struct {
	ID           string
	Status       string
	MasterPrompt *ProviderResponse
	Responses    map[string]*ProviderResponse
	Error        string
	CreatedAt    time.Time
	mu           sync.RWMutex
}

// SessionManager manages all active sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}
	// Start cleanup goroutine for expired sessions
	go sm.cleanupExpiredSessions()
	return sm
}

// CreateSession creates a new session and returns its ID
func (sm *SessionManager) CreateSession(providerNames []string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := uuid.NewString()
	responses := make(map[string]*ProviderResponse)
	for _, name := range providerNames {
		responses[name] = &ProviderResponse{Status: "pending"}
	}

	session := &Session{
		ID:        id,
		Status:    "pending",
		Responses: responses,
		CreatedAt: time.Now(),
	}
	sm.sessions[id] = session
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

// cleanupExpiredSessions removes sessions older than 30 minutes
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		sm.mu.Lock()
		for id, session := range sm.sessions {
			if time.Since(session.CreatedAt) > 30*time.Minute {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// Server holds the HTTP server and its dependencies
type Server struct {
	config         *Config
	sessionManager *SessionManager
	port           int
	httpServer     *http.Server
}

// NewServer creates a new HTTP server
func NewServer(config *Config, port int) *Server {
	return &Server{
		config:         config,
		sessionManager: NewSessionManager(),
		port:           port,
	}
}

// Start starts the HTTP server and opens the browser
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/consensus", s.handleConsensus)
	mux.HandleFunc("/api/session/", s.handleSession)
	mux.HandleFunc("/api/providers", s.handleProviders)
	mux.HandleFunc("/api/health", s.handleHealth)

	// Serve embedded UI files
	uiFS, _ := fs.Sub(uiFiles, "ui")
	fileServer := http.FileServer(http.FS(uiFS))
	mux.Handle("/", fileServer)

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	fmt.Printf("Starting Consensus AI server at http://localhost%s\n", addr)

	// Open browser after a short delay to allow server to start
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://localhost%s", addr))
	}()

	err := s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil // Graceful shutdown
	}
	return err
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	if s.httpServer != nil {
		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Printf("Open your browser to: %s\n", url)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open browser: %v\nOpen your browser to: %s\n", err, url)
	}
}

// handleConsensus handles POST /api/consensus - creates a new consensus session
func (s *Server) handleConsensus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConsensusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Prompt == "" {
		s.jsonError(w, "Prompt is required", http.StatusBadRequest)
		return
	}
	if len(req.ResponseProviders) == 0 {
		s.jsonError(w, "At least one response provider is required", http.StatusBadRequest)
		return
	}

	// Validate providers exist
	for _, providerName := range req.ResponseProviders {
		if _, exists := s.config.AllProviders[providerName]; !exists {
			s.jsonError(w, fmt.Sprintf("Unknown provider: %s", providerName), http.StatusBadRequest)
			return
		}
	}
	if req.UseMasterPrompt && req.MasterProvider != "" {
		if _, exists := s.config.AllProviders[req.MasterProvider]; !exists {
			s.jsonError(w, fmt.Sprintf("Unknown master provider: %s", req.MasterProvider), http.StatusBadRequest)
			return
		}
	}

	// Create session
	session := s.sessionManager.CreateSession(req.ResponseProviders)
	if req.UseMasterPrompt {
		session.MasterPrompt = &ProviderResponse{Status: "pending"}
	}

	// Start processing in background
	go s.processSession(session, req)

	// Return session ID immediately
	s.jsonResponse(w, map[string]string{"sessionId": session.ID})
}

// handleSession handles GET /api/session/{id} - returns session status
func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from path
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/session/")
	if sessionID == "" {
		s.jsonError(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	session := s.sessionManager.GetSession(sessionID)
	if session == nil {
		s.jsonError(w, "Session not found", http.StatusNotFound)
		return
	}

	session.mu.RLock()
	response := SessionResponse{
		SessionID:    session.ID,
		Status:       session.Status,
		MasterPrompt: session.MasterPrompt,
		Responses:    session.Responses,
		Error:        session.Error,
	}
	session.mu.RUnlock()

	s.jsonResponse(w, response)
}

// handleProviders handles GET /api/providers - returns available providers
func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providers := make([]string, 0, len(s.config.AllProviders))
	for name := range s.config.AllProviders {
		providers = append(providers, name)
	}

	s.jsonResponse(w, map[string][]string{"providers": providers})
}

// handleHealth handles GET /api/health - health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// processSession processes a consensus request in the background
func (s *Server) processSession(session *Session, req ConsensusRequest) {
	session.mu.Lock()
	session.Status = "processing"
	session.mu.Unlock()

	// Create output manager for file and email outputs
	emailTo := strings.Join(req.EmailRecipients, ",")
	outputManager := NewOutputManager(s.config, emailTo)

	// Save the original request
	outputManager.Send(req.Prompt, session.ID, "request", "request")

	prompt := req.Prompt

	// Run master prompt optimization if enabled
	if req.UseMasterPrompt {
		masterProviderName := req.MasterProvider
		if masterProviderName == "" {
			masterProviderName = s.config.PromptProvider.Name()
		}

		masterProvider, exists := s.config.AllProviders[masterProviderName]
		if !exists {
			session.mu.Lock()
			session.MasterPrompt.Status = "error"
			session.MasterPrompt.Error = fmt.Sprintf("Master provider not found: %s", masterProviderName)
			session.Status = "error"
			session.Error = session.MasterPrompt.Error
			session.mu.Unlock()
			return
		}

		optimizedPrompt, err := OptimizePrompt(req.Prompt, masterProvider, session.ID)
		if err != nil {
			session.mu.Lock()
			session.MasterPrompt.Status = "error"
			session.MasterPrompt.Error = err.Error()
			session.mu.Unlock()
			// Continue with original prompt on master prompt failure
			prompt = req.Prompt
		} else {
			session.mu.Lock()
			session.MasterPrompt.Status = "complete"
			session.MasterPrompt.Content = optimizedPrompt
			session.mu.Unlock()
			prompt = optimizedPrompt
			outputManager.Send(prompt, session.ID, "prompt", "prompt")
		}
	}

	// Send to all providers concurrently
	var wg sync.WaitGroup
	for _, providerName := range req.ResponseProviders {
		wg.Add(1)
		go func(pName string) {
			defer wg.Done()
			s.processProvider(session, pName, prompt, outputManager)
		}(providerName)
	}

	wg.Wait()

	// Mark session as complete
	session.mu.Lock()
	session.Status = "complete"
	session.mu.Unlock()

	// Give the UI time to poll for the final status, then shut down
	go func() {
		time.Sleep(3 * time.Second)
		s.Shutdown()
	}()
}

// processProvider sends a prompt to a single provider and updates the session
func (s *Server) processProvider(session *Session, providerName, prompt string, outputManager *output.Manager) {
	provider, exists := s.config.AllProviders[providerName]
	if !exists {
		session.mu.Lock()
		session.Responses[providerName].Status = "error"
		session.Responses[providerName].Error = "Provider not found"
		session.mu.Unlock()
		return
	}

	fmt.Printf("Consulting %s...\n", providerName)
	response, err := provider.Send(prompt, nil)
	if err != nil {
		session.mu.Lock()
		session.Responses[providerName].Status = "error"
		session.Responses[providerName].Error = err.Error()
		session.mu.Unlock()
		fmt.Printf("Error from %s: %v\n", providerName, err)
		return
	}

	session.mu.Lock()
	session.Responses[providerName].Status = "complete"
	session.Responses[providerName].Content = response
	session.mu.Unlock()

	// Save to file and send email
	outputManager.Send(response, session.ID, provider.Name(), provider.Type())
	fmt.Printf("%s responded!\n", providerName)
}

// jsonResponse writes a JSON response
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func (s *Server) jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

