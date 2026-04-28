package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/randodev95/eventcanvas/internal/tui"
	"github.com/randodev95/eventcanvas/pkg/ast"
	"github.com/randodev95/eventcanvas/pkg/normalization"
	"github.com/randodev95/eventcanvas/pkg/validator"
)

type Server struct {
	Plan    *ast.TrackingPlan
	Updates chan tui.EventMsg
	clients map[chan tui.EventMsg]bool
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

	// 1. Normalize
	normalized, err := normalization.Normalize(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Resolve Schema
	schema, err := s.Plan.ResolveEventSchema(normalized.Event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// 3. Validate
	result, err := validator.Validate(normalized, schema)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg := tui.EventMsg{
		Name:    normalized.Event,
		IsValid: result.Valid,
		Errors:  result.Errors,
	}

	if s.Updates != nil {
		s.Updates <- msg
	}

	for client := range s.clients {
		client <- msg
	}

	if !result.Valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
