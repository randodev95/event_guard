package validator

import (
	"testing"

	"github.com/eventcanvas/eventcanvas/pkg/normalization"
)

func TestValidate_Success(t *testing.T) {
	event := &normalization.NormalizedEvent{
		Event: "Order Completed",
		Properties: map[string]interface{}{
			"total": 50.0,
		},
		UserID: "user_123",
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
		UserID: "user_123",
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
		UserID: "user_123",
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
