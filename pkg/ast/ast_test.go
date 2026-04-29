package ast

import (
	"testing"
)

func TestTaxonomy_CircularDependency(t *testing.T) {
	tax := &Taxonomy{
		Mixins: map[string]Mixin{
			"A": {Imports: []string{"B"}},
			"B": {Imports: []string{"A"}},
		},
		Events: map[string]EventV2{
			"Login": {Imports: []string{"A"}},
		},
	}

	_, err := tax.FlattenEvent("Login")
	if err == nil {
		t.Error("Expected error for circular mixin dependency, got nil")
	} else if err.Error() != "circular dependency detected in mixins: [A -> B -> A]" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
