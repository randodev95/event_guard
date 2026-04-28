package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCommand_Version(t *testing.T) {
	rootCmd := NewRootCmd()
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := b.String()
	if !strings.Contains(out, "canvas version 0.1.0") {
		t.Errorf("Expected version 0.1.0, got: %s", out)
	}
}

func TestInitCommand(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewInitCmd())

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify files created
	if _, err := os.Stat("canvas.yaml"); os.IsNotExist(err) {
		t.Errorf("canvas.yaml was not created")
	}
	defer os.Remove("canvas.yaml")

	if _, err := os.Stat(".canvas/canvas.db"); os.IsNotExist(err) {
		t.Errorf(".canvas/canvas.db was not created")
	}
	defer os.RemoveAll(".canvas")
}

func TestImpactCheckCommand(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewImpactCheckCmd())

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"impact-check", "--prev-sha", "main"})

	err := rootCmd.Execute()
	// Should fail because database is missing in this test context
	if err == nil {
		t.Errorf("Expected impact-check to fail without database")
	}
}

func TestGenerateCommand(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewGenerateCmd())

	// Create a dummy canvas.yaml for the test
	yamlData := `
events:
  "Test Event":
    category: "INTERACTION"
    entity_type: "Component"
    properties:
      prop1: { type: string, required: true }
      userId: { type: string, required: true }
`
	os.WriteFile("canvas_test.yaml", []byte(yamlData), 0644)
	defer os.Remove("canvas_test.yaml")

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"generate", "--target", "dbt", "--plan", "canvas_test.yaml"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := b.String()
	if !strings.Contains(out, "name: Test Event") {
		t.Errorf("Generate output missing event name. Got: %s", out)
	}
}

func TestValidateCommand(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewValidateCmd())

	// Create dummy plan
	yamlData := `
events:
  "Login":
    category: "INTERACTION"
    entity_type: "User"
    properties:
      userId: { type: string, required: true }
`
	os.WriteFile("canvas_val.yaml", []byte(yamlData), 0644)
	defer os.Remove("canvas_val.yaml")

	// Create valid payload
	payload := `{"event": "Login", "properties": {"userId": "123"}}`
	os.WriteFile("payload.json", []byte(payload), 0644)
	defer os.Remove("payload.json")

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"validate", "payload.json", "--plan", "canvas_val.yaml"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := b.String()
	if !strings.Contains(out, "VALID") {
		t.Errorf("Expected VALID result, got: %s", out)
	}
}
