package generator

import (
	"github.com/randodev95/event_guard/pkg/ast"
	"sort"
)

// ResolvedEvent is the common logical representation of an event after inheritance resolution.
type ResolvedEvent struct {
	Name        string
	Description string
	Properties  map[string]ast.PropertyV2
}

// getResolvedEvents returns a sorted list of events with their properties resolved.
// This ensures stable output across all generators.
func getResolvedEvents(plan *ast.TrackingPlan) ([]ResolvedEvent, error) {
	events := plan.Taxonomy.Events
	resolved := make([]ResolvedEvent, 0, len(events))

	for name, event := range events {
		props, err := plan.Taxonomy.FlattenEvent(name)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, ResolvedEvent{
			Name:        name,
			Description: event.Description,
			Properties:  props,
		})
	}

	// Sort by Name for stability
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Name < resolved[j].Name
	})

	return resolved, nil
}

// getSortedPropertyNames returns the keys of the property map sorted alphabetically.
func getSortedPropertyNames(props map[string]ast.PropertyV2) []string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
