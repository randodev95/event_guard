package ast

import "sort"

// FlowLineage represents the upstream and downstream connections for a business flow.
type FlowLineage struct {
	FlowID     string
	Upstream   []string
	Downstream []string
}

// GetLineage calculates the cross-flow connections based on shared (State, Event) steps.
func (p *TrackingPlan) GetLineage() map[string]*FlowLineage {
	lineage := make(map[string]*FlowLineage)
	for _, flow := range p.Flows {
		lineage[flow.ID] = &FlowLineage{FlowID: flow.ID}
	}

	// nodeToFlows maps a (State, Event) pair to a list of flows that contain it.
	type stepKey struct {
		State string
		Event string
	}
	nodeToFlows := make(map[stepKey][]string)

	for _, flow := range p.Flows {
		for _, step := range flow.Steps {
			key := stepKey{State: step.State, Event: step.Event}
			nodeToFlows[key] = append(nodeToFlows[key], flow.ID)
		}
	}

	// Link flows that share nodes
	for _, flow := range p.Flows {
		if len(flow.Steps) == 0 {
			continue
		}

		// First step can be a downstream of any flow that ends with or contains this node
		firstStep := flow.Steps[0]
		firstKey := stepKey{State: firstStep.State, Event: firstStep.Event}
		
		// Last step can be an upstream of any flow that starts with or contains this node
		lastStep := flow.Steps[len(flow.Steps)-1]
		lastKey := stepKey{State: lastStep.State, Event: lastStep.Event}

		// Simple logic for tracer bullet:
		// If Flow A's last step is Flow B's first step, they are linked.
		for _, otherFlowID := range nodeToFlows[firstKey] {
			if otherFlowID != flow.ID {
				// flow starts where otherFlow contains this node
				// For the tracer bullet, let's just check if it's the LAST step of the other flow
				if p.isLastStep(otherFlowID, firstKey) {
					p.addLink(lineage, otherFlowID, flow.ID)
				}
			}
		}

		for _, otherFlowID := range nodeToFlows[lastKey] {
			if otherFlowID != flow.ID {
				// flow ends where otherFlow contains this node
				if p.isFirstStep(otherFlowID, lastKey) {
					p.addLink(lineage, flow.ID, otherFlowID)
				}
			}
		}

		// 3. Explicit Lineage
		for _, upID := range flow.UpstreamFlows {
			p.addLink(lineage, upID, flow.ID)
		}
	}

	return lineage
}

func (p *TrackingPlan) isFirstStep(flowID string, key struct{State string; Event string}) bool {
	for _, f := range p.Flows {
		if f.ID == flowID && len(f.Steps) > 0 {
			return f.Steps[0].State == key.State && f.Steps[0].Event == key.Event
		}
	}
	return false
}

func (p *TrackingPlan) isLastStep(flowID string, key struct{State string; Event string}) bool {
	for _, f := range p.Flows {
		if f.ID == flowID && len(f.Steps) > 0 {
			last := f.Steps[len(f.Steps)-1]
			return last.State == key.State && last.Event == key.Event
		}
	}
	return false
}

func (p *TrackingPlan) addLink(lineage map[string]*FlowLineage, upstreamID, downstreamID string) {
	u := lineage[upstreamID]
	d := lineage[downstreamID]

	// Add to upstream's downstream list
	if !sliceContains(u.Downstream, downstreamID) {
		u.Downstream = append(u.Downstream, downstreamID)
		sort.Strings(u.Downstream)
	}

	// Add to downstream's upstream list
	if !sliceContains(d.Upstream, upstreamID) {
		d.Upstream = append(d.Upstream, upstreamID)
		sort.Strings(d.Upstream)
	}
}

func sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// stepKey uniquely identifies a node in the tracking graph.
type stepKey struct {
	State string
	Event string
}

// DiscoverFlows identifies unique maximal journeys by merging overlapping flows.
func (p *TrackingPlan) DiscoverFlows() []Flow {

	// adj maps a node to its possible next steps
	adj := make(map[stepKey]map[stepKey]bool)
	// nodes stores all unique nodes
	nodes := make(map[stepKey]bool)
	// inDegree tracks incoming edges for source detection
	inDegree := make(map[stepKey]int)

	for _, flow := range p.Flows {
		for i := 0; i < len(flow.Steps); i++ {
			u := stepKey{State: flow.Steps[i].State, Event: flow.Steps[i].Event}
			nodes[u] = true
			if i < len(flow.Steps)-1 {
				v := stepKey{State: flow.Steps[i+1].State, Event: flow.Steps[i+1].Event}
				if adj[u] == nil {
					adj[u] = make(map[stepKey]bool)
				}
				if !adj[u][v] {
					adj[u][v] = true
					inDegree[v]++
				}
			}
		}
	}

	// 2. Incorporate Event-level triggers
	for eventName, event := range p.Events {
		for _, trigger := range event.Triggers {
			u := stepKey{State: trigger.FromState, Event: "UNKNOWN"} // Initial state often has no prior event
			v := stepKey{State: trigger.FromState, Event: eventName}
			nodes[u] = true
			nodes[v] = true
			if adj[u] == nil {
				adj[u] = make(map[stepKey]bool)
			}
			if !adj[u][v] {
				adj[u][v] = true
				inDegree[v]++
			}
		}
	}

	var maximalFlows []Flow
	visited := make(map[stepKey]bool)

	// Start DFS from all source nodes (in-degree 0)
	for node := range nodes {
		if inDegree[node] == 0 {
			paths := p.dfsPaths(node, adj, nil)
			for i, path := range paths {
				maximalFlows = append(maximalFlows, Flow{
					ID:    "discovered_" + node.State + "_" + string(rune(i)),
					Name:  "Discovered Flow from " + node.State,
					Steps: path,
				})
			}
			visited[node] = true
		}
	}

	return maximalFlows
}

func (p *TrackingPlan) dfsPaths(u stepKey, adj map[stepKey]map[stepKey]bool, currentPath []FlowStep) [][]FlowStep {
	step := FlowStep{State: u.State, Event: u.Event}
	currentPath = append(currentPath, step)

	if len(adj[u]) == 0 {
		return [][]FlowStep{currentPath}
	}

	var allPaths [][]FlowStep
	for v := range adj[u] {
		// Avoid cycles in a single path
		if p.isNodeInPath(v, currentPath) {
			allPaths = append(allPaths, currentPath)
			continue
		}
		subPaths := p.dfsPaths(v, adj, append([]FlowStep(nil), currentPath...))
		allPaths = append(allPaths, subPaths...)
	}
	return allPaths
}

func (p *TrackingPlan) isNodeInPath(node stepKey, path []FlowStep) bool {
	for _, s := range path {
		if s.State == node.State && s.Event == node.Event {
			return true
		}
	}
	return false
}
