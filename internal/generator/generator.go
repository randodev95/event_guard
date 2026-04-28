package generator

import (
	"sort"
	"github.com/randodev95/eventcanvas/pkg/ast"
)

// ResolvedEvent is the common logical representation of an event after inheritance resolution.
type ResolvedEvent struct {
	Name       string
	Category   string
	EntityType string
	Properties map[string]ast.Property
}

// getResolvedEvents returns a sorted list of events with their properties resolved.
// This ensures stable output across all generators.
func getResolvedEvents(plan *ast.TrackingPlan) ([]ResolvedEvent, error) {
	type result struct {
		event ResolvedEvent
		err   error
	}
	results := make(chan result, len(plan.Events))
	
	for name, event := range plan.Events {
		go func(n string, e ast.Event) {
			props, err := plan.ResolveProperties(n)
			if err != nil {
				results <- result{err: err}
				return
			}
			results <- result{
				event: ResolvedEvent{
					Name:       n,
					Category:   e.Category,
					EntityType: e.EntityType,
					Properties: props,
				},
			}
		}(name, event)
	}

	resolved := make([]ResolvedEvent, 0, len(plan.Events))
	for i := 0; i < len(plan.Events); i++ {
		res := <-results
		if res.err != nil {
			return nil, res.err
		}
		resolved = append(resolved, res.event)
	}

	// Sort by Name for stability
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Name < resolved[j].Name
	})

	return resolved, nil
}

// getSortedPropertyNames returns the keys of the property map sorted alphabetically.
func getSortedPropertyNames(props map[string]ast.Property) []string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
