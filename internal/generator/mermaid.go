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

	for flowName, flow := range plan.Flows {
		sb.WriteString(fmt.Sprintf("    subgraph %s [\"%s\"]\n", flowName, flow.Namespace))
		for nodeName, node := range flow.Nodes {
			// Node definition
			label := nodeName
			if node.Event != "" {
				label = fmt.Sprintf("%s<br/>(Trigger: %s)", nodeName, node.Event)
			} else if node.ListenFor != "" {
				label = fmt.Sprintf("%s<br/>(Listen: %s)", nodeName, node.ListenFor)
			}
			sb.WriteString(fmt.Sprintf("        %s_%s[\"%s\"]\n", flowName, nodeName, label))

			// Transitions
			for _, tr := range node.Transitions {
				sb.WriteString(fmt.Sprintf("        %s_%s --> %s_%s\n", flowName, nodeName, flowName, tr.Target))
			}
			// Conditions
			for _, cond := range node.Conditions {
				label := cond.If
				if cond.Default {
					label = "default"
				}
				sb.WriteString(fmt.Sprintf("        %s_%s -- \"%s\" --> %s_%s\n", flowName, nodeName, label, flowName, cond.Target))
			}
			// Timeout
			if node.Timeout != nil {
				sb.WriteString(fmt.Sprintf("        %s_%s -- \"timeout: %s\" --> %s_%s\n", flowName, nodeName, node.Timeout.Duration, flowName, node.Timeout.Target))
			}
		}
		sb.WriteString("    end\n")
	}

	return sb.String(), nil
}
