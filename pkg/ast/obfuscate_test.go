package ast

import (
	"testing"
)

func TestTrackingPlan_Obfuscate(t *testing.T) {
	plan := &TrackingPlan{
		Events: map[string]Event{
			"Order Completed": {Category: "A", EntityType: "T"},
		},
	}

	obfuscated := plan.Obfuscate()
	
	// Check "Order Completed" no longer exists
	if _, ok := obfuscated.Events["Order Completed"]; ok {
		t.Error("Expected 'Order Completed' to be removed")
	}

	// Should have one event with a hashed name (SHA256 of "Order Completed")
	if len(obfuscated.Events) != 1 {
		t.Errorf("Expected 1 obfuscated event, got %d", len(obfuscated.Events))
	}
}
