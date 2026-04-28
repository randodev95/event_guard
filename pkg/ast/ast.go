package ast

import (
	"encoding/json"
	"fmt"
)

type TrackingPlan struct {
	Version            string             `yaml:"version"`
	Contexts           map[string]Context `yaml:"contexts"`
	Events             map[string]Event   `yaml:"events"`
	Flows              []Flow             `yaml:"flows"`
	IdentityProperties []string           `yaml:"identity_properties"` // e.g., ["wallet_address", "anonymousId"]
}

type Context struct {
	Inherits   []string            `yaml:"inherits"`
	EntityType string              `yaml:"entity_type"` // e.g., User, Session, Device
	Properties map[string]Property `yaml:"properties"`
}

type Event struct {
	Category   string              `yaml:"category"` // e.g., PAGE_VIEW, INTERACTION, VISIBILITY
	EntityType string              `yaml:"entity_type"`
	Inherits   []string            `yaml:"inherits"`
	Properties map[string]Property `yaml:"properties"`
}

type Property struct {
	Type     string   `yaml:"type"`
	Required bool     `yaml:"required"`
	NoNull   bool     `yaml:"no_null"`
	Unique   bool     `yaml:"unique"`
	Min      *float64 `yaml:"min"`
	Max      *float64 `yaml:"max"`
	Enum     []interface{} `yaml:"enum"`
}

func (p *TrackingPlan) ResolveEventSchema(eventName string) (string, error) {
	allProps, err := p.ResolveProperties(eventName)
	if err != nil {
		return "", err
	}

	var required []string
	for name, prop := range allProps {
		if prop.Required {
			required = append(required, name)
		}
	}

	// Generate JSON Schema
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   required,
	}

	props := schema["properties"].(map[string]interface{})
	for name, prop := range allProps {
		propSchema := map[string]interface{}{"type": prop.Type}
		if prop.NoNull {
			propSchema["nullable"] = false
		}
		if prop.Unique {
			propSchema["unique"] = true
		}
		if prop.Min != nil {
			propSchema["minimum"] = *prop.Min
		}
		if prop.Max != nil {
			propSchema["maximum"] = *prop.Max
		}
		if len(prop.Enum) > 0 {
			propSchema["enum"] = prop.Enum
		}
		props[name] = propSchema
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	return string(schemaJSON), nil
}

func (p *TrackingPlan) ValidateTaxonomy() error {
	for name, event := range p.Events {
		if event.Category == "" {
			return fmt.Errorf("event %s: missing category (e.g., PAGE_VIEW, INTERACTION)", name)
		}
		if event.EntityType == "" {
			return fmt.Errorf("event %s: missing entity_type (e.g., User, Session)", name)
		}
	}
	return nil
}

func (p *TrackingPlan) ValidateIntegrity() error {
	keys := p.IdentityProperties
	if len(keys) == 0 {
		keys = []string{"userId"}
	}

	for name := range p.Events {
		props, err := p.ResolveProperties(name)
		if err != nil {
			return err
		}

		// Logical completeness: prevent "Ghost Users"
		found := false
		for _, k := range keys {
			if _, ok := props[k]; ok {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("event %s: missing at least one logical identifier from %v (integrity breach)", name, keys)
		}
	}
	return nil
}

func (p *TrackingPlan) ResolveProperties(eventName string) (map[string]Property, error) {
	event, ok := p.Events[eventName]
	if !ok {
		return nil, fmt.Errorf("event %s not found", eventName)
	}

	allProps := make(map[string]Property)
	visited := make(map[string]bool)

	// Resolve inheritance recursively
	for _, ctxName := range event.Inherits {
		if err := p.resolveContext(ctxName, allProps, visited); err != nil {
			return nil, err
		}
	}

	// Add event specific properties with immutable checks
	for name, prop := range event.Properties {
		if parentProp, exists := allProps[name]; exists {
			if parentProp.Type != prop.Type {
				return nil, fmt.Errorf("property %s: cannot change type from %s to %s", name, parentProp.Type, prop.Type)
			}
			if parentProp.Required && !prop.Required {
				return nil, fmt.Errorf("property %s: cannot make optional", name)
			}
		}
		allProps[name] = prop
	}

	return allProps, nil
}

func (p *TrackingPlan) resolveContext(name string, allProps map[string]Property, visited map[string]bool) error {
	if visited[name] {
		return fmt.Errorf("circular inheritance detected in context %s", name)
	}
	visited[name] = true

	ctx, ok := p.Contexts[name]
	if !ok {
		return fmt.Errorf("context %s not found", name)
	}

	// Resolve parent contexts first
	for _, parentName := range ctx.Inherits {
		if err := p.resolveContext(parentName, allProps, visited); err != nil {
			return err
		}
	}

	// Add current context properties
	for propName, prop := range ctx.Properties {
		allProps[propName] = prop
	}

	// Standard DFS cycle detection
	visited[name] = false // Remove from recursion stack
	return nil
}
