package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// errRankingsDown is a ranking store that fails.
var errRankingsDown = errors.New("store down")

// ownerToken is the only session fakeOwners takes as the owner's.
const ownerToken = "ofsess_owner"

// fakeOwners takes ownerToken and refuses every other session with err.
type fakeOwners struct{ err error }

func (f fakeOwners) CheckOwner(_ context.Context, _, token string) error {
	if token == ownerToken {
		return nil
	}
	return f.err
}

// fakeRankings keeps sources and the owner's order in memory, and records
// the revaluations each change enqueued.
type fakeRankings struct {
	sources []valuation.Source
	placed  map[valuation.MarketDataKind][]string
	revals  []valuation.Revaluation
	err     error
}

func (f *fakeRankings) SaveSources(_ context.Context, sources []valuation.Source) error {
	if f.err != nil {
		return f.err
	}
	f.sources = append(f.sources, sources...)
	return nil
}

func (f *fakeRankings) SourceRanking(_ context.Context, kind valuation.MarketDataKind) (valuation.Ranking, error) {
	if f.err != nil {
		return valuation.Ranking{}, f.err
	}
	return valuation.RankSources(kind, f.sources, f.placed[kind]), nil
}

func (f *fakeRankings) SaveSourceRanking(_ context.Context, kind valuation.MarketDataKind, placed []string, revals []valuation.Revaluation) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	if slices.Equal(f.placed[kind], placed) {
		return false, nil
	}
	if f.placed == nil {
		f.placed = map[valuation.MarketDataKind][]string{}
	}
	f.placed[kind], f.revals = placed, append(f.revals, revals...)
	return true, nil
}

// newRankings returns the ranking use cases over store and two accounts.
func newRankings(t *testing.T, store *fakeRankings) *app.Rankings {
	t.Helper()
	l := twoAccounts(t)
	return app.NewRankings(app.NewPlanner(l, l), store, fakeOwners{err: app.ErrNotOwner})
}

func TestRankSources(t *testing.T) {
	t.Parallel()
	store := &fakeRankings{}
	r := newRankings(t, store)
	ctx := t.Context()
	sources := []valuation.Source{
		{Kind: valuation.KindPrices, Name: "prices-eu", Institution: "eodhd"},
		{Kind: valuation.KindPrices, Name: "ibkr-main", Institution: "ibkr"},
	}
	if err := r.RecordSources(ctx, sources); err != nil {
		t.Fatalf("RecordSources: %v", err)
	}
	all, err := r.SourceRankings(ctx, "https://openfief.example", ownerToken)
	if err != nil || len(all) != 3 {
		t.Fatalf("SourceRankings = %v, %v; want one ranking per kind", all, err)
	}
	if got := all[0].Names(); !slices.Equal(got, []string{"prices-eu", "ibkr-main"}) {
		t.Errorf("default price ranking = %v, want EODHD first", got)
	}
	got, err := r.RankSources(ctx, "https://openfief.example", ownerToken, valuation.KindPrices, []string{"ibkr-main", "prices-eu"})
	if err != nil {
		t.Fatalf("RankSources: %v", err)
	}
	if names := got.Names(); !slices.Equal(names, []string{"ibkr-main", "prices-eu"}) {
		t.Errorf("ranking after the owner's change = %v", names)
	}
	if diff := cmp.Diff(revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05"), store.revals); diff != "" {
		t.Errorf("a change revalues every account from its start (-want +got):\n%s", diff)
	}
	store.revals = nil
	if _, err := r.RankSources(ctx, "https://openfief.example", ownerToken, valuation.KindPrices, []string{"ibkr-main", "prices-eu"}); err != nil || store.revals != nil {
		t.Errorf("the same ranking again = %v, revaluations %v; want none", err, store.revals)
	}
}

func TestRankSourcesRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		kind    valuation.MarketDataKind
		placed  []string
		sources []valuation.Source
		err     error
		wantErr error
	}{
		{name: "not the owner", token: "ofsess_read", kind: valuation.KindPrices, wantErr: app.ErrNotOwner},
		{name: "a source twice", token: ownerToken, kind: valuation.KindPrices, placed: []string{"a", "a"}, wantErr: valuation.ErrInvalidRanking},
		{name: "no kind", token: ownerToken, wantErr: valuation.ErrUnknownMarketDataKind},
		{name: "the store fails", token: ownerToken, kind: valuation.KindFXRates, err: errRankingsDown, wantErr: errRankingsDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeRankings{err: tt.err}
			r := newRankings(t, store)
			_, err := r.RankSources(t.Context(), "https://openfief.example", tt.token, tt.kind, tt.placed)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("RankSources error (-want +got):\n%s", diff)
			}
			if _, err := r.SourceRankings(t.Context(), "https://openfief.example", tt.token); tt.token != ownerToken && !errors.Is(err, app.ErrNotOwner) {
				t.Errorf("SourceRankings by another session = %v, want ErrNotOwner", err)
			}
		})
	}
}

func TestRecordSourcesRefuses(t *testing.T) {
	t.Parallel()
	store := &fakeRankings{}
	r := newRankings(t, store)
	err := r.RecordSources(t.Context(), []valuation.Source{{Kind: valuation.KindPrices, Name: ""}})
	if !errors.Is(err, app.ErrInvalidInput) || len(store.sources) != 0 {
		t.Errorf("RecordSources of a nameless source = %v, stored %v; want ErrInvalidInput and nothing", err, store.sources)
	}
	if err := r.RecordSources(t.Context(), nil); err != nil {
		t.Errorf("RecordSources of none = %v", err)
	}
}
