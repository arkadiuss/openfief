package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// TestSetSeriesSource checks the owner's switch of one series' source
// (D-94): the owner may pick any source with a value of the series; a
// change revalues, from their start, the accounts that held the instrument,
// or every account for an FX rate or an index; the same source again
// revalues nothing; anything else is refused and changes nothing.
func TestSetSeriesSource(t *testing.T) {
	t.Parallel()
	etf := valuation.PriceSeries(instrument(t, etfISIN, "SYNTH").String())
	usd := valuation.FXSeries(money(t, "1", "USD").Currency())
	tests := []struct {
		name       string
		token      string
		series     valuation.Series
		source     string
		wantErr    error
		wantRevals []valuation.Revaluation
	}{
		{
			name: "an instrument's prices revalue only the account that held it", token: ownerToken, series: etf, source: "vendor-b",
			wantRevals: revals(t, "acc-1", "2026-03-02"),
		},
		{
			name: "an FX rate revalues every account", token: ownerToken, series: usd, source: "ecb-fx",
			wantRevals: revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05"),
		},
		{name: "the same source again revalues nothing", token: ownerToken, series: etf, source: "eodhd"},
		{name: "a source without a value of the series", token: ownerToken, series: etf, source: "nbp-a", wantErr: app.ErrInvalidInput},
		{
			name: "not a series", token: ownerToken, series: valuation.Series{Kind: valuation.KindFXRates, Key: "pln"}, source: "nbp-a",
			wantErr: valuation.ErrInvalidSeries,
		},
		{name: "not the owner", token: "ofsess_read", series: etf, source: "vendor-b", wantErr: app.ErrNotOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeRankings{
				series: []valuation.SeriesSource{{Series: etf, Source: "eodhd"}, {Series: usd, Source: "nbp-a"}},
				available: map[valuation.Series][]string{
					etf: {"eodhd", "vendor-b"}, usd: {"ecb-fx", "nbp-a"},
				},
			}
			err := newRankings(t, store).SetSeriesSource(t.Context(), "https://openfief.example", tt.token, tt.series, tt.source)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRevals, store.revals); diff != "" {
				t.Errorf("revaluations (-want +got):\n%s", diff)
			}
			if tt.wantErr == nil {
				checkOwnersSource(t, newRankings(t, store), tt.series, tt.source)
			}
		})
	}
	if _, err := newRankings(t, &fakeRankings{}).SeriesSources(t.Context(), "https://openfief.example", "ofsess_read"); err == nil {
		t.Error("SeriesSources by another session succeeded, want ErrNotOwner")
	}
}

// checkOwnersSource checks that r lists source as the owner's choice for
// series.
func checkOwnersSource(t *testing.T, r *app.Rankings, series valuation.Series, source string) {
	t.Helper()
	got, err := r.SeriesSources(t.Context(), "https://openfief.example", ownerToken)
	if err != nil {
		t.Fatalf("SeriesSources: %v", err)
	}
	for _, ss := range got {
		if ss.Series == series && (ss.Source != source || !ss.ByOwner) {
			t.Errorf("source of %s = %+v, want %s chosen by the owner", series, ss, source)
		}
	}
}

// TestSetGapThreshold checks the owner's gap threshold of a kind (D-94):
// it is stored and shown with the kind's ranking; an invalid one or
// another session is refused.
func TestSetGapThreshold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		kind    valuation.MarketDataKind
		days    int
		wantErr error
		want    int
	}{
		{name: "the owner sets 14 days for prices", token: ownerToken, kind: valuation.KindPrices, days: 14, want: 14},
		{name: "zero days", token: ownerToken, kind: valuation.KindPrices, wantErr: app.ErrInvalidInput, want: 7},
		{name: "no kind", token: ownerToken, days: 5, wantErr: valuation.ErrUnknownMarketDataKind, want: 7},
		{name: "not the owner", token: "ofsess_read", kind: valuation.KindPrices, days: 3, wantErr: app.ErrNotOwner, want: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeRankings{}
			r := newRankings(t, store)
			err := r.SetGapThreshold(t.Context(), "https://openfief.example", tt.token, tt.kind, tt.days)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			all, err := r.SourceRankings(t.Context(), "https://openfief.example", ownerToken)
			if err != nil || all[0].GapThreshold != tt.want {
				t.Errorf("price gap threshold = %v, %v; want %d", all, err, tt.want)
			}
		})
	}
}

