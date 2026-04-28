package generator

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/eventcanvas/eventcanvas/pkg/parser"
)

func TestGenerateSQLMesh_Golden(t *testing.T) {
	yamlData := `
version: "1.0.0"
identity_properties: ["userId"]
events:
  "Order Completed":
    category: "INTERACTION"
    entity_type: "Transaction"
    properties:
      userId: { type: string, required: true }
      total: { type: number, required: true }
`
	plan, err := parser.ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse plan: %v", err)
	}

	got, err := GenerateSQLMesh(plan)
	if err != nil {
		t.Fatalf("GenerateSQLMesh failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "sqlmesh_order_completed.yaml")
	
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		_ = os.MkdirAll("testdata", 0755)
		err := os.WriteFile(goldenPath, []byte(got), 0644)
		if err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v. Run with UPDATE_GOLDEN=true to create it.", err)
	}

	if string(want) != got {
		t.Errorf("Output mismatch. Run with UPDATE_GOLDEN=true to update.\nGot:\n%s\nWant:\n%s", got, string(want))
	}
}
