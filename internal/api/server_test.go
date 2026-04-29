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

func mockPlanYAML(eventName string) string {
	return `version: "1.0.0"
taxonomy:
  events:
    "` + eventName + `":
      properties:
        userId: {type: string, required: true}
flows:
  Test:
    nodes:
      Start: { type: TriggerNode, event: "` + eventName + `", transitions: [{target: End}] }
      End: { type: TerminalNode }
`
}

func TestServer_ValidateHandler(t *testing.T) {
	yamlData := []byte(mockPlanYAML("Login"))
	plan, _ := parser.ParseYAML(yamlData)
	engine := validator.NewEngine(plan)
	server := NewServer(engine)

	t.Run("Valid Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"event":      "Login",
			"userId":     "user_123",
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
	initialYAML := mockPlanYAML("Login")
	plan, _ := parser.ParseYAML([]byte(initialYAML))
	engine := validator.NewEngine(plan)
	engine.Warmup()

	srv := NewServer(engine)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	t.Run("Initial Validation", func(t *testing.T) {
		payload := []byte(`{"event": "Login", "userId": "u1"}`)
		resp, _ := http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(payload))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Initial validation failed: %v", resp.StatusCode)
		}
	})

	t.Run("Hot Reload Cycle", func(t *testing.T) {
		updatedYAML := `version: "1.1.0"
taxonomy:
  events:
    "Login":
      properties:
        userId: {type: string, required: true}
        new_v: {type: string, required: true}
flows:
  Test:
    nodes:
      Start: { type: TriggerNode, event: "Login", transitions: [{target: End}] }
      End: { type: TerminalNode }
`
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

		resp, _ := http.Post(ts.URL+"/admin/reload", "application/json", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatal("Reload failed")
		}

		oldPayload := []byte(`{"event": "Login", "userId": "u1"}`)
		resp, _ = http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(oldPayload))
		
		var result validator.Result
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Valid {
			t.Error("Expected INVALID after reload due to missing 'new_v' property")
		}
	})
}

func TestEventGuard_E2E_SinkSafety(t *testing.T) {
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer webhook.Close()

	initialYAML := mockPlanYAML("E")
	plan, _ := parser.ParseYAML([]byte(initialYAML))
	engine := validator.NewEngine(plan)
	srv := NewServer(engine)
	baseSink := NewWebhookSink(webhook.URL)
	sink := NewAsyncSink(baseSink, 1, 1)
	defer sink.Close()
	srv.SetSink(sink)

	ts := httptest.NewServer(srv)
	defer ts.Close()

	t.Run("Non-Blocking Validation on Sink Failure", func(t *testing.T) {
		payload := []byte(`{"event": "E", "properties": {}}`)
		start := time.Now()
		resp, _ := http.Post(ts.URL+"/validate", "application/json", bytes.NewBuffer(payload))
		duration := time.Since(start)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Validation endpoint blocked or failed: %v", resp.StatusCode)
		}
		if duration > 100*time.Millisecond {
			t.Errorf("Ingestion path blocked for %v, expected < 100ms", duration)
		}
	})
}
