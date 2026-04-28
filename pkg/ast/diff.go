package ast

import (
	"fmt"
)

// DiffPlans compares two tracking plans and returns a list of breaking changes.
func DiffPlans(old, new *TrackingPlan) []string {
	var breaches []string

	// 1. Check for removed events
	for name := range old.Events {
		if _, exists := new.Events[name]; !exists {
			breaches = append(breaches, fmt.Sprintf("Event [%s] was removed", name))
		}
	}

	// 2. Check for changes in existing events
	for name, oldEvent := range old.Events {
		newEvent, exists := new.Events[name]
		if !exists {
			continue
		}

		// Check for removed properties
		for propName, oldProp := range oldEvent.Properties {
			newProp, propExists := newEvent.Properties[propName]
			if !propExists {
				breaches = append(breaches, fmt.Sprintf("Event [%s]: Property [%s] was removed", name, propName))
				continue
			}

			// Check for type changes
			if oldProp.Type != newProp.Type {
				breaches = append(breaches, fmt.Sprintf("Event [%s]: Property [%s] type changed from %s to %s", name, propName, oldProp.Type, newProp.Type))
			}

			// Check for requirement changes (making optional property required)
			if !oldProp.Required && newProp.Required {
				breaches = append(breaches, fmt.Sprintf("Event [%s]: Property [%s] became required", name, propName))
			}
		}
	}

	return breaches
}
