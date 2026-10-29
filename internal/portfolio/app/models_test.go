package app_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/valuation"
)

const origin = "http://localhost:8080"

// sixtyForty is a valid model of two instruments, on.
func sixtyForty(t *testing.T) portfolio.Model {
	t.Helper()
	return portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		{Kind: portfolio.TargetInstrument, Name: etfIE, Weight: mustDec(t, "60")},
		{Kind: portfolio.TargetInstrument, Name: etfPL, Weight: mustDec(t, "40")},
	}}
}

func TestSettingsAndSave(t *testing.T) {
	t.Parallel()
	store := &fakeStore{}
	m := newModels(t, store, &fakeMarket{})
	ctx := t.Context()
	got, err := m.Settings(ctx, origin, owner)
	if err != nil || got.Enabled || len(got.Targets) != 0 {
		t.Fatalf("Settings before any save = %+v, %v; want off and empty", got, err)
	}
	want := sixtyForty(t)
	saved, err := m.Save(ctx, origin, owner, want)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if diff := cmp.Diff(want, saved); diff != "" {
		t.Errorf("Save returned (-want +got):\n%s", diff)
	}
	if got, err = m.Settings(ctx, origin, owner); err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Settings after Save (-want +got):\n%s", diff)
	}
	off := portfolio.Model{Targets: want.Targets}
	if _, err := m.Save(ctx, origin, owner, off); err != nil || store.model.Enabled {
		t.Errorf("turning it off: %v, stored %+v", err, store.model)
	}
}

