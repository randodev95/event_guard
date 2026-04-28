package generator

import (
	"strings"
	"testing"
	"github.com/randodev95/eventcanvas/pkg/ast"
)

func TestGenerateDBT(t *testing.T) {
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "number", Required: true},
				},
			},
		},
	}

	out, err := GenerateDBT(plan)
	if err != nil {
		t.Fatalf("GenerateDBT failed: %v", err)
	}

	if !strings.Contains(out, "name: Order Completed") {
		t.Errorf("dbt output missing event name")
	}

	if !strings.Contains(out, "name: total") {
		t.Errorf("dbt output missing property name")
	}

	if !strings.Contains(out, "dbt_expectations.expect_column_to_exist") {
		t.Errorf("dbt output missing dbt_expectations")
	}
}

func TestGenerateSQLMesh(t *testing.T) {
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "number", Required: true},
				},
			},
		},
	}

	out, err := GenerateSQLMesh(plan)
	if err != nil {
		t.Fatalf("GenerateSQLMesh failed: %v", err)
	}

	if !strings.Contains(out, "MODEL (") {
		t.Errorf("SQLMesh output missing MODEL header")
	}

	if !strings.Contains(out, "total") || !strings.Contains(out, "DOUBLE") {
		t.Errorf("SQLMesh output missing column definition for 'total' as DOUBLE")
	}
}

func TestGenerateHTML(t *testing.T) {
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "number", Required: true},
				},
			},
		},
	}

	out, err := GenerateHTML(plan)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	if !strings.Contains(out, "<html") {
		t.Errorf("HTML output missing <html tag")
	}

	if !strings.Contains(out, "Order Completed") {
		t.Errorf("HTML output missing event name")
	}
}
