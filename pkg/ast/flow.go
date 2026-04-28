package ast

import "fmt"

// Flow represents a business‑level user journey.
// Each Flow is a directed graph of steps (states) that map to events.
// Analysts can use the flow ID to filter downstream data.
type Flow struct {
    ID    string     `yaml:"id"`
    Name  string     `yaml:"name"`
    Steps []FlowStep `yaml:"steps"`
}

type FlowStep struct {
    // State is the UI page or screen the user is on.
    State string `yaml:"state"`
    // Event is the tracking event that fires when the step is reached.
    Event string `yaml:"event"`
    // Triggers describe *how* the event can be arrived at.
    // e.g. DIRECT_LOAD, UI_NAVIGATION, UI_BACK, BROWSER_BACK, DEEP_LINK
    Triggers []string `yaml:"triggers"`
}

// ValidateFlows ensures each flow step references a defined event and allowed triggers.
func (p *TrackingPlan) ValidateFlows() error {
    // Allowed trigger identifiers – easy to extend later.
    allowed := map[string]bool{
        "DIRECT_LOAD":   true,
        "UI_NAVIGATION": true,
        "UI_BACK":       true,
        "BROWSER_BACK":  true,
        "DEEP_LINK":     true,
    }
    // Detect duplicate flow IDs
    ids := make(map[string]bool)
    for _, flow := range p.Flows {
        if ids[flow.ID] {
            return fmt.Errorf("duplicate flow id %s", flow.ID)
        }
        ids[flow.ID] = true
        if flow.ID == "" {
            return fmt.Errorf("flow missing id")
        }
        for i, step := range flow.Steps {
            if step.State == "" {
                return fmt.Errorf("flow %s step %d missing state", flow.ID, i)
            }
            if step.Event == "" {
                return fmt.Errorf("flow %s step %d missing event", flow.ID, i)
            }
            if _, ok := p.Events[step.Event]; !ok {
                return fmt.Errorf("flow %s step %d references unknown event %s", flow.ID, i, step.Event)
            }
            if len(step.Triggers) == 0 {
                return fmt.Errorf("flow %s step %d has no triggers", flow.ID, i)
            }
            for _, tr := range step.Triggers {
                if !allowed[tr] {
                    return fmt.Errorf("flow %s step %d has unknown trigger %s", flow.ID, i, tr)
                }
            }
        }
    }
    return nil
}
