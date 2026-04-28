package validator

import (
	"testing"

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
