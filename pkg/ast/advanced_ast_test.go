package ast

import (
	"testing"
)

func TestResolveEventSchema_ImmutableConstraints(t *testing.T) {
	plan := &TrackingPlan{
		Contexts: map[string]Context{
			"Base": {
				Properties: map[string]Property{
					"userId": {Type: "string", Required: true},
				},
			},
		},
		Events: map[string]Event{
			"Violator": {
				Inherits: []string{"Base"},
				Properties: map[string]Property{
					"userId": {Type: "number", Required: true}, // Trying to change type
				},
			},
		},
	}

	_, err := plan.ResolveEventSchema("Violator")
	if err == nil {
		t.Error("Expected error when overriding parent property type, but got nil")
	}
}

func TestResolveEventSchema_ImmutableRequired(t *testing.T) {
	plan := &TrackingPlan{
		Contexts: map[string]Context{
			"Base": {
				Properties: map[string]Property{
					"userId": {Type: "string", Required: true},
				},
			},
		},
		Events: map[string]Event{
			"Violator": {
				Inherits: []string{"Base"},
				Properties: map[string]Property{
					"userId": {Type: "string", Required: false}, // Trying to make optional
				},
			},
		},
	}

	_, err := plan.ResolveEventSchema("Violator")
	if err == nil {
		t.Error("Expected error when weakening 'Required' constraint, but got nil")
	}
}

func TestValidateTaxonomy_MissingCategory(t *testing.T) {
	plan := &TrackingPlan{
		Events: map[string]Event{
			"Order Completed": {
				EntityType: "Transaction",
				// Missing Category
			},
		},
	}

	err := plan.ValidateTaxonomy()
	if err == nil {
		t.Error("Expected error for missing Category, but got nil")
	}
}

func TestValidateTaxonomy_MissingEntityType(t *testing.T) {
	plan := &TrackingPlan{
		Events: map[string]Event{
			"Order Completed": {
				Category: "INTERACTION",
				// Missing EntityType
			},
		},
	}

	err := plan.ValidateTaxonomy()
	if err == nil {
		t.Error("Expected error for missing EntityType, but got nil")
	}
}

func TestValidateIntegrity_MissingUserId(t *testing.T) {
	plan := &TrackingPlan{
		Events: map[string]Event{
			"Order Completed": {
				Category:   "INTERACTION",
				EntityType: "Transaction",
				Properties: map[string]Property{
					"total": {Type: "number", Required: true},
				},
				// Missing userId in properties and no inheritance providing it
			},
		},
	}

	err := plan.ValidateIntegrity()
	if err == nil {
		t.Error("Expected error for missing userId (integrity check), but got nil")
	}
}

func TestResolveProperties_Hierarchical(t *testing.T) {
	plan := &TrackingPlan{
		Contexts: map[string]Context{
			"Base": {Properties: map[string]Property{"p1": {Type: "string"}}},
			"Child": {
				Inherits:   []string{"Base"},
				Properties: map[string]Property{"p2": {Type: "number"}},
			},
		},
		Events: map[string]Event{
			"E": {Inherits: []string{"Child"}},
		},
	}

	props, err := plan.ResolveProperties("E")
	if err != nil {
		t.Fatalf("ResolveProperties failed: %v", err)
	}

	if _, ok := props["p1"]; !ok {
		t.Error("Missing property p1 from Base context")
	}
	if _, ok := props["p2"]; !ok {
		t.Error("Missing property p2 from Child context")
	}
}

func TestResolveProperties_Cycles(t *testing.T) {
	plan := &TrackingPlan{
		Contexts: map[string]Context{
			"A": {Inherits: []string{"B"}},
			"B": {Inherits: []string{"A"}},
		},
		Events: map[string]Event{
			"E": {Inherits: []string{"A"}},
		},
	}

	_, err := plan.ResolveProperties("E")
	if err == nil {
		t.Error("Expected error for circular inheritance (A -> B -> A), but got nil")
	}
}
