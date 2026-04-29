package ast

import (
	"encoding/json"
	"fmt"
)

// ResolveEventSchema generates a JSON Schema for the event.
func (t *Taxonomy) ResolveEventSchema(eventName string) (string, error) {
	props, err := t.FlattenEvent(eventName)
	if err != nil {
		return "", err
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}

	required := []string{}
	schemaProps := schema["properties"].(map[string]interface{})

	for name, prop := range props {
		// Basic type
		pSchema := map[string]interface{}{
			"type": prop.Type,
		}

		// Rules
		if prop.Rules.Min != nil {
			pSchema["minimum"] = *prop.Rules.Min
		}
		if prop.Rules.Max != nil {
			pSchema["maximum"] = *prop.Rules.Max
		}
		if prop.Rules.MinLength != nil {
			pSchema["minLength"] = *prop.Rules.MinLength
		}
		if prop.Rules.MaxLength != nil {
			pSchema["maxLength"] = *prop.Rules.MaxLength
		}
		if prop.Rules.Pattern != "" {
			pSchema["pattern"] = prop.Rules.Pattern
		}

		// Enum resolution
		if prop.Type == "enum" && prop.Ref != "" {
			if enumValues, ok := t.Enums[prop.Ref]; ok {
				pSchema["type"] = "string"
				pSchema["enum"] = enumValues
			}
		}

		schemaProps[name] = pSchema
		if prop.Required {
			required = append(required, name)
		}
	}

	schema["required"] = required

	b, err := json.Marshal(schema)
	return string(b), err
}

// FlattenEvent merges mixins and local properties. Local overrides win.
func (t *Taxonomy) FlattenEvent(name string) (map[string]PropertyV2, error) {
	event, ok := t.Events[name]
	if !ok {
		return nil, fmt.Errorf("event %s not found", name)
	}

	flat := make(map[string]PropertyV2)
	visited := make(map[string]bool)

	// 1. Merge imports
	for _, mixinName := range event.Imports {
		if err := t.mergeMixin(mixinName, flat, visited, []string{}); err != nil {
			return nil, err
		}
	}

	// 2. Local overrides
	for pName, prop := range event.Properties {
		flat[pName] = prop
	}

	return flat, nil
}

func (t *Taxonomy) mergeMixin(name string, flat map[string]PropertyV2, visited map[string]bool, stack []string) error {
	// Check for circular dependency
	for _, s := range stack {
		if s == name {
			return fmt.Errorf("circular dependency detected in mixins: [%s -> %s]", formatStack(stack), name)
		}
	}

	mixin, ok := t.Mixins[name]
	if !ok {
		return fmt.Errorf("mixin %s not found", name)
	}

	// Avoid redundant processing
	if visited[name] {
		return nil
	}

	newStack := append(stack, name)

	// Recursive imports
	for _, subMixin := range mixin.Imports {
		if err := t.mergeMixin(subMixin, flat, visited, newStack); err != nil {
			return err
		}
	}

	// Merge properties
	for pName, prop := range mixin.Properties {
		flat[pName] = prop
	}

	visited[name] = true
	return nil
}

func formatStack(stack []string) string {
	var res string
	for i, s := range stack {
		if i > 0 {
			res += " -> "
		}
		res += s
	}
	return res
}

// Taxonomy defines the data contract domain.
type Taxonomy struct {
	Version            string              `yaml:"version"`
	Namespace          string              `yaml:"namespace"`
	Enums              map[string][]string `yaml:"enums"`
	Mixins             map[string]Mixin    `yaml:"mixins"`
	Events             map[string]EventV2  `yaml:"events"`
	IdentityProperties []string            `yaml:"identity_properties"`
}

func (t *Taxonomy) Validate() error {
	// TODO: check for circular mixin dependencies
	return nil
}

func (t *Taxonomy) GetEventNames() []string {
	names := make([]string, 0, len(t.Events))
	for name := range t.Events {
		names = append(names, name)
	}
	return names
}

func (t *Taxonomy) GetIdentityProperties() []string {
	return t.IdentityProperties
}

// Mixin represents a reusable block of properties.
type Mixin struct {
	Description string              `yaml:"description"`
	Imports     []string            `yaml:"imports"`
	Properties  map[string]PropertyV2 `yaml:"properties"`
}

// EventV2 represents a telemetry event in the new taxonomy.
type EventV2 struct {
	Description string              `yaml:"description"`
	EntityType  string              `yaml:"entity_type,omitempty"`
	Imports     []string            `yaml:"imports"`
	Properties  map[string]PropertyV2 `yaml:"properties"`
}

// PropertyV2 defines a field with strict rules.
type PropertyV2 struct {
	Type        string        `yaml:"type"`
	Required    bool          `yaml:"required"`
	Enum        []string      `yaml:"enum,omitempty"`
	Ref         string        `yaml:"ref,omitempty"` // For enums
	Description string        `yaml:"description,omitempty"`
	Rules       PropertyRules `yaml:"rules,omitempty"`
}

// PropertyRules contains validation constraints.
type PropertyRules struct {
	Min       *float64 `yaml:"min,omitempty"`
	Max       *float64 `yaml:"max,omitempty"`
	MinLength *int     `yaml:"min_length,omitempty"`
	MaxLength *int     `yaml:"max_length,omitempty"`
	Pattern   string   `yaml:"pattern,omitempty"`
}
