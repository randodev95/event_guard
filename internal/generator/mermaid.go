package generator

import (
	"fmt"
	"github.com/randodev95/event_guard/pkg/ast"
	"strings"
)

// GenerateMermaid produces a Mermaid.js diagram representing the flows.
func GenerateMermaid(plan *ast.TrackingPlan) (string, error) {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// 1. Render Subgraphs
	for _, flow := range plan.Flows {
		sb.WriteString(fmt.Sprintf("    subgraph %s [\"%s\"]\n", flow.ID, flow.Name))
		for i := 0; i < len(flow.Steps); i++ {
			step := flow.Steps[i]
			// Node definition: Unique per flow to avoid Mermaid layout mess
			nodeID := fmt.Sprintf("%s_%s", flow.ID, strings.ReplaceAll(step.State, " ", "_"))
			sb.WriteString(fmt.Sprintf("        %s[\"%s<br/>(%s)\"]\n", nodeID, step.State, step.Event))

			// Connection within flow
			if i < len(flow.Steps)-1 {
				nextStep := flow.Steps[i+1]
				nextID := fmt.Sprintf("%s_%s", flow.ID, strings.ReplaceAll(nextStep.State, " ", "_"))
				triggers := strings.Join(step.Triggers, ", ")
				sb.WriteString(fmt.Sprintf("        %s -- \"%s\" --> %s\n", nodeID, triggers, nextID))
			}
		}
		sb.WriteString("    end\n")
	}

	// 2. Render Lineage (Cross-Flow Connections)
	lineage := plan.GetLineage()
	for flowID, l := range lineage {
		for _, downID := range l.Downstream {
			// Link the last node of upstream to the first node of downstream
			upFlow := getFlow(plan, flowID)
			downFlow := getFlow(plan, downID)
			
			if upFlow != nil && downFlow != nil && len(upFlow.Steps) > 0 && len(downFlow.Steps) > 0 {
				upNode := fmt.Sprintf("%s_%s", flowID, strings.ReplaceAll(upFlow.Steps[len(upFlow.Steps)-1].State, " ", "_"))
				downNode := fmt.Sprintf("%s_%s", downID, strings.ReplaceAll(downFlow.Steps[0].State, " ", "_"))
				sb.WriteString(fmt.Sprintf("    %s -. \"Lineage\" .-> %s\n", upNode, downNode))
			}
		}
	}

	return sb.String(), nil
}

func getFlow(p *ast.TrackingPlan, id string) *ast.Flow {
	for i := range p.Flows {
		if p.Flows[i].ID == id {
			return &p.Flows[i]
		}
	}
	return nil
}
