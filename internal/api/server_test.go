package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/pkg/validator"
)

func TestServer_ValidateHandler(t *testing.T) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId"]
events:
  "Login":
    category: "INTERACTION"
    entity_type: "User"
    properties:
      userId: { type: string, required: true }
`)
	plan, _ := parser.ParseYAML(yamlData)
	engine := validator.NewEngine(plan)
	server := NewServer(engine)

	t.Run("Valid Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"event":      "Login",
			"userId":     "user_123",
			"properties": map[string]interface{}{},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp validator.Result
		json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp.Valid {
			t.Error("Expected event to be valid")
		}
	})

	t.Run("Invalid Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"event":      "Login",
			"properties": map[string]interface{}{},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp validator.Result
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Valid {
			t.Error("Expected event to be invalid (missing userId)")
		}
	})
}
