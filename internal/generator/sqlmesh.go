package generator

import (
	"fmt"
	"strings"
	"github.com/randodev95/event_guard/pkg/ast"
)

func GenerateSQLMesh(plan *ast.TrackingPlan) (string, error) {
	resolved, err := getResolvedEvents(plan)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	for _, event := range resolved {
		safeName := strings.ToLower(strings.ReplaceAll(event.Name, " ", "_"))
		sb.WriteString(fmt.Sprintf("MODEL (\n  name analytics.%s,\n  kind FULL,\n  cron '@daily'\n);\n\n", safeName))
		sb.WriteString("SELECT\n")
		
		var cols []string
		propNames := getSortedPropertyNames(event.Properties)
		for _, propName := range propNames {
			prop := event.Properties[propName]
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
