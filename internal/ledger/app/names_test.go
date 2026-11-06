package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// memNames is an in-memory InstrumentNameStore with the stores' rule: an
// owner's name always, a connector's only while none of its instrument is
// stored.
type memNames struct {
	mu      sync.Mutex
	records []ledger.InstrumentName
}

func (m *memNames) AppendInstrumentNames(_ context.Context, names []ledger.InstrumentName) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, n := range names {
		delivered := false
		for _, r := range m.records {
			delivered = delivered || (!r.ByOwner && r.ISIN.Equal(n.ISIN))
		}
		if n.ByOwner || !delivered {
			m.records = append(m.records, n)
		}
	}
	return nil
}

func (m *memNames) InstrumentNames(context.Context) ([]ledger.InstrumentName, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ledger.InstrumentName(nil), m.records...), nil
}

// namesOver returns the names over a ledger whose acc-1 holds isinA,
// asking owners.
func namesOver(t *testing.T, owners app.Owners) (*app.Names, *memNames) {
	t.Helper()
	l, _ := seeded(t)
	buy := must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), instrument(t, isinA, "SYNA"), qty(t, "3"), money(t, "30", "PLN"), money(t, "0", "PLN")))
	if _, err := l.Import(t.Context(), []ledger.Transaction{buy}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	store := &memNames{}
	return &app.Names{Store: store, Owners: owners, Postings: l}, store
}

func TestNamesDeliverAndRename(t *testing.T) {
	t.Parallel()
	isin := instrument(t, isinA, "SYNA").ISIN
	n, _ := namesOver(t, fakeOwners{})
	ctx := t.Context()
	if err := n.Deliver(ctx, "xtb-main", []ledger.InstrumentName{{ISIN: isin, Name: " Synthetic  A "}, {ISIN: isin, Name: "Ignored"}}); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if err := n.Deliver(ctx, "ibkr-main", []ledger.InstrumentName{{ISIN: isin, Name: "SYNTHETIC A ETF"}}); err != nil {
		t.Fatalf("Deliver again: %v", err)
	}
	checkName(t, n, isin, "Synthetic A")
	got, err := n.Rename(ctx, "http://localhost", "owner", isin, "  My  A ")
	if err != nil || got != "My A" {
		t.Fatalf("Rename = %q, %v; want My A", got, err)
	}
	checkName(t, n, isin, "My A")
}

func TestNamesRenameRefusals(t *testing.T) {
	t.Parallel()
	isin := instrument(t, isinA, "SYNA").ISIN
	tests := []struct {
		name, token, newName string
		owners               fakeOwners
		want                 error
	}{
		{name: "not the owner", token: "reader", newName: "Mine", want: app.ErrNotOwner},
		{name: "an empty name", token: "owner", newName: "  ", want: ledger.ErrInvalidInstrumentName},
		{name: "an account the owner may not change", token: "owner", newName: "Mine", owners: fakeOwners{denied: []ledger.AccountID{acc1}}, want: app.ErrNotOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, store := namesOver(t, tt.owners)
			if _, err := n.Rename(t.Context(), "http://localhost", tt.token, isin, tt.newName); !errors.Is(err, tt.want) {
				t.Errorf("Rename = %v, want %v", err, tt.want)
			}
			if len(store.records) != 0 {
				t.Errorf("records = %v, want none", store.records)
			}
		})
	}
}

// checkName fails t unless n shows want for isin.
func checkName(t *testing.T, n *app.Names, isin ledger.ISIN, want string) {
	t.Helper()
	current, err := n.Current(t.Context())
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if diff := cmp.Diff(want, current[isin]); diff != "" {
		t.Errorf("name of %s (-want +got):\n%s", isin, diff)
	}
}
