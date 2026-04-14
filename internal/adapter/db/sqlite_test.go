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
}
