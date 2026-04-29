package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/randodev95/event_guard/pkg/validator"
)

// Sink defines a pluggable destination for rejected events (DLQ).
type Sink interface {
	Push(payload []byte, errors []string) error
	Close() error
}

// Server implements the HTTP transport layer for EventGuard.
type Server struct {
	engine atomic.Pointer[validator.Engine]
	mux    *http.ServeMux
	logger *slog.Logger
	sink   Sink
	onReload func() error
}

// NewServer initializes a new API server with the provided validation engine.
func NewServer(engine *validator.Engine) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		logger: slog.Default(),
	}
	s.engine.Store(engine)
	s.routes()
	return s
}

// SetSink configures the DLQ sink.
func (s *Server) SetSink(sink Sink) {
	s.sink = sink
}

// SetReloadHandler sets the callback for administrative reloads.
func (s *Server) SetReloadHandler(h func() error) {
	s.onReload = h
}

func (s *Server) UpdateEngine(engine *validator.Engine) {
	old := s.engine.Swap(engine)
	if old != nil {
		old.ResetCache()
	}
	s.logger.Info("validation engine reloaded")
}

func (s *Server) routes() {
	s.mux.HandleFunc("/validate", s.handleValidate())
	s.mux.HandleFunc("/admin/reload", s.handleAdminReload())
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func (s *Server) handleAdminReload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST", http.StatusMethodNotAllowed)
			return
		}
		
		if s.onReload != nil {
			if err := s.onReload(); err != nil {
				s.logger.Error("reload failed", "err", err)
				http.Error(w, "Reload failed", http.StatusInternalServerError)
				return
			}
		}
		w.Write([]byte("reload successful"))
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleValidate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Error("failed to read request body", "err", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		engine := s.engine.Load()
		if engine == nil {
			http.Error(w, "Engine not initialized", http.StatusInternalServerError)
			return
		}

		result, err := engine.ValidateJSON(body)
		if err != nil {
			s.logger.Error("validation system error", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !result.Valid {
			s.logger.Warn("event rejected", "errors", result.Errors)
			if s.sink != nil {
				if err := s.sink.Push(body, result.Errors); err != nil {
					s.logger.Error("sink push failed", "err", err)
				}
			}
		} else {
			s.logger.Info("event validated")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