func TestMarketDataTargets(t *testing.T) {
	t.Parallel()
	withIndex := sixtyForty(t)
	withIndex.Targets[1] = portfolio.Target{Kind: portfolio.TargetIndex, Name: "FAKE20", Currency: mustCurrency(t, "PLN"), Weight: mustDec(t, "40")}
	broken := sixtyForty(t)
	broken.Targets[0].Name = "XX0000000000/BAD"
	ie, err := app.ParseInstrument(etfIE)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		model       portfolio.Model
		storeErr    error
		wantIndices []string
		wantCount   int
		wantErr     bool
	}{
		{name: "off: nothing", model: portfolio.Model{Targets: sixtyForty(t).Targets}},
		{name: "on: instruments and indices", model: withIndex, wantIndices: []string{"FAKE20"}, wantCount: 1},
		{name: "a stored instrument that doesn't parse", model: broken, wantErr: true},
		{name: "the store fails", storeErr: errStoreDown, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newModels(t, &fakeStore{model: tt.model, err: tt.storeErr}, &fakeMarket{})
			instruments, indices, err := m.MarketDataTargets(t.Context())
			if (err != nil) != tt.wantErr {
				t.Fatalf("MarketDataTargets error = %v, want an error: %v", err, tt.wantErr)
			}
			if len(instruments) != tt.wantCount || (tt.wantCount == 1 && instruments[0] != ie) {
				t.Errorf("instruments = %v, want %d of %s", instruments, tt.wantCount, etfIE)
			}
			if diff := cmp.Diff(tt.wantIndices, indices); diff != "" {
				t.Errorf("indices (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOwnerOnly(t *testing.T) {
	t.Parallel()
	for _, refusal := range []error{app.ErrNoSession, app.ErrNotOwner, errStoreDown} {
		store := &fakeStore{model: sixtyForty(t)}
		m, err := app.New(store, &fakeMarket{}, fakeOwners{err: refusal}, mustCurrency(t, "PLN"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.Settings(t.Context(), origin, "ofsess_read"); !cmp.Equal(refusal, err, cmpopts.EquateErrors()) {
			t.Errorf("Settings with a refused session = %v, want %v", err, refusal)
		}
		if _, err := m.Save(t.Context(), origin, "ofsess_read", portfolio.Model{}); !cmp.Equal(refusal, err, cmpopts.EquateErrors()) {
			t.Errorf("Save with a refused session = %v, want %v", err, refusal)
		}
		if store.saves != 0 {
			t.Errorf("a refused session stored the model %d times", store.saves)
		}
	}
}

func TestSaveRejects(t *testing.T) {
	t.Parallel()
	inst := func(name, weight string) portfolio.Target {
		return portfolio.Target{Kind: portfolio.TargetInstrument, Name: name, Weight: mustDec(t, weight)}
	}
	tests := []struct {
		name    string
		targets []portfolio.Target
	}{
		{name: "weights summing to 90", targets: []portfolio.Target{inst(etfIE, "50"), inst(etfPL, "40")}},
		{name: "a wrong check digit", targets: []portfolio.Target{inst("IETEST000027/ETF1", "100")}},
		{name: "no symbol", targets: []portfolio.Target{inst("IETEST000026", "100")}},
		{name: "an empty symbol", targets: []portfolio.Target{inst("IETEST000026/", "100")}},
		{name: "a lower-case ISIN", targets: []portfolio.Target{inst("ietest000026/ETF1", "100")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeStore{}
			_, err := newModels(t, store, &fakeMarket{}).Save(t.Context(), origin, owner, portfolio.Model{Enabled: true, Targets: tt.targets})
			if !cmp.Equal(portfolio.ErrInvalidModel, err, cmpopts.EquateErrors()) {
				t.Errorf("Save = %v, want ErrInvalidModel", err)
			}
			if store.saves != 0 {
				t.Errorf("an invalid model was stored")
			}
		})
	}
	store := &fakeStore{err: errStoreDown}
	if _, err := newModels(t, store, &fakeMarket{}).Save(t.Context(), origin, owner, sixtyForty(t)); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("Save on a failing store = %v", err)
	}
}

func TestParseInstrument(t *testing.T) {
	t.Parallel()
	in, err := app.ParseInstrument(etfIE)
	if err != nil || in.String() != etfIE {
		t.Errorf("ParseInstrument(%q) = %v, %v", etfIE, in, err)
	}
}

func TestNewRefusesOtherReportingCurrencies(t *testing.T) {
	t.Parallel()
	_, err := app.New(&fakeStore{}, &fakeMarket{}, fakeOwners{}, mustCurrency(t, "EUR"))
	unsupported, ok := errors.AsType[*app.UnsupportedReportingCurrencyError](err)
	if !ok || unsupported.Error() != "reporting currency EUR is not supported: only PLN" {
		t.Errorf("New(EUR) = %v", err)
	}
}

// reportMarket is a USD ETF and a PLN index, with USD rates, in January
// 2025.
func reportMarket(t *testing.T) *fakeMarket {
	t.Helper()
	usd, pln := mustCurrency(t, "USD"), mustCurrency(t, "PLN")
	price := func(on, closing string) valuation.Price {
		return valuation.Price{Instrument: etfIE, Date: mustDate(t, on), Close: kernel.NewMoney(mustDec(t, closing), usd)}
	}
	level := func(on, value string) valuation.IndexLevel {
		return valuation.IndexLevel{Index: "FAKE20", Date: mustDate(t, on), Level: mustDec(t, value)}
	}
	return &fakeMarket{
		prices: map[string][]valuation.Price{etfIE: {price("2025-01-02", "10"), price("2025-01-03", "12"), price("2025-01-04", "9"), price("2025-01-09", "20")}},
		levels: map[string][]valuation.IndexLevel{"FAKE20": {level("2025-01-01", "1000"), level("2025-01-03", "1000"), level("2025-01-04", "1100")}},
		rates: map[kernel.Currency][]valuation.FXRate{
			usd: {{Currency: usd, Date: mustDate(t, "2025-01-01"), Rate: mustDec(t, "4")}},
			pln: {{Currency: usd, Date: mustDate(t, "2025-01-01"), Rate: mustDec(t, "999")}}, // never asked for
		},
	}
}

func TestReport(t *testing.T) {
	t.Parallel()
	model := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		{Kind: portfolio.TargetInstrument, Name: etfIE, Weight: mustDec(t, "50")},
		{Kind: portfolio.TargetIndex, Name: "FAKE20", Currency: mustCurrency(t, "PLN"), Weight: mustDec(t, "50")},
	}}
	got, on, err := newModels(t, &fakeStore{model: model}, reportMarket(t)).Report(t.Context(), mustDate(t, "2025-01-04"))
	if err != nil || !on {
		t.Fatalf("Report = %v, %v", on, err)
	}
	// 1.25 units of the ETF at 40, 48, 36 PLN and 0.05 of the index at
	// 1000, 1000, 1100.
	wantDays := []portfolio.Point{
		{Date: mustDate(t, "2025-01-02"), Value: mustDec(t, "100")},
		{Date: mustDate(t, "2025-01-03"), Value: mustDec(t, "110")},
		{Date: mustDate(t, "2025-01-04"), Value: mustDec(t, "100")},
	}
	if diff := cmp.Diff(wantDays, got.Backtest.Days); diff != "" {
		t.Errorf("backtest (-want +got):\n%s", diff)
	}
	wantDrawdown := portfolio.Drawdown{
		Peak: mustDate(t, "2025-01-03"), Trough: mustDate(t, "2025-01-04"),
		PeakValue: mustDec(t, "110"), TroughValue: mustDec(t, "100"), Depth: mustDec(t, "0.090909090909"),
	}
	if diff := cmp.Diff(wantDrawdown, got.Drawdown); diff != "" {
		t.Errorf("model drawdown (-want +got):\n%s", diff)
	}
	wantTargets := []app.TargetReport{
		{Target: model.Targets[0], First: mustDate(t, "2025-01-02"), Last: mustDate(t, "2025-01-04"), Drawdown: portfolio.Drawdown{
			Peak: mustDate(t, "2025-01-03"), Trough: mustDate(t, "2025-01-04"), PeakValue: mustDec(t, "48"), TroughValue: mustDec(t, "36"), Depth: mustDec(t, "0.25"),
		}},
		{Target: model.Targets[1], First: mustDate(t, "2025-01-01"), Last: mustDate(t, "2025-01-04"), Drawdown: portfolio.Drawdown{
			Peak: mustDate(t, "2025-01-01"), Trough: mustDate(t, "2025-01-01"), PeakValue: mustDec(t, "1000"), TroughValue: mustDec(t, "1000"), Depth: mustDec(t, "0"),
		}},
	}
	if diff := cmp.Diff(wantTargets, got.Targets); diff != "" {
		t.Errorf("targets (-want +got):\n%s", diff)
	}
	if !got.Reporting.Equal(mustCurrency(t, "PLN")) || !got.Model.Enabled {
		t.Errorf("report currency %s, model %+v", got.Reporting, got.Model)
	}
}

func TestReportOff(t *testing.T) {
	t.Parallel()
	market := reportMarket(t)
	off := sixtyForty(t)
	off.Enabled = false
	got, on, err := newModels(t, &fakeStore{model: off}, market).Report(t.Context(), mustDate(t, "2025-01-04"))
	if err != nil || on || len(got.Targets) != 0 || len(got.Model.Targets) != 0 {
		t.Errorf("Report of an off model = %+v, %v, %v; want nothing", got, on, err)
	}
	if market.reads != 0 {
		t.Errorf("an off model read the market data %d times", market.reads)
	}
}

func TestReportMissingAndFailures(t *testing.T) {
	t.Parallel()
	model := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		{Kind: portfolio.TargetInstrument, Name: etfIE, Weight: mustDec(t, "50")},
		{Kind: portfolio.TargetIndex, Name: "NONE", Currency: mustCurrency(t, "PLN"), Weight: mustDec(t, "50")},
	}}
	got, on, err := newModels(t, &fakeStore{model: model}, reportMarket(t)).Report(t.Context(), mustDate(t, "2025-01-04"))
	if err != nil || !on || len(got.Backtest.Days) != 0 || got.Targets[1].Missing != portfolio.MissingCloses || !got.Drawdown.Depth.IsZero() {
		t.Errorf("Report with an index without levels = %+v, %v, %v", got, on, err)
	}
	for _, tt := range []struct {
		name   string
		store  *fakeStore
		market *fakeMarket
	}{
		{name: "the store", store: &fakeStore{err: errStoreDown}, market: reportMarket(t)},
		{name: "the market data", store: &fakeStore{model: model}, market: &fakeMarket{err: errStoreDown}},
	} {
		if _, _, err := newModels(t, tt.store, tt.market).Report(t.Context(), mustDate(t, "2025-01-04")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
			t.Errorf("%s failing: Report = %v", tt.name, err)
		}
	}
}
