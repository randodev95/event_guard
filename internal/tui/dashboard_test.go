package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboard_View_Initial(t *testing.T) {
	m := NewDashboard(nil)
	view := m.View()

	if !strings.Contains(view, "Waiting for events...") {
		t.Errorf("Expected view to contain 'Waiting for events...', got: %s", view)
	}
}

func TestDashboard_Init(t *testing.T) {
	updates := make(chan EventMsg)
	m := NewDashboard(updates)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Expected Init to return a command")
	}
}

func TestDashboard_Update_Quit(t *testing.T) {
	m := NewDashboard(nil)
	
	// Test 'q' to quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Expected 'q' to return a command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
	_ = newModel
}

func TestDashboard_View_MultipleEvents(t *testing.T) {
	m := NewDashboard(nil)
	
	yaml := `
identity_properties: ["userId"]
events:
  "Test Event":
    category: "TEST"
    entity_type: "User"
    properties:
      userId: { type: string, required: true }
`
	_ = yaml

	res, _ := m.Update(EventMsg{Name: "Login", IsValid: true})
	m = res.(Dashboard)
	
	res, _ = m.Update(EventMsg{Name: "Purchase", IsValid: false, Errors: []string{"missing total"}})
	m = res.(Dashboard)
	
	view := m.View()
	
	if !strings.Contains(view, "Login") || !strings.Contains(view, "✓ VALID") {
		t.Errorf("Expected valid Login event in view")
	}
	if !strings.Contains(view, "Purchase") || !strings.Contains(view, "✗ INVALID") || !strings.Contains(view, "missing total") {
		t.Errorf("Expected invalid Purchase event with error in view")
	}
	if !strings.Contains(view, "Total events: 2") {
		t.Errorf("Expected total count 2, view: %s", view)
	}
}
