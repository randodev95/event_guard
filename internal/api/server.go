package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/eventcanvas/eventcanvas/pkg/validator"
)

// Server implements the HTTP transport layer for EventCanvas.
type Server struct {
	engine *validator.Engine
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewServer(engine *validator.Engine) *Server {
	s := &Server{
		engine: engine,
		mux:    http.NewServeMux(),
		logger: slog.Default(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/validate", s.handleValidate())
	// Senior Pattern: Standard health check for Kubernetes/Load Balancers
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

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

		result, err := s.engine.ValidateJSON(body)
		if err != nil {
			s.logger.Error("validation system error", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.logger.Info("event validated", 
			"valid", result.Valid, 
			"errors_count", len(result.Errors),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
