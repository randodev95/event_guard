package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/validator"
)

func TestServer_Reload(t *testing.T) {
	// 1. Initial Plan: Staked event required
	plan1 := &ast.TrackingPlan{
		IdentityProperties: []string{"userId"},
		Events: map[string]ast.Event{
			"Staked": {Category: "A", EntityType: "T", Properties: map[string]ast.Property{
				"userId": {Type: "string", Required: true},
			}},
		},
	}
	s := &Server{
		Plan:   plan1,
		Engine: validator.NewEngine(plan1),
	}

	// 2. Validate valid event
	payload := []byte(`{"event": "Staked", "userId": "u1"}`)
	req := httptest.NewRequest("POST", "/event", bytes.NewBuffer(payload))
	resp := httptest.NewRecorder()
	s.HandleEvent(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected OK, got %d", resp.Code)
	}

	// 3. Update Plan: Staked now requires amount
	plan2 := &ast.TrackingPlan{
		IdentityProperties: []string{"userId"},
		Events: map[string]ast.Event{
			"Staked": {Category: "A", EntityType: "T", Properties: map[string]ast.Property{
				"userId": {Type: "string", Required: true},
				"amount": {Type: "number", Required: true},
			}},
		},
	}
	s.UpdatePlan(plan2)

	// 4. Same payload should now fail
	req2 := httptest.NewRequest("POST", "/event", bytes.NewBuffer(payload))
	resp2 := httptest.NewRecorder()
	s.HandleEvent(resp2, req2)

	if resp2.Code != http.StatusBadRequest {
		t.Errorf("Expected BadRequest after reload, got %d", resp2.Code)
	}
}
