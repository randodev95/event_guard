package generator

import (
	"fmt"
	"strings"
	"github.com/eventcanvas/eventcanvas/pkg/ast"
)

func GenerateSQLMesh(plan *ast.TrackingPlan) (string, error) {
	var sb strings.Builder

	for eventName, event := range plan.Events {
		sb.WriteString(fmt.Sprintf("MODEL (\n  name analytics.%s,\n  kind FULL,\n  cron '@daily'\n);\n\n", strings.ToLower(strings.ReplaceAll(eventName, " ", "_"))))
		sb.WriteString("SELECT\n")
		
		var cols []string
		for propName, prop := range event.Properties {
			sqlType := "TEXT"
			switch prop.Type {
			case "number":
				sqlType = "DOUBLE"
			case "integer":
				sqlType = "INT"
			case "boolean":
				sqlType = "BOOLEAN"
			}
			cols = append(cols, fmt.Sprintf("  CAST(%s AS %s) AS %s", propName, sqlType, propName))
		}
		sb.WriteString(strings.Join(cols, ",\n"))
		sb.WriteString("\nFROM raw_events;\n\n")
	}

	return sb.String(), nil
}
