package postgres_test

import (
	"testing"

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

func TestBindings(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	for _, b := range []ingest.Binding{{Instance: "xtb-main", Account: "acc-xtb"}, {Instance: "ibkr-main", Account: "acc-ibkr"}, {Instance: "ibkr-main", Account: "acc-cash"}} {
		if err := s.Bind(ctx, b.Instance, b.Account, t0()); err != nil {
			t.Fatalf("Bind: %v", err)
		}
	}
	got, err := s.Bindings(ctx)
	want := []ingest.Binding{{Instance: "ibkr-main", Account: "acc-cash"}, {Instance: "ibkr-main", Account: "acc-ibkr"}, {Instance: "xtb-main", Account: "acc-xtb"}}
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("Bindings = %v (-want +got):\n%s", err, diff)
	}
}
