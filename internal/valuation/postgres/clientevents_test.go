package postgres_test

import "testing"

// TestClientEventsSwitchRoundTrip stores the owner's switch for the UI's client
// events (D-227): with nothing stored they are on, turning them off is kept, and
// turning them on again removes the setting.
func TestClientEventsSwitchRoundTrip(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	off, err := store.ClientEventsOff(ctx)
	check(t, "off before any setting", false, off, err)
	for _, want := range []bool{true, true, false, false} {
		if saveErr := store.SaveClientEventsOff(ctx, want); saveErr != nil {
			t.Fatalf("SaveClientEventsOff(%v): %v", want, saveErr)
		}
		off, err = store.ClientEventsOff(ctx)
		check(t, "off after saving", want, off, err)
	}
	var rows int
	if scanErr := pool.QueryRow(ctx, `SELECT count(*) FROM valuation.settings WHERE key = 'client_events'`).Scan(&rows); scanErr != nil {
		t.Fatal(scanErr)
	}
	if rows != 0 {
		t.Errorf("%d client_events rows remain after turning them on, want none: on is the absence of the setting", rows)
	}
}
