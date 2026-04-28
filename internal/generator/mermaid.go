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

	for _, flow := range plan.Flows {
		sb.WriteString(fmt.Sprintf("    subgraph %s [%s]\n", flow.ID, flow.Name))
		for i := 0; i < len(flow.Steps); i++ {
			step := flow.Steps[i]
			// Node definition
			stateNode := strings.ReplaceAll(step.State, " ", "_")
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", stateNode, step.State))

			// Connection to next step
			if i < len(flow.Steps)-1 {
				nextStep := flow.Steps[i+1]
				nextStateNode := strings.ReplaceAll(nextStep.State, " ", "_")
				triggers := strings.Join(step.Triggers, ", ")
				sb.WriteString(fmt.Sprintf("        %s -- \"%s: %s\" --> %s\n", stateNode, triggers, step.Event, nextStateNode))
			}
		}
		sb.WriteString("    end\n")
	}

	return sb.String(), nil
}
