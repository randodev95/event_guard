package ast

import (
	"testing"
)

func TestDiffPlans_BreakingChanges(t *testing.T) {
	oldPlan := &TrackingPlan{
		Version: "1.0.0",
		Events: map[string]Event{
			"Event A": {
				Properties: map[string]Property{
					"Prop1": {Type: "string", Required: true},
					"Prop2": {Type: "number"},
				},
			},
		},
	}

	newPlan := &TrackingPlan{
		Version: "1.1.0",
		Events: map[string]Event{
			"Event A": {
				Properties: map[string]Property{
					"Prop1": {Type: "integer", Required: true}, // Type change: BREAKING
					// Prop2 removed: BREAKING
				},
			},
		},
	}

	diffs := DiffPlans(oldPlan, newPlan)
	
	expectedBreaches := 2
	if len(diffs) != expectedBreaches {
		t.Errorf("Expected %d breaking changes, got %d: %v", expectedBreaches, len(diffs), diffs)
	}

	// Check for specific errors (simple string matching for now)
	foundTypeChange := false
	foundRemoval := false
	for _, d := range diffs {
		if contains(d, "Prop1") && contains(d, "type") {
			foundTypeChange = true
		}
		if contains(d, "Prop2") && contains(d, "removed") {
			foundRemoval = true
		}
	}

	if !foundTypeChange {
		t.Error("Expected breaking change for type change not found")
	}
	if !foundRemoval {
		t.Error("Expected breaking change for property removal not found")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
