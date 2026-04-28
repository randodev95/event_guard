package parser

import (
	"github.com/randodev95/event_guard/pkg/ast"
	"gopkg.in/yaml.v3"
)

// ParseYAML ingests a raw YAML byte slice and returns a validated TrackingPlan AST.
// It performs taxonomy, integrity, and flow validation during ingestion.
func ParseYAML(data []byte) (*ast.TrackingPlan, error) {
	var plan ast.TrackingPlan
	err := yaml.Unmarshal(data, &plan)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidateTaxonomy(); err != nil {
		return nil, err
	}

	if err := plan.ValidateIntegrity(); err != nil {
		return nil, err
	}

	if err := plan.ValidateFlows(); err != nil {
		return nil, err
	}

	return &plan, nil
}
