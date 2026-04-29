package validator

import (
	"os"
	"testing"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/normalization"
	"github.com/randodev95/event_guard/pkg/parser"
)

func TestEngine_Validate(t *testing.T) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId"]
events:
  "Staked":
    category: "TRANSACTION"
    entity_type: "Wallet"
    properties:
      userId: { type: string, required: true }
      amount: { type: number, required: true }
`)
	plan, err := parser.ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}
	engine := NewEngine(plan)

	// Valid event
	payload := []byte(`{"event": "Staked", "userId": "0x123", "properties": {"amount": 10.5}}`)
	result, err := engine.ValidateJSON(payload)
	if err != nil || !result.Valid {
		t.Errorf("Expected valid event, got err: %v, result: %v", err, result)
	}

	// Invalid event (missing property)
	invalidPayload := []byte(`{"event": "Staked", "userId": "0x123", "properties": {}}`)
	result, _ = engine.ValidateJSON(invalidPayload)
	if result.Valid {
		t.Error("Expected invalid event (missing required property), but it passed")
	}
}

func TestEngine_CasingMapping(t *testing.T) {
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
	engine := NewEngine(plan)

	// Test snake_case input mapping to camelCase schema
	payload := []byte(`{"event": "Login", "user_id": "user_123"}`)
	result, err := engine.ValidateJSON(payload)
	if err != nil || !result.Valid {
		t.Errorf("Expected valid mapping from user_id to userId, got err: %v, result: %v", err, result)
	}
}

func TestValidate_Success(t *testing.T) {
	event := &normalization.NormalizedEvent{
		Event: "Order Completed",
		Properties: map[string]interface{}{
			"total": 50.0,
		},
		Identity: map[string]string{
			"userId": "user_123",
		},
	}

	schema := `{
		"type": "object",
		"properties": {
			"total": { "type": "number" }
		},
		"required": ["total"]
	}`

	result, err := Validate(event, schema)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected event to be valid, but got errors: %v", result.Errors)
	}
}

func TestValidate_TypeMismatch(t *testing.T) {
	event := &normalization.NormalizedEvent{
		Event: "Order Completed",
		Properties: map[string]interface{}{
			"total": "fifty", // Should be number
		},
		Identity: map[string]string{
			"userId": "user_123",
		},
	}

	schema := `{
		"type": "object",
		"properties": {
			"total": { "type": "number" }
		}
	}`

	result, err := Validate(event, schema)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected event to be invalid due to type mismatch")
	}
}

func TestValidate_EnvelopeInjection(t *testing.T) {
	event := &normalization.NormalizedEvent{
		Event: "Login",
		Properties: map[string]interface{}{
			"foo": "bar",
		},
		Identity: map[string]string{
			"userId": "user_123",
		},
	}

	// Schema requires userId, which is only at the root (event.UserID)
	schema := `{
		"type": "object",
		"properties": {
			"userId": { "type": "string" }
		},
		"required": ["userId"]
	}`

	result, err := Validate(event, schema)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected event to be valid via envelope injection, but got errors: %v", result.Errors)
	}
}

func TestEngine_IdentityGuard(t *testing.T) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId", "anonymousId"]
events:
  "Page View":
    category: "PAGE_VIEW"
    entity_type: "Page"
    properties:
      userId: { type: string }
      anonymousId: { type: string }
      url: { type: string, required: true }
`)
	plan, err := parser.ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}
	engine := NewEngine(plan)

	// Case 1: Valid identity (userId)
	payload := []byte(`{"event": "Page View", "userId": "user_1", "properties": {"url": "/home"}}`)
	res, _ := engine.ValidateJSON(payload)
	if !res.Valid {
		t.Errorf("Expected valid identity to pass, got: %v", res.Errors)
	}

	// Case 2: Missing identity (Production Guard should block)
	invalidPayload := []byte(`{"event": "Page View", "properties": {"url": "/home"}}`)
	res, _ = engine.ValidateJSON(invalidPayload)
	if res.Valid {
		t.Error("Expected IdentityGuard to fail event with NO identity properties")
	}
	
	foundIdentityError := false
	for _, e := range res.Errors {
		if e == "identity_required: no recognized identity properties found" {
			foundIdentityError = true
		}
	}
	if !foundIdentityError {
		t.Errorf("Expected specific identity error, got: %v", res.Errors)
	}
}

func TestEngine_Warmup(t *testing.T) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId"]
events:
  "E1": { category: "A", entity_type: "T", properties: { userId: {type: string} } }
  "E2": { category: "B", entity_type: "T", properties: { userId: {type: string} } }
