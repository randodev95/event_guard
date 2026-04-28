package tui

import (
	"strings"
	"testing"
)

func TestDashboard_View_Initial(t *testing.T) {
	m := NewDashboard(nil)
	view := m.View()

	if !strings.Contains(view, "Waiting for events...") {
		t.Errorf("Expected view to contain 'Waiting for events...', got: %s", view)
	}
}

func TestDashboard_Update(t *testing.T) {
	m := NewDashboard(nil)

	newModel, _ := m.Update(EventMsg{Name: "Order Completed", IsValid: true})
	dashboard := newModel.(Dashboard)

	if len(dashboard.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(dashboard.events))
	}

	view := dashboard.View()
	if !strings.Contains(view, "VALID") {
		t.Errorf("Expected view to contain 'VALID', got: %s", view)
	}
}
