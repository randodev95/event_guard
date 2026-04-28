package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/randodev95/event_guard/pkg/ast"
)

func TestHandleEvent_Integration(t *testing.T) {
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "number", Required: true},
				},
			},
		},
	}

	srv := &Server{
		Plan:    plan,
		Updates: nil,
	}

	t.Run("Valid Event", func(t *testing.T) {
		payload := []byte(`{"event": "Order Completed", "userId": "user1", "properties": {"total": 100.50}}`)
		req := httptest.NewRequest("POST", "/track", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		srv.HandleEvent(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("Invalid Event (Missing Property)", func(t *testing.T) {
		payload := []byte(`{"event": "Order Completed", "userId": "user1", "properties": {}}`)
		req := httptest.NewRequest("POST", "/track", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		srv.HandleEvent(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %d", rr.Code)
		}
	})

	t.Run("Unknown Event", func(t *testing.T) {
		payload := []byte(`{"event": "Unknown Event", "userId": "user1"}`)
		req := httptest.NewRequest("POST", "/track", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		srv.HandleEvent(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %d", rr.Code)
		}
	})
}
