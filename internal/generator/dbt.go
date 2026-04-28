package generator

import (
	"fmt"
	"strings"

	"github.com/eventcanvas/eventcanvas/pkg/ast"
)

func GenerateDBT(plan *ast.TrackingPlan) (string, error) {
	resolved, err := getResolvedEvents(plan)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("version: 2\n\nmodels:\n")

	for _, event := range resolved {
		sb.WriteString(fmt.Sprintf("  - name: %s\n", event.Name))
		sb.WriteString("    columns:\n")
		
		propNames := getSortedPropertyNames(event.Properties)
		for _, propName := range propNames {
			prop := event.Properties[propName]
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
