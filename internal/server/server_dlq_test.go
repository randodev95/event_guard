package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/validator"
)

func TestServer_DLQ(t *testing.T) {
	plan := &ast.TrackingPlan{
		IdentityProperties: []string{"userId"},
		Events: map[string]ast.Event{
			"Staked": {Category: "A", EntityType: "T", Properties: map[string]ast.Property{
				"userId": {Type: "string", Required: true},
			}},
		},
	}
	dlq := make(chan []byte, 1)
	s := &Server{
		Plan:   plan,
		Engine: validator.NewEngine(plan),
		DLQ:    dlq,
	}

	// 1. Send invalid event (missing userId)
	payload := []byte(`{"event": "Staked", "amount": 10}`)
	req := httptest.NewRequest("POST", "/event", bytes.NewBuffer(payload))
	resp := httptest.NewRecorder()
	s.HandleEvent(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected BadRequest, got %d", resp.Code)
	}

	// 2. Verify DLQ received payload
	select {
	case p := <-dlq:
		if !bytes.Equal(p, payload) {
			t.Errorf("DLQ payload mismatch")
		}
	default:
		t.Error("DLQ did not receive event")
	}
}
