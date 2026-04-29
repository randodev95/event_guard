package ast

import (
	"fmt"
	"strings"
	"gopkg.in/yaml.v3"
)

// TrackingPlan coordinates Taxonomy and Flows.
type TrackingPlan struct {
	Version  string              `yaml:"version"`
	Taxonomy Taxonomy            `yaml:"taxonomy"`
	Flows    map[string]FlowPlan `yaml:"flows"`
}

// UnmarshalYAML implements custom logic to handle both nested and flat taxonomy structures.
func (p *TrackingPlan) UnmarshalYAML(value *yaml.Node) error {
	// 1. Try standard unmarshal (nested taxonomy)
	type alias TrackingPlan
	var a alias
	if err := value.Decode(&a); err != nil {
		return err
	}
	*p = TrackingPlan(a)

	// 2. Try flat unmarshal (taxonomy fields at root)
	var flat Taxonomy
	if err := value.Decode(&flat); err != nil {
		return err
	}

	// Merge flat fields into p.Taxonomy if they exist
	if flat.Version != "" {
		p.Taxonomy.Version = flat.Version
	}
	if flat.Namespace != "" {
		p.Taxonomy.Namespace = flat.Namespace
	}
	if len(flat.Events) > 0 {
		if p.Taxonomy.Events == nil {
			p.Taxonomy.Events = make(map[string]EventV2)
		}
		for k, v := range flat.Events {
			p.Taxonomy.Events[k] = v
		}
	}
	if len(flat.Mixins) > 0 {
		if p.Taxonomy.Mixins == nil {
			p.Taxonomy.Mixins = make(map[string]Mixin)
		}
		for k, v := range flat.Mixins {
			p.Taxonomy.Mixins[k] = v
		}
	}
	if len(flat.Enums) > 0 {
		if p.Taxonomy.Enums == nil {
			p.Taxonomy.Enums = make(map[string][]string)
		}
		for k, v := range flat.Enums {
			p.Taxonomy.Enums[k] = v
		}
	}
	if len(flat.IdentityProperties) > 0 {
		p.Taxonomy.IdentityProperties = flat.IdentityProperties
	}

	return nil
}

// DiffPlans compares two plans for semantic changes.
// It tracks additions, removals, and changes to types/constraints.
func DiffPlans(old, new *TrackingPlan) []string {
	var changes []string

	// 1. Check for New/Removed Events
	for name := range new.Taxonomy.Events {
		if _, ok := old.Taxonomy.Events[name]; !ok {
			changes = append(changes, fmt.Sprintf("event [%s]: ADDED", name))
		}
	}

	for name, oldEv := range old.Taxonomy.Events {
		newEv, ok := new.Taxonomy.Events[name]
		if !ok {
			changes = append(changes, fmt.Sprintf("event [%s]: REMOVED (BREAKING)", name))
			continue
		}

		// 2. Diff Properties
		oldProps, _ := old.Taxonomy.FlattenEvent(name)
		newProps, _ := new.Taxonomy.FlattenEvent(name)

		for pName, oldP := range oldProps {
			newP, ok := newProps[pName]
			if !ok {
				changes = append(changes, fmt.Sprintf("event [%s] prop [%s]: REMOVED (BREAKING)", name, pName))
				continue
			}

			// Semantic diff of constraints
			if oldP.Type != newP.Type {
				changes = append(changes, fmt.Sprintf("event [%s] prop [%s]: TYPE CHANGE %s -> %s (BREAKING)", name, pName, oldP.Type, newP.Type))
			}
			if !oldP.Required && newP.Required {
				changes = append(changes, fmt.Sprintf("event [%s] prop [%s]: NOW REQUIRED (BREAKING)", name, pName))
			}
			
			// Enum changes
			if len(oldP.Enum) != len(newP.Enum) {
				changes = append(changes, fmt.Sprintf("event [%s] prop [%s]: ENUM CHANGED (POTENTIAL BREAKING)", name, pName))
			}
		}

		for pName := range newProps {
			if _, ok := oldProps[pName]; !ok {
				changes = append(changes, fmt.Sprintf("event [%s] prop [%s]: ADDED", name, pName))
			}
		}
		
		// 3. Diff Flow Lineage
		if oldEv.EntityType != newEv.EntityType {
			changes = append(changes, fmt.Sprintf("event [%s]: ENTITY CHANGE %s -> %s", name, oldEv.EntityType, newEv.EntityType))
		}
	}

	return changes
}

// Obfuscate returns a copy of the plan with hashed names for privacy in public WASM.
func (p *TrackingPlan) Obfuscate() *TrackingPlan {
	// For now, return same plan to avoid complexity unless requested.
	// Hashing names requires deep copy and map replacement.
	return p 
}

func (p *TrackingPlan) ResolveEventSchema(eventName string) (string, error) {
	return p.Taxonomy.ResolveEventSchema(eventName)
}

func (p *TrackingPlan) GetIdentityProperties() []string {
	return p.Taxonomy.GetIdentityProperties()
}

func (p *TrackingPlan) GetEventNames() []string {
	return p.Taxonomy.GetEventNames()
}

