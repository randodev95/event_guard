package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/randodev95/event_guard/pkg/ast"
	"gopkg.in/yaml.v3"
)

// LoadPlan intelligently loads a tracking plan from either a single YAML file or a directory of YAML files.
func LoadPlan(path string) (*ast.TrackingPlan, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return LoadProject(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseYAML(data)
}

// ParseYAML ingests a raw YAML byte slice and returns a validated TrackingPlan AST.
func ParseYAML(data []byte) (*ast.TrackingPlan, error) {
	var plan ast.TrackingPlan
	err := yaml.Unmarshal(data, &plan)
	if err != nil {
		return nil, err
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return &plan, nil
}

// LoadProject scans a directory for YAML files and merges them into a single TrackingPlan.
func LoadProject(dir string) (*ast.TrackingPlan, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	merged := &ast.TrackingPlan{
		Taxonomy: ast.Taxonomy{
			Events: make(map[string]ast.EventV2),
			Mixins: make(map[string]ast.Mixin),
			Enums:  make(map[string][]string),
		},
		Flows: make(map[string]ast.FlowPlan),
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		var p ast.TrackingPlan
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("error parsing %s: %w", file.Name(), err)
		}

		if p.Version != "" {
			merged.Version = p.Version
		}

		// Merge Taxonomy
		if p.Taxonomy.Version != "" {
			merged.Taxonomy.Version = p.Taxonomy.Version
		}
		if p.Taxonomy.Namespace != "" {
			merged.Taxonomy.Namespace = p.Taxonomy.Namespace
		}
		for k, v := range p.Taxonomy.Events {
			merged.Taxonomy.Events[k] = v
		}
		for k, v := range p.Taxonomy.Mixins {
			merged.Taxonomy.Mixins[k] = v
		}
		for k, v := range p.Taxonomy.Enums {
			merged.Taxonomy.Enums[k] = v
		}
		if len(p.Taxonomy.IdentityProperties) > 0 {
			merged.Taxonomy.IdentityProperties = p.Taxonomy.IdentityProperties
		}

		// Merge Flows
		for k, v := range p.Flows {
			merged.Flows[k] = v
		}
	}

	if err := merged.Validate(); err != nil {
		return nil, err
	}

	return merged, nil
}
