package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestEventGuard_E2E_DeploymentCycle(t *testing.T) {
	// 1. Setup Initial Plan
	initialYAML := `
version: "1.0.0"
identity_properties: ["userId"]
events:
  "Login":
    category: "AUTH"
    entity_type: "User"
    properties:
      userId: {type: string, required: true}
      platform: {type: string, required: true}
`
	plan, err := parser.ParseYAML([]byte(initialYAML))
	if err != nil {
		t.Fatalf("Failed to parse initial plan: %v", err)
	}
	engine := validator.NewEngine(plan)
	engine.Warmup()

	srv := NewServer(engine)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// 2. Test Success Path
	t.Run("Initial Validation", func(t *testing.T) {
		payload := []byte(`{"event": "Login", "userId": "u1", "properties": {"platform": "ios"}}`)
		resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(payload))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Initial validation failed: %v", resp.StatusCode)
		}
	})

	// 3. Simulate Hot Reload (Analyst updates plan)
	t.Run("Hot Reload Cycle", func(t *testing.T) {
		updatedYAML := `
version: "1.1.0"
identity_properties: ["userId"]
events:
  "Login":
    category: "AUTH"
    entity_type: "User"
    properties:
      userId: {type: string, required: true}
      platform: {type: string, required: true}
      version: {type: string, required: true}
`
		// Setup reload handler
		srv.SetReloadHandler(func() error {
			newPlan, err := parser.ParseYAML([]byte(updatedYAML))
			if err != nil {
				return err
			}
			newEngine := validator.NewEngine(newPlan)
			newEngine.Warmup()
			srv.UpdateEngine(newEngine)
			return nil
		})

		// Trigger reload
		resp, _ := http.Post(ts.URL+"/admin/reload", "application/json", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatal("Reload failed")
		}

		// Verify new contract is enforced
		oldPayload := []byte(`{"event": "Login", "userId": "u1", "properties": {"platform": "ios"}}`)
		resp, _ = http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(oldPayload))
		
		var result validator.Result
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Valid {
			t.Error("Expected INVALID after reload due to missing 'version' property")
		}
	})
}

func TestEventGuard_E2E_SinkSafety(t *testing.T) {
	// 1. Setup Hanging Webhook
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Beyond the 2s timeout
	}))
	defer webhook.Close()

	initialYAML := `
version: "1.0.0"
identity_properties: ["userId"]
events:
  "E": { category: "A", entity_type: "T", properties: { userId: {type: string, required: true}, p: {type: string, required: true} } }
`
	plan, err := parser.ParseYAML([]byte(initialYAML))
	if err != nil {
		t.Fatalf("Failed to parse sink test plan: %v", err)
	}
	engine := validator.NewEngine(plan)
	
	srv := NewServer(engine)
	
	// Use AsyncSink with small buffer and 1 worker
	baseSink := NewWebhookSink(webhook.URL)
	sink := NewAsyncSink(baseSink, 1, 1)
	defer sink.Close()
	srv.SetSink(sink)

	ts := httptest.NewServer(srv)
	defer ts.Close()

	t.Run("Non-Blocking Validation on Sink Failure", func(t *testing.T) {
		// Send INVALID event (trigger sink)
		payload := []byte(`{"event": "E", "properties": {}}`)
		
		start := time.Now()
		resp, _ := http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(payload))
		duration := time.Since(start)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Validation endpoint blocked or failed: %v", resp.StatusCode)
		}
		
		// If it blocks for 2 seconds, the sink is blocking the ingestion path
		if duration > 100*time.Millisecond {
			t.Errorf("Ingestion path blocked for %v, expected < 100ms", duration)
		}
	})
}
