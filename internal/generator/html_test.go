package generator

import (
	"github.com/randodev95/event_guard/pkg/parser"
	"strings"
	"testing"
)

func TestGenerateHTML_Snapshot(t *testing.T) {
	yamlData := `
identity_properties: ["wallet_address"]
contexts:
  Wallet_Context:
    properties:
      wallet_address: { type: string, required: true }
events:
  "Test Event":
    category: "TEST"
    entity_type: "TestEntity"
    inherits: ["Wallet_Context"]
    properties:
      test_prop: { type: string, required: true }
flows:
  - id: "test_flow"
    name: "Test Flow"
    steps:
      - state: "Start"
        event: "Test Event"
        triggers: ["DIRECT_LOAD"]
`
	plan, err := parser.ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse plan: %v", err)
	}

	html, err := GenerateHTML(plan)
	if err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Basic assertions to ensure core elements are present
	expectedSnippets := []string{
		"EventGuard Live Docs",
		"Flow: Test Flow (test_flow)",
		"Test Event",
		"test_prop",
		"wallet_address",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(html, snippet) {
			t.Errorf("Expected HTML to contain %q, but it didn't", snippet)
		}
	}
}