// TestRevalueRaisesGaps checks D-94's gap rule through a revaluation: a
// series a holding needs today whose source has had no value for longer
// than its kind's threshold raises an attention item for the account and
// series, naming the source and its last value's date; every other series
// read is cleared, so the item resolves once values resume.
func TestRevalueRaisesGaps(t *testing.T) {
	t.Parallel()
	etf := valuation.PriceSeries(instrument(t, etfISIN, "SYNTH").String())
	tests := []struct {
		name        string
		lastPrice   string
		gaps        map[valuation.MarketDataKind]int
		wantRaised  []string
		wantCleared []string
	}{
		{
			name: "a price gap of 9 days, over the default 7", lastPrice: "2026-03-03",
			wantRaised:  []string{app.GapKind + " acc-1/" + etf.String() + ": no value from vendor since 2026-03-03"},
			wantCleared: []string{app.GapKind + " acc-1/fx_rates/USD"},
		},
		{
			name: "the owner's threshold of 10 days covers it", lastPrice: "2026-03-03",
			gaps:        map[valuation.MarketDataKind]int{valuation.KindPrices: 10},
			wantCleared: []string{app.GapKind + " acc-1/" + etf.String(), app.GapKind + " acc-1/fx_rates/USD"},
		},
		{
			name: "values resumed", lastPrice: "2026-03-12",
			wantCleared: []string{app.GapKind + " acc-1/" + etf.String(), app.GapKind + " acc-1/fx_rates/USD"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newFakeStore()
			s.prices = []valuation.Price{usdPrice(t, "2026-03-02", "100"), usdPrice(t, tt.lastPrice, "101.5")}
			s.rates = []valuation.FXRate{rate(t, "USD", "2026-03-02", "4.0000"), rate(t, "USD", "2026-03-11", "4.0100")}
			s.gaps = valuation.GapThresholds{Days: tt.gaps}
			attention := &fakeAttention{only: app.GapKind}
			r := revaluer(t, etfLedger(t), s, "2026-03-12").WithAttention(attention)
			if _, err := r.Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
				t.Fatalf("Revalue: %v", err)
			}
			if diff := cmp.Diff([][]string{tt.wantRaised, tt.wantCleared}, [][]string{attention.raised, attention.cleared}); diff != "" {
				t.Errorf("raised, cleared (-want +got):\n%s", diff)
			}
		})
	}
}

// TestNeededSeries checks which series the holdings valued on a day need:
// a holding at market its prices and, in another currency than PLN, its
// FX rates; a bond its type's period index; nothing for a fixed-rate bond,
// a holding without a price's currency, or another day.
func TestNeededSeries(t *testing.T) {
	t.Parallel()
	day := date(t, "2026-03-12")
	etf := instrument(t, etfISIN, "SYNTH")
	edo := instrument(t, bondISIN, "EDO0336")
	ots := instrument(t, otherISIN, "OTS0626")
	rows := []valuation.HoldingValuation{
		{Instrument: etf, Date: date(t, "2026-03-11"), Basis: valuation.Market, Value: money(t, "10", "EUR")},
		{Instrument: etf, Date: day, Basis: valuation.Market, Value: money(t, "10", "USD")},
		{Instrument: instrument(t, otherISIN, "PLSYNT"), Date: day, Basis: valuation.Market, Value: money(t, "10", "PLN")},
		{Instrument: instrument(t, otherISIN, "NOPRICE"), Date: day, Basis: valuation.Market},
		{Instrument: edo, Date: day, Basis: valuation.GrossAccrued},
		{Instrument: ots, Date: day, Basis: valuation.GrossAccrued},
	}
	indexation := func(isin ledger.ISIN) pl.Indexation {
		if isin.Equal(edo.ISIN) {
			return pl.Inflation
		}
		return pl.Fixed
	}
	want := []valuation.Series{
		valuation.PriceSeries(etf.String()), valuation.FXSeries(money(t, "1", "USD").Currency()),
		valuation.PriceSeries(instrument(t, otherISIN, "PLSYNT").String()),
		valuation.PriceSeries(instrument(t, otherISIN, "NOPRICE").String()),
		valuation.IndexSeries(valuation.CPI),
	}
	if diff := cmp.Diff(want, app.NeededSeries(rows, day, money(t, "0", "PLN").Currency(), indexation)); diff != "" {
		t.Errorf("NeededSeries (-want +got):\n%s", diff)
	}
}
