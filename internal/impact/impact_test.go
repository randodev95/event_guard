package impact

import (
	"testing"
	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/internal/storage"
	"os"
)

func TestCheckParity(t *testing.T) {
	dbPath := "test_impact.db"
	db, _ := storage.NewSQLiteDB(dbPath)
	defer os.Remove(dbPath)

	// 1. Save snapshot
	sha := "old-sha"
	db.SaveSnapshot(storage.Snapshot{
		SHA:       sha,
		EventName: "Login",
		Payloads:  []string{`{"event": "Login", "properties": {"user": "alice"}}`},
	})

	// 2. Test with valid plan
	plan := &ast.TrackingPlan{
		Taxonomy: ast.Taxonomy{
			Events: map[string]ast.EventV2{
				"Login": {
					Properties: map[string]ast.PropertyV2{
						"user": {Type: "string"},
					},
				},
			},
		},
	}

	breaches, err := CheckParity(db, sha, plan)
	if err != nil {
		t.Fatalf("CheckParity failed: %v", err)
	}
	if len(breaches) > 0 {
		t.Errorf("Expected 0 breaches, got %d", len(breaches))
	}

	// 3. Test with breaking change (added a new required property)
	planBroken := &ast.TrackingPlan{
		Taxonomy: ast.Taxonomy{
			Events: map[string]ast.EventV2{
				"Login": {
					Properties: map[string]ast.PropertyV2{
						"user": {Type: "string"},
						"new_req": {Type: "string", Required: true},
					},
				},
			},
		},
	}

	breaches, err = CheckParity(db, sha, planBroken)
	if err != nil {
		t.Fatalf("CheckParity failed: %v", err)
	}
	if len(breaches) == 0 {
		t.Error("Expected breach for new required property, got none")
	}
}