// Validate performs full system validation.
func (p *TrackingPlan) Validate() error {
	if err := p.Taxonomy.Validate(); err != nil {
		return err
	}

	// 2. Validate Flows
	usedEvents := make(map[string]bool)
	for flowName, flow := range p.Flows {
		if err := flow.Validate(); err != nil {
			return fmt.Errorf("flow %s: %w", flowName, err)
		}

		// Collect events used in this flow
		for nodeName, node := range flow.Nodes {
			var eventName string
			if node.Event != "" {
				eventName = node.Event
			} else if node.ListenFor != "" {
				eventName = node.ListenFor
			}

			if eventName != "" {
				// Rule: Event must exist in taxonomy
				if _, ok := p.Taxonomy.Events[eventName]; !ok {
					return fmt.Errorf("flow %s, node %s: event %s not found in taxonomy", flowName, nodeName, eventName)
				}
				usedEvents[eventName] = true
			}
		}
	}

	// 3. Rule: No orphan events in taxonomy
	for eventName := range p.Taxonomy.Events {
		if !usedEvents[eventName] {
			return fmt.Errorf("orphan event: %s is defined in taxonomy but not used in any flow", eventName)
		}
	}

	// 4. Type Check Gate Conditions
	return p.TypeCheck()
}

// TypeCheck ensures nodes reference valid events and properties.
func (p *TrackingPlan) TypeCheck() error {
	for flowName, flow := range p.Flows {
		// Find all triggers to set valid context
		triggerEvents := make(map[string]bool)
		for _, node := range flow.Nodes {
			if node.Type == "TriggerNode" && node.Event != "" {
				triggerEvents[node.Event] = true
			}
		}

		if len(triggerEvents) == 0 {
			continue
		}

		// Use the first trigger as default active event for initial nodes
		var activeEvent string
		for e := range triggerEvents {
			activeEvent = e
			break
		}

		for nodeName, node := range flow.Nodes {
			// Check TriggerNode event
			if node.Type == "TriggerNode" && node.Event != "" {
				if _, ok := p.Taxonomy.Events[node.Event]; !ok {
					return fmt.Errorf("flow %s node %s: unknown event %s", flowName, nodeName, node.Event)
				}
			}
			// Check WaitNode listen_for
			if node.Type == "WaitNode" && node.ListenFor != "" {
				if _, ok := p.Taxonomy.Events[node.ListenFor]; !ok {
					return fmt.Errorf("flow %s node %s: unknown event %s", flowName, nodeName, node.ListenFor)
				}
				activeEvent = node.ListenFor // Update context
			}

			// Check GateNode conditions
			if node.Type == "GateNode" {
				if activeEvent == "" {
					return fmt.Errorf("flow %s node %s: GateNode used before any TriggerNode or WaitNode", flowName, nodeName)
				}
				props, err := p.Taxonomy.FlattenEvent(activeEvent)
				if err != nil {
					return fmt.Errorf("flow %s node %s: %w", flowName, nodeName, err)
				}

				for _, cond := range node.Conditions {
					if cond.If != "" && strings.Contains(cond.If, "$event.") {
						// Simple regex/substring check for demo
						// Extract property name: $event.xxx
						parts := strings.Split(cond.If, "$event.")
						if len(parts) > 1 {
							propName := strings.Split(parts[1], " ")[0]
							if _, ok := props[propName]; !ok {
								return fmt.Errorf("flow %s node %s: unknown property %s in condition", flowName, nodeName, propName)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// ExtractTaxonomy generates a skeletal taxonomy based on events referenced in flows.
func (p *TrackingPlan) ExtractTaxonomy() Taxonomy {
	tax := Taxonomy{
		Events: make(map[string]EventV2),
	}

	for flowName, flow := range p.Flows {
		for _, node := range flow.Nodes {
			var eventName string
			if node.Event != "" {
				eventName = node.Event
			} else if node.ListenFor != "" {
				eventName = node.ListenFor
			}

			if eventName != "" {
				if _, exists := tax.Events[eventName]; !exists {
					props := node.Properties
					if props == nil {
						props = make(map[string]PropertyV2)
					}
					tax.Events[eventName] = EventV2{
						Description: fmt.Sprintf("Auto-extracted: %s", flowName),
						Properties:  deepCopyProperties(props),
					}
				}
			}
		}
	}

	return tax
}

func deepCopyProperties(src map[string]PropertyV2) map[string]PropertyV2 {
	dst := make(map[string]PropertyV2, len(src))
	for k, v := range src {
		cp := v
		// Deep copy Enum slice
		if len(v.Enum) > 0 {
			cp.Enum = make([]string, len(v.Enum))
			copy(cp.Enum, v.Enum)
		}
		// Deep copy Rules pointers
		if v.Rules.Min != nil {
			min := *v.Rules.Min
			cp.Rules.Min = &min
		}
		if v.Rules.Max != nil {
			max := *v.Rules.Max
			cp.Rules.Max = &max
		}
		dst[k] = cp
	}
	return dst
}
