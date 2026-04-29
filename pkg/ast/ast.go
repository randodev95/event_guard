package ast

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// TrackingPlan represents the root of the telemetry taxonomy.
// It defines the version, global contexts, events, and flows.
type TrackingPlan struct {
	Version            string             `yaml:"version"`
	Contexts           map[string]Context `yaml:"contexts"`
	Events             map[string]Event   `yaml:"events"`
	Flows              []Flow             `yaml:"flows"`
	IdentityProperties []string           `yaml:"identity_properties"` // e.g., ["wallet_address", "anonymousId"]
}

func (p *TrackingPlan) GetIdentityProperties() []string {
	return p.IdentityProperties
}

func (p *TrackingPlan) GetEventNames() []string {
	names := make([]string, 0, len(p.Events))
	for name := range p.Events {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Context defines a reusable set of properties that can be inherited by events.
// It typically represents a domain entity like User, Session, or Device.
type Context struct {
	Inherits   []string            `yaml:"inherits"`
	EntityType string              `yaml:"entity_type"` // e.g., User, Session, Device
	Properties map[string]Property `yaml:"properties"`
}

// Event represents a single tracking point in the application.
type Event struct {
	Category   string              `yaml:"category"` // e.g., PAGE_VIEW, INTERACTION, VISIBILITY
	EntityType string              `yaml:"entity_type"`
	Inherits   []string            `yaml:"inherits"`
	Properties map[string]Property `yaml:"properties"`
	Triggers   []Trigger           `yaml:"triggers"`
}

// Trigger describes how an event is arrived at from a specific state.
type Trigger struct {
	FromState string `yaml:"from_state"`
	Type      string `yaml:"type"` // e.g., UI_NAVIGATION, DIRECT_LOAD
}

// Property defines the constraints and type for a specific data field.
type Property struct {
	Type     string        `yaml:"type"`
	Required bool          `yaml:"required"`
	NoNull   bool          `yaml:"no_null"`
	Unique   bool          `yaml:"unique"`
	Min      *float64      `yaml:"min"`
	Max      *float64      `yaml:"max"`
	Enum     []interface{} `yaml:"enum"`
}

// ResolveEventSchema generates a JSON Schema representation for a specific event.
// It resolves all inherited properties and constraints.
func (p *TrackingPlan) ResolveEventSchema(eventName string) (string, error) {
	allProps, err := p.ResolveProperties(eventName)
	if err != nil {
		return "", fmt.Errorf("property resolution failed for %s: %w", eventName, err)
	}

	required := []string{}
	for name, prop := range allProps {
		if prop.Required {
			required = append(required, name)
		}
	}

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

// ValidateTaxonomy checks if all events have mandatory metadata fields like category and entity_type.
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

// ValidateIntegrity ensures that each event contains at least one identity property, preventing "Ghost Users".
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

// ResolveProperties flattens all properties for an event, including inherited contexts.
func (p *TrackingPlan) ResolveProperties(eventName string) (map[string]Property, error) {
	event, ok := p.Events[eventName]
	if !ok {
		return nil, fmt.Errorf("event %s not found", eventName)
	}

	allProps := make(map[string]Property)
	visited := make(map[string]bool)

	for _, ctxName := range event.Inherits {
		if err := p.resolveContext(ctxName, allProps, visited, 0); err != nil {
			return nil, err
		}
	}

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

func (p *TrackingPlan) resolveContext(name string, allProps map[string]Property, visited map[string]bool, depth int) error {
	if depth > 20 {
		return fmt.Errorf("inheritance depth exceeded limit (20) in context %s", name)
	}
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
		if err := p.resolveContext(parentName, allProps, visited, depth+1); err != nil {
			return err
		}
	}

	// Add current context properties
	for propName, prop := range ctx.Properties {
		allProps[propName] = prop
	}

	visited[name] = false
	return nil
}

// Obfuscate creates a copy of the plan with all event and context names hashed.
// This is used for public WASM exports to protect business logic.
func (p *TrackingPlan) Obfuscate() *TrackingPlan {
	newPlan := &TrackingPlan{
		Version:            p.Version,
		Contexts:           make(map[string]Context),
		Events:             make(map[string]Event),
		IdentityProperties: p.IdentityProperties,
	}

	hash := func(s string) string {
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])
	}

	// Map old names to new hashed names
	ctxMap := make(map[string]string)
	for name := range p.Contexts {
		ctxMap[name] = hash(name)
	}

	for name, ctx := range p.Contexts {
		newCtx := Context{
			EntityType: ctx.EntityType,
			Properties: ctx.Properties,
			Inherits:   make([]string, len(ctx.Inherits)),
		}
		for i, parent := range ctx.Inherits {
			newCtx.Inherits[i] = ctxMap[parent]
		}
		newPlan.Contexts[ctxMap[name]] = newCtx
	}

	for name, event := range p.Events {
		newEvent := Event{
			Category:   event.Category,
			EntityType: event.EntityType,
			Properties: event.Properties,
			Triggers:   event.Triggers,
			Inherits:   make([]string, len(event.Inherits)),
		}
		for i, parent := range event.Inherits {
			newEvent.Inherits[i] = ctxMap[parent]
		}
		newPlan.Events[hash(name)] = newEvent
	}

	return newPlan
}
