package generator

import (
	"fmt"
	"strings"
	"github.com/eventcanvas/eventcanvas/pkg/ast"
)

func GenerateDBT(plan *ast.TrackingPlan) (string, error) {
	var sb strings.Builder
	sb.WriteString("version: 2\n\nmodels:\n")

	for eventName, event := range plan.Events {
		sb.WriteString(fmt.Sprintf("  - name: %s\n", eventName))
		sb.WriteString("    columns:\n")
		for propName, prop := range event.Properties {
			sb.WriteString(fmt.Sprintf("      - name: %s\n", propName))
			if prop.Required {
				sb.WriteString("        tests:\n")
				sb.WriteString("          - dbt_expectations.expect_column_to_exist\n")
				sb.WriteString("          - not_null\n")
			}
		}
	}

	return sb.String(), nil
}
