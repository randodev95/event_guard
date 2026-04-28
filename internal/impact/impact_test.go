package impact

import (
	"os"
	"testing"
	"github.com/eventcanvas/eventcanvas/internal/storage"
	"github.com/eventcanvas/eventcanvas/pkg/ast"
)

func TestCheckParity_Success(t *testing.T) {
	dbPath := "test_impact.db"
	defer os.Remove(dbPath)

	db, err := storage.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite DB: %v", err)
	}
	defer db.Close()

	prevSHA := "old-sha"
	err = db.SaveSnapshot(storage.Snapshot{
		SHA:       prevSHA,
		EventName: "Order Completed",
		Payloads:  []string{`{"event": "Order Completed", "properties": {"total": 100}}`},
	})
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "number", Required: true},
				},
			},
		},
	}

	breaches, err := CheckParity(db, prevSHA, plan)
	if err != nil {
		t.Fatalf("CheckParity failed: %v", err)
	}

	if len(breaches) > 0 {
		t.Errorf("Expected 0 breaches, got %d: %v", len(breaches), breaches)
	}
}

func TestCheckParity_BreakingChange(t *testing.T) {
	dbPath := "test_impact_fail.db"
	defer os.Remove(dbPath)

	db, err := storage.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite DB: %v", err)
	}
	defer db.Close()

	prevSHA := "old-sha"
	// Old sample had "total" as a number
	err = db.SaveSnapshot(storage.Snapshot{
		SHA:       prevSHA,
		EventName: "Order Completed",
		Payloads:  []string{`{"event": "Order Completed", "properties": {"total": 100}}`},
	})
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// New plan changes "total" to a string
	plan := &ast.TrackingPlan{
		Events: map[string]ast.Event{
			"Order Completed": {
				Properties: map[string]ast.Property{
					"total": {Type: "string", Required: true},
				},
			},
		},
	}

	breaches, err := CheckParity(db, prevSHA, plan)
	if err != nil {
		t.Fatalf("CheckParity failed: %v", err)
	}

	if len(breaches) == 0 {
		t.Errorf("Expected at least one breach for type mismatch")
	}
}
