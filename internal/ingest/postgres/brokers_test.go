package postgres_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// day returns the date s, failing t on error.
func day(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// completeOpening runs and completes one opening batch id of instance at
// from, as a sync does.
func completeOpening(t *testing.T, s *postgres.Store, instance string, id ingest.BatchID, from time.Time) {
	t.Helper()
	claim(t, s, instance, from)
	if err := s.StartBatch(t.Context(), instance, app.OpenBatch{ID: id, Mode: ingest.ModeOpening}, from); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}
	b := &ingest.Batch{ID: id, Institution: "ibkr", Source: app.SourceSync, Mode: ingest.ModeOpening}
	if err := s.CompleteSync(t.Context(), instance, b, "", from.Add(time.Hour), from.Add(time.Minute)); err != nil {
		t.Fatalf("CompleteSync: %v", err)
	}
}

func TestBrokerAccountStore(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	const main, side ingest.BrokerAccountID = "U0000001", "U0000002"
	if err := s.RecordBrokerAccounts(ctx, "ibkr-main", "ibkr", []ingest.BrokerAccountID{main, side}, day(t, "2026-03-02")); err != nil {
		t.Fatalf("RecordBrokerAccounts: %v", err)
	}
	if err := s.RecordBrokerAccounts(ctx, "ibkr-main", "ibkr", []ingest.BrokerAccountID{main}, day(t, "2026-03-03")); err != nil {
		t.Fatalf("RecordBrokerAccounts again: %v", err)
	}
	checkMapping(t, s, main)
	want := []ingest.BrokerAccount{
		{Instance: "ibkr-main", ID: main, Institution: "ibkr", Account: "acc-ibkr", LastSeen: day(t, "2026-03-03"), InLastDelivery: true},
		{Instance: "ibkr-main", ID: side, Institution: "ibkr", LastSeen: day(t, "2026-03-02")},
	}
	got, err := s.BrokerAccounts(ctx, "ibkr-main")
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("BrokerAccounts = %v (-want +got):\n%s", err, diff)
	}
	if markErr := s.MarkSeeded(ctx, "ibkr-main", []ingest.BrokerAccountID{main, side}); markErr != nil {
		t.Fatalf("MarkSeeded: %v", markErr)
	}
	want[0].Seeded = true // an unmapped account is never seeded
	all, err := s.AllBrokerAccounts(ctx)
	if diff := cmp.Diff(want, all); err != nil || diff != "" {
		t.Errorf("AllBrokerAccounts = %v (-want +got):\n%s", err, diff)
	}
	if err := s.RecordBrokerAccounts(ctx, "ibkr-main", "ibkr", []ingest.BrokerAccountID{"U000 0003"}, day(t, "2026-03-04")); err == nil {
		t.Error("a broker account with a space was stored")
	}
}

// checkMapping maps ibkr-main's broker account id to acc-ibkr, twice, and
// checks that it can't be mapped to another account or for another
// instance.
func checkMapping(t *testing.T, s *postgres.Store, id ingest.BrokerAccountID) {
	t.Helper()
	ctx := t.Context()
	for range 2 {
		if err := s.MapBrokerAccount(ctx, "ibkr-main", id, "acc-ibkr", t0()); err != nil {
			t.Fatalf("MapBrokerAccount: %v", err)
		}
	}
	if err := s.MapBrokerAccount(ctx, "ibkr-main", id, "acc-other", t0()); !cmp.Equal(err, app.ErrBrokerAccountMapped, cmpopts.EquateErrors()) {
		t.Errorf("mapping to another account = %v, want ErrBrokerAccountMapped", err)
	}
	if err := s.MapBrokerAccount(ctx, "ibkr-other", id, "acc-ibkr", t0()); !cmp.Equal(err, app.ErrUnknownBrokerAccount, cmpopts.EquateErrors()) {
		t.Errorf("mapping one another instance reported = %v, want ErrUnknownBrokerAccount", err)
	}
}

// TestBrokerAccountsReseedTheInstance checks that an instance counts as
// seeded only while every broker account mapped for it was seeded: mapping
// one after the instance's opening batch makes the next sync an opening
// one again (D-72, D-23).
func TestBrokerAccountsReseedTheInstance(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	if err := s.RecordBrokerAccounts(ctx, "ibkr-main", "ibkr", []ingest.BrokerAccountID{"U0000001"}, day(t, "2026-03-02")); err != nil {
		t.Fatalf("RecordBrokerAccounts: %v", err)
	}
	completeOpening(t, s, "ibkr-main", "b-1", t0())
	if c := claim(t, s, "ibkr-main", t0().Add(time.Hour)); !c.Seeded {
		t.Errorf("claim = %+v, want seeded after an opening batch with nothing mapped", c)
	}
	if err := s.MapBrokerAccount(ctx, "ibkr-main", "U0000001", "acc-ibkr", t0()); err != nil {
		t.Fatalf("MapBrokerAccount: %v", err)
	}
	later := t0().Add(2 * time.Hour)
	if c := claim(t, s, "ibkr-main", later); c.Seeded {
		t.Errorf("claim = %+v, want not seeded once an unseeded broker account is mapped", c)
	}
	if err := s.MarkSeeded(ctx, "ibkr-main", []ingest.BrokerAccountID{"U0000001"}); err != nil {
		t.Fatalf("MarkSeeded: %v", err)
	}
	if c := claim(t, s, "ibkr-main", later.Add(time.Hour)); !c.Seeded {
		t.Errorf("claim = %+v, want seeded once the mapped account is seeded", c)
	}
}
