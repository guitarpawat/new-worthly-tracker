package db

import "testing"

func TestOpen_UsesInMemoryDatabaseWhenPathIsEmpty(t *testing.T) {
	t.Parallel()

	database, err := Open(SQLiteConfig{})
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	var databaseFile string
	if err := database.Get(&databaseFile, "SELECT file FROM pragma_database_list WHERE name = 'main'"); err != nil {
		t.Fatalf("select database file returned error: %v", err)
	}
	if databaseFile != "" {
		t.Fatalf("expected in-memory database file to be empty, got %q", databaseFile)
	}
	if maxOpenConnections := database.Stats().MaxOpenConnections; maxOpenConnections != 1 {
		t.Fatalf("expected one SQLite connection, got %d", maxOpenConnections)
	}

	if _, err := database.Exec(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create parent table: %v", err)
	}
	if _, err := database.Exec(`
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,
			FOREIGN KEY (parent_id) REFERENCES parent(id)
		)
	`); err != nil {
		t.Fatalf("create child table: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO child (id, parent_id) VALUES (1, 999)`); err == nil {
		t.Fatal("expected foreign key constraint to reject missing parent")
	}
}