`)
	plan, err := parser.ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}
	engine := NewEngine(plan)

	if err := engine.Warmup(); err != nil {
		t.Fatalf("Warmup failed: %v", err)
	}
}

func BenchmarkEngine_ValidateJSON(b *testing.B) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId"]
events:
  "Staked":
    category: "TRANSACTION"
    entity_type: "Wallet"
    properties:
      userId: { type: string, required: true }
      amount: { type: number, required: true }
`)
	plan, _ := parser.ParseYAML(yamlData)
	engine := NewEngine(plan)
	engine.Warmup()

	payload := []byte(`{"event": "Staked", "userId": "0x123", "properties": {"amount": 10.5}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ValidateJSON(payload)
	}
}

type mockObserver struct {
	starts []string
	ends   []string
}

func (m *mockObserver) OnValidationStart(eventName string) {
	m.starts = append(m.starts, eventName)
}
func (m *mockObserver) OnValidationEnd(eventName string, duration int64, valid bool, err error) {
	m.ends = append(m.ends, eventName)
}

func TestEngine_ObservationHandler(t *testing.T) {
	yamlData := []byte(`
version: "1.0.0"
identity_properties: ["userId"]
events:
  "E1": { category: "A", entity_type: "T", properties: { userId: {type: string} } }
`)
	plan, _ := parser.ParseYAML(yamlData)
	engine := NewEngine(plan)
	observer := &mockObserver{}
	engine.SetObservationHandler(observer)

	payload := []byte(`{"event": "E1", "userId": "user_1"}`)
	_, _ = engine.ValidateJSON(payload)

	if len(observer.starts) != 1 || observer.starts[0] != "E1" {
		t.Errorf("Expected start observation for E1, got %v", observer.starts)
	}
	if len(observer.ends) != 1 || observer.ends[0] != "E1" {
		t.Errorf("Expected end observation for E1, got %v", observer.ends)
	}
}

func BenchmarkValidateJSON_DiamondPlan(b *testing.B) {
	data, err := os.ReadFile("../../canvas.yaml")
	if err != nil {
		b.Fatalf("Failed to read canvas.yaml: %v", err)
	}
	plan, err := parser.ParseYAML(data)
	if err != nil {
		b.Fatalf("Failed to parse canvas.yaml: %v", err)
	}
	engine := NewEngine(plan)
	engine.Warmup()

	payload := []byte(`{
		"event": "Order Completed",
		"userId": "u123",
		"anonymousId": "anon_999",
		"context": {
			"browser": {
				"user_agent": "Mozilla/5.0...",
				"viewport_width": 1920,
				"viewport_height": 1080,
				"language": "en-US"
			},
			"page": {
				"url": "https://shop.com/checkout",
				"path": "/checkout",
				"trigger": "click"
			},
			"campaign": {
				"source": "google",
				"medium": "cpc",
				"name": "spring_sale"
			}
		},
		"properties": {
			"library_name": "event_guard_js",
			"library_version": "2.1.0",
			"session_id": "sess_444",
			"order_id": "ord_555",
			"total": 120.50,
			"revenue": 100.00,
			"currency": "USD",
			"products": {
				"count": 4
			}
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ValidateJSON(payload)
	}
}

func TestEngine_ComplexDeepPayload(t *testing.T) {
	plan := &ast.TrackingPlan{
		IdentityProperties: []string{"userId", "anonymousId"},
		Events: map[string]ast.Event{
			"Order Completed": {
				Category:   "TRANSACTION",
				EntityType: "Order",
				Properties: map[string]ast.Property{
					"total":           {Type: "number", Required: true},
					"currency":        {Type: "string", Required: true, Enum: []interface{}{"USD", "EUR"}},
					"products.count":  {Type: "number", Required: true},
					"context.os.name": {Type: "string", Required: true},
				},
			},
		},
	}

	engine := NewEngine(plan)
	engine.Warmup()

	t.Run("Valid Rudderstack-style Payload", func(t *testing.T) {
		payload := []byte(`{
			"event": "Order Completed",
			"userId": "u123",
			"properties": {
				"total": 99.99,
				"currency": "USD",
				"products": {
					"count": 3
				}
			},
			"context": {
				"os": {
					"name": "iOS"
				}
			}
		}`)

		result, err := engine.ValidateJSON(payload)
		if err != nil {
			t.Fatalf("Validation error: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("Invalid Nested Property (Enum Violation)", func(t *testing.T) {
		payload := []byte(`{
			"event": "Order Completed",
			"userId": "u123",
			"properties": {
				"total": 99.99,
				"currency": "BITCOIN",
				"products": {
					"count": 3
				}
			},
			"context": {
				"os": {
					"name": "iOS"
				}
			}
		}`)

		result, _ := engine.ValidateJSON(payload)
		if result.Valid {
			t.Error("Expected invalid due to currency enum, but got valid")
		}
	})
}
