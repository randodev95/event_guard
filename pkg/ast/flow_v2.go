package ast

import "fmt"

// Validate ensures graph integrity (no orphaned targets).
func (f *FlowPlan) Validate() error {
	for name, node := range f.Nodes {
		// Check Transitions
		for _, tr := range node.Transitions {
			if _, ok := f.Nodes[tr.Target]; !ok {
				return fmt.Errorf("node %s: orphaned transition target %s", name, tr.Target)
			}
		}
		// Check Conditions
		for _, cond := range node.Conditions {
			if _, ok := f.Nodes[cond.Target]; !ok {
				return fmt.Errorf("node %s: orphaned condition target %s", name, cond.Target)
			}
		}
		// Check Timeout
		if node.Timeout != nil {
			if _, ok := f.Nodes[node.Timeout.Target]; !ok {
				return fmt.Errorf("node %s: orphaned timeout target %s", name, node.Timeout.Target)
			}
		}
	}
	return nil
}

// FlowPlan defines the directed graph for user journeys.
type FlowPlan struct {
	Version   string          `yaml:"version"`
	Namespace string          `yaml:"namespace"`
	Nodes     map[string]Node `yaml:"nodes"`
}

// Node represents a step in the flow.
type Node struct {
	Type        string           `yaml:"type"`
	Description string           `yaml:"description,omitempty"`
	Event       string           `yaml:"event,omitempty"`       // For TriggerNode
	ListenFor   string           `yaml:"listen_for,omitempty"`  // For WaitNode
	Timeout     *TimeoutConfig   `yaml:"timeout,omitempty"`     // For WaitNode
	Transitions []Transition     `yaml:"transitions,omitempty"`
	Conditions  []Condition      `yaml:"conditions,omitempty"`  // For GateNode
	ActionType  string           `yaml:"action_type,omitempty"` // For TerminalNode
	Payload     map[string]interface{} `yaml:"payload,omitempty"`     // For TerminalNode
	Properties  map[string]PropertyV2  `yaml:"properties,omitempty"`  // Inline definition for extraction
}

// TimeoutConfig defines the timeout behavior for WaitNode.
type TimeoutConfig struct {
	Duration string `yaml:"duration"`
	Target   string `yaml:"target"`
}

// Transition represents an edge in the graph.
type Transition struct {
	Target string `yaml:"target"`
}

// Condition defines a logical branch for GateNode.
type Condition struct {
	If      string `yaml:"if,omitempty"`
	Default bool   `yaml:"default,omitempty"`
	Target  string `yaml:"target"`
}
