package storage

import (
	"database/sql"
	"os"
	"testing"
)

func TestNewSQLiteDB(t *testing.T) {
	dbPath := "test_canvas.db"
	defer os.Remove(dbPath)

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite DB: %v", err)
	}
	defer db.Close()

	// Check if table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='snapshots'").Scan(&name)
	if err != nil {
		t.Errorf("Snapshots table does not exist: %v", err)
	}
}

func TestSaveAndGetSnapshots(t *testing.T) {
	dbPath := "test_snapshots.db"
	defer os.Remove(dbPath)

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite DB: %v", err)
	}
	defer db.Close()

	sha := "abc12345"
	snapshot := Snapshot{
		SHA:       sha,
		EventName: "Order Completed",
		Payloads:  []string{`{"total": 100}`, `{"total": 50}`},
	}

	err = db.SaveSnapshot(snapshot)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	snapshots, err := db.GetSnapshots(sha)
	if err != nil {
		t.Fatalf("GetSnapshots failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("Expected 1 snapshot, got %d", len(snapshots))
	}

	if snapshots[0].EventName != "Order Completed" {
		t.Errorf("Expected event 'Order Completed', got '%s'", snapshots[0].EventName)
	}

	if len(snapshots[0].Payloads) != 2 {
		t.Errorf("Expected 2 payloads, got %d", len(snapshots[0].Payloads))
	}
}

func TestMigration_ShadowCopy(t *testing.T) {
	dbPath := "test_migrate.db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + ".tmp")

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite DB: %v", err)
	}
	db.Close()

	// Define a migration (e.g., adding a new table)
	migrationSQL := "CREATE TABLE migrations (id INTEGER PRIMARY KEY);"

	err = Migrate(dbPath, migrationSQL)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify the new table exists in the original file (which should have been replaced)
	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open migrated DB: %v", err)
	}
	defer rawDB.Close()

	var name string
	err = rawDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&name)
	if err != nil {
		t.Errorf("Migrations table does not exist after migration: %v", err)
	}
}
