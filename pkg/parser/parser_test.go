package parser

import (
	"testing"
)

func TestParseYAML_Simple(t *testing.T) {
	yamlData := []byte(`
events:
  "Order Completed":
    category: "INTERACTION"
    entity_type: "Transaction"
    properties:
      userId: { type: string, required: true }
      total:
        type: number
        required: true
`)

	plan, err := ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	event, ok := plan.Events["Order Completed"]
	if !ok {
		t.Fatal("Expected 'Order Completed' event not found")
	}

	prop, ok := event.Properties["total"]
	if !ok {
		t.Fatal("Expected 'total' property not found")
	}

	if prop.Type != "number" {
		t.Errorf("Expected type 'number', got '%s'", prop.Type)
	}
}

func TestParseYAML_Inheritance(t *testing.T) {
	yamlData := []byte(`
contexts:
  Universal_Page_Context:
    properties:
      url:
        type: string
        required: true
      userId:
        type: string
        required: true
events:
  "Button Clicked":
    category: "INTERACTION"
    entity_type: "Button"
    inherits: ["Universal_Page_Context"]
    properties:
      button_id:
        type: string
        required: true
`)

	plan, err := ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	event, ok := plan.Events["Button Clicked"]
	if !ok {
		t.Fatal("Expected 'Button Clicked' event not found")
	}

	if len(event.Inherits) == 0 || event.Inherits[0] != "Universal_Page_Context" {
		t.Errorf("Expected event to inherit from 'Universal_Page_Context'")
	}
}

func TestResolveEventSchema(t *testing.T) {
	yamlData := []byte(`
contexts:
  Universal_Page_Context:
    properties:
      url:
        type: string
        required: true
      userId:
        type: string
        required: true
events:
  "Button Clicked":
    category: "INTERACTION"
    entity_type: "Button"
    inherits: ["Universal_Page_Context"]
    properties:
      button_id:
        type: string
        required: true
`)

	plan, err := ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	schema, err := plan.ResolveEventSchema("Button Clicked")
	if err != nil {
		t.Fatalf("ResolveEventSchema failed: %v", err)
	}

	// Simple check for presence of both properties in the resulting JSON schema string
	if !contains(schema, "url") || !contains(schema, "button_id") {
		t.Errorf("Resolved schema missing properties. Got: %s", schema)
	}
}

func TestParseYAML_IntegrityFailure(t *testing.T) {
	yamlData := []byte(`
events:
  "Login":
    category: "INTERACTION"
    entity_type: "User"
    properties:
      timestamp: { type: string, required: true }
    # Missing userId
`)
	_, err := ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error for integrity failure (missing userId), but got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
