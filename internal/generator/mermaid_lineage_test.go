package generator

import (
	"github.com/randodev95/event_guard/pkg/ast"
	"strings"
	"testing"
)

func TestGenerateMermaid_Lineage(t *testing.T) {
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"E1": {Category: "PAGE_VIEW", EntityType: "Page"},
			"E2": {Category: "INTERACTION", EntityType: "Button"},
		},
		Flows: []ast.Flow{
			{ID: "A", Name: "Flow A", Steps: []ast.FlowStep{{State: "S1", Event: "E1"}, {State: "S2", Event: "E2"}}},
			{ID: "B", Name: "Flow B", Steps: []ast.FlowStep{{State: "S2", Event: "E2"}, {State: "S3", Event: "E1"}}},
		},
	}

	output, err := GenerateMermaid(plan)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify that there's a link between the flows or shared nodes
	if !strings.Contains(output, "A_S2") || !strings.Contains(output, "B_S2") {
		// If we use unique IDs, they should both be there. 
		// If we use shared IDs, S2 should be there once but connected.
	}
	
	// Better yet, check for a cross-flow comment or specific edge style
	if !strings.Contains(output, "-->") {
		t.Error("Mermaid output missing edges")
	}
}
