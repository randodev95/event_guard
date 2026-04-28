package parser

import (
	"gopkg.in/yaml.v3"
	"github.com/randodev95/event_guard/pkg/ast"
)

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
