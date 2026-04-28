package storage

import (
	"database/sql"
	"io"
	"os"

	_ "modernc.org/sqlite"
)

// DB wraps the underlying SQL database connection.
type DB struct {
	*sql.DB
}

// Snapshot represents a collection of event payloads associated with a specific git SHA.
type Snapshot struct {
	SHA       string
	EventName string
	Payloads  []string
}

// NewSQLiteDB initializes a new SQLite database at the given path.
func NewSQLiteDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sha TEXT,
		event_name TEXT,
		payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// SaveSnapshot persists an event snapshot to the database.
func (db *DB) SaveSnapshot(s Snapshot) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, payload := range s.Payloads {
		_, err := tx.Exec("INSERT INTO snapshots (sha, event_name, payload) VALUES (?, ?, ?)", s.SHA, s.EventName, payload)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSnapshots retrieves all event snapshots for a specific git SHA.
func (db *DB) GetSnapshots(sha string) ([]Snapshot, error) {
	rows, err := db.Query("SELECT event_name, payload FROM snapshots WHERE sha = ?", sha)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshotMap := make(map[string]*Snapshot)

	for rows.Next() {
		var eventName, payload string
		if err := rows.Scan(&eventName, &payload); err != nil {
			return nil, err
		}

		if _, ok := snapshotMap[eventName]; !ok {
			snapshotMap[eventName] = &Snapshot{
				SHA:       sha,
				EventName: eventName,
				Payloads:  []string{},
			}
		}
		snapshotMap[eventName].Payloads = append(snapshotMap[eventName].Payloads, payload)
	}

	var result []Snapshot
	for _, s := range snapshotMap {
		result = append(result, *s)
	}

	return result, nil
}

// Migrate performs a shadow copy migration to ensure zero-downtime updates to the local database.
func Migrate(path string, migrationSQL string) error {
	tmpPath := path + ".tmp"

	// 1. Shadow Copy
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	// 2. Migrate the copy
	tmpDB, err := sql.Open("sqlite", tmpPath)
	if err != nil {
		return err
	}
	defer tmpDB.Close()

	if _, err := tmpDB.Exec(migrationSQL); err != nil {
		return err
	}
	tmpDB.Close() // Close before renaming

	// 3. Atomic Rename
	return os.Rename(tmpPath, path)
}
