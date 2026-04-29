package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/randodev95/event_guard/internal/tui"
	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/validator"
	"sync"
)

// Server represents the development mock server that receives and validates live events.
type Server struct {
	mu      sync.RWMutex
	Plan    *ast.TrackingPlan
	Engine  *validator.Engine
	DLQ     chan []byte // Dead Letter Queue for failed events
	Updates chan tui.EventMsg
	clients map[chan tui.EventMsg]bool
}

// UpdatePlan reloads the validation engine with a new tracking plan.
func (s *Server) UpdatePlan(plan *ast.TrackingPlan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Plan = plan
	s.Engine = validator.NewEngine(plan)
	s.Engine.Warmup()
}

// HandleEvents streams live validation events to clients via SSE.
func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan tui.EventMsg)
	if s.clients == nil {
		s.clients = make(map[chan tui.EventMsg]bool)
	}
	s.clients[clientChan] = true
	defer func() {
		delete(s.clients, clientChan)
		close(clientChan)
	}()

	for msg := range clientChan {
		data, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// HandlePlan returns the current tracking plan as JSON.
func (s *Server) HandlePlan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Plan)
}

// HandleEvent receives a single event, validates it, and streams the result.
func (s *Server) HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.mu.RLock()
	engine := s.Engine
	s.mu.RUnlock()

	if engine == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	result, err := engine.ValidateJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	msg := tui.EventMsg{
		Name:    "unknown", // Will be refined if valid
		IsValid: result.Valid,
		Errors:  result.Errors,
	}

	// For TUI/SSE, we still want the event name if possible
	if norm, err := engine.GetMapper().Map(body); err == nil {
		msg.Name = norm.Event
	}

	if s.Updates != nil {
		s.Updates <- msg
	}

	s.mu.RLock()
	clients := s.clients
	s.mu.RUnlock()
	for client := range clients {
		client <- msg
	}

	if !result.Valid {
		if s.DLQ != nil {
			s.DLQ <- body
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(result)
		return
	}

	w.WriteHeader(http.StatusOK)
}
