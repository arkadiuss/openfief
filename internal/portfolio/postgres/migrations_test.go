package postgres

import "testing"

// TestEmbeddedMigrationsLoad checks, without a database, that every embedded
// migration has a version number and no two share one, so a merge that adds
// a migration with a taken number fails here instead of at the core's start.
func TestEmbeddedMigrationsLoad(t *testing.T) {
	t.Parallel()

	got, err := migrations()
	if err != nil {
		t.Fatalf("loading the embedded migrations: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("no embedded migrations")
	}
	for i, m := range got {
		if m.version != i+1 {
			t.Errorf("migration %s has version %d, want %d: versions must run 1, 2, 3 with no gap", m.name, m.version, i+1)
		}
	}
}
