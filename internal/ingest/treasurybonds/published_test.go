package treasurybonds_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// errOutage is a synthetic failure to read the published data.
var errOutage = errors.New("synthetic outage")

// published is a synthetic source of published data that counts its reads:
// EDO1136 at 6.11% with a 2.22% margin, EDO1236 at 6.55% with a 2% margin
// and TOS1129 at 4.33%, with the synthetic registry ISINs PL9999000018 of
// EDO1136 and PL9999000026 of EDO1236, and PL9999000034 of COI1130, whose
// rates it doesn't list yet (D-214).
type published struct {
	reads int
	err   error
}

func (p *published) Published(context.Context) (*pl.Published, error) {
	p.reads++
	if p.err != nil {
		return nil, p.err
	}
	edo, err1 := pl.ParseSeries("EDO1136")
	tos, err2 := pl.ParseSeries("TOS1129")
	later, err3 := pl.ParseSeries("EDO1236")
	coi, err4 := pl.ParseSeries("COI1130")
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		return nil, err
	}
	return &pl.Published{
		Rates: []pl.PublishedRates{
			{Series: edo, FirstPeriodRate: kernel.NewDecimal(611, -4), Margin: kernel.NewDecimal(222, -4)},
			{Series: later, FirstPeriodRate: kernel.NewDecimal(655, -4), Margin: kernel.NewDecimal(200, -4)},
			{Series: tos, FirstPeriodRate: kernel.NewDecimal(433, -4)},
		},
		ISINs: []pl.PublishedISIN{
			{Series: edo, ISIN: "PL9999000018"}, {Series: later, ISIN: "PL9999000026"}, {Series: coi, ISIN: "PL9999000034"},
		},
		Attribution: syntheticAttribution("read"),
	}, nil
}

// syntheticAttribution is the synthetic data's attribution with processing.
func syntheticAttribution(processing string) pl.Attribution {
	made, err := kernel.NewDate(2026, time.October, 26)
	if err != nil {
		return pl.Attribution{}
	}
	return pl.Attribution{
		Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
		URL: "https://www.example.org/attachment/x", Made: made,
		Fetched: time.Date(2026, time.November, 7, 6, 0, 0, 0, time.UTC), Processing: processing,
	}
}

// TestCompleteTerms fills in what the owner left to the connector (D-97):
// published rates with their attribution, which also says where the fee
// came from, the reviewed table's fee by sale month, carried forward as
// assumed past the reviewed month, and nothing for terms the owner
// entered, which win. A lot that left everything to a series the data
// doesn't list yet stays pending. The data is read once whenever there is
// a lot, for its series' registry ISIN (D-214).
func TestCompleteTerms(t *testing.T) {
	t.Parallel()
	table := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief took it from its reviewed table of the issue letters")
	assumed := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief assumed the latest fee of its reviewed table of the issue letters, for the owner to confirm")
	entered := syntheticAttribution("read; the early-redemption fee is not in the data: the owner entered it")
	tests := []struct {
		name       string
		lot        func(l *ingest.BondLot)
		want       func(l *ingest.BondLot)
		wantErrors []error
	}{
		{
			name: "nothing left out: the owner's terms",
			lot:  func(*ingest.BondLot) {},
			want: func(*ingest.BondLot) {},
		},
		{
			name: "everything left out",
			lot:  func(l *ingest.BondLot) { l.PublishedRates, l.ReviewedFee = true, true },
			want: func(l *ingest.BondLot) {
				l.Terms.FirstPeriodRate, l.Terms.Margin = kernel.NewDecimal(611, -4), kernel.NewDecimal(222, -4)
				l.Terms.EarlyRedemptionFee, l.Terms.Attribution, l.Terms.FeeSource = pln(t, "3.00"), &table, pl.FeeFromTable
			},
		},
		{
			name: "everything left out, sold after the fee table was reviewed: the fee assumed",
			lot: func(l *ingest.BondLot) {
				*l = lot(t, "EDO1236", "10", "2026-11-03", "0.05")
				l.PublishedRates, l.ReviewedFee = true, true
			},
			want: func(l *ingest.BondLot) {
				*l = lot(t, "EDO1236", "10", "2026-11-03", "0.05")
				l.Terms.FirstPeriodRate, l.Terms.Margin = kernel.NewDecimal(655, -4), kernel.NewDecimal(200, -4)
				l.Terms.EarlyRedemptionFee, l.Terms.Attribution, l.Terms.FeeSource = pln(t, "3.00"), &assumed, pl.FeeAssumed
			},
		},
		{
			name: "everything left out of a series the data doesn't list yet: pending, no error",
			lot: func(l *ingest.BondLot) {
				*l = lot(t, "COI1130", "10", "2026-11-03", "0.05")
				l.PublishedRates, l.ReviewedFee = true, true
			},
			want: func(l *ingest.BondLot) {
				*l = lot(t, "COI1130", "10", "2026-11-03", "0.05")
				l.PublishedRates, l.ReviewedFee = true, true
			},
		},
		{
			name: "the rates left out, the fee entered",
			lot:  func(l *ingest.BondLot) { l.PublishedRates = true },
			want: func(l *ingest.BondLot) {
				l.Terms.FirstPeriodRate, l.Terms.Margin, l.Terms.Attribution = kernel.NewDecimal(611, -4), kernel.NewDecimal(222, -4), &entered
			},
		},
		{
			name: "only the fee left out: the table's, no attribution",
			lot:  func(l *ingest.BondLot) { l.ReviewedFee = true },
			want: func(l *ingest.BondLot) {
				l.Terms.EarlyRedemptionFee, l.Terms.FeeSource = pln(t, "3.00"), pl.FeeFromTable
			},
		},
		{
			name: "only the fee left out, sold after the fee table was reviewed: assumed",
			lot: func(l *ingest.BondLot) {
				*l = lot(t, "EDO1236", "10", "2026-11-03", "0.05")
				l.ReviewedFee = true
			},
			want: func(l *ingest.BondLot) {
				*l = lot(t, "EDO1236", "10", "2026-11-03", "0.05")
				l.Terms.EarlyRedemptionFee, l.Terms.FeeSource = pln(t, "3.00"), pl.FeeAssumed
			},
		},
		{
			name: "the rates of a series the data doesn't list, the fee entered",
			lot: func(l *ingest.BondLot) {
				*l = lot(t, "COI1130", "10", "2026-11-03", "0.05")
				l.PublishedRates = true
			},
			wantErrors: []error{pl.ErrNotPublished},
		},
		{
			name: "a series sold before its type's first fee row",
			lot: func(l *ingest.BondLot) {
				*l = lot(t, "EDO1223", "10", "2013-12-03", "0.05")
				l.ReviewedFee = true
			},
			wantErrors: []error{pl.ErrNoReviewedFee},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := lot(t, "EDO1136", "10", "2026-11-03", "0.05")
			tt.lot(&l)
			src := &published{}
			e := ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{l}}
			_, err := treasurybonds.CompleteTerms(context.Background(), src, &e)
			if src.reads != 1 {
				t.Errorf("read the published data %d times, want once", src.reads)
			}
			if tt.wantErrors != nil {
				wantErrors(t, err, tt.wantErrors)
				return
			}
			want := lot(t, "EDO1136", "10", "2026-11-03", "0.05")
			tt.want(&want)
			wantLots(t, err, []ingest.BondLot{want}, e.Lots)
		})
	}
}

// wantErrors fails t unless err wraps every one of want.
func wantErrors(t *testing.T, err error, want []error) {
	t.Helper()
	for _, w := range want {
		if !errors.Is(err, w) {
			t.Errorf("CompleteTerms = %v, want %v", err, w)
		}
	}
}

// wantLots fails t unless CompleteTerms succeeded with the lots want.
func wantLots(t *testing.T, err error, want, got []ingest.BondLot) {
	t.Helper()
	if err != nil {
		t.Fatalf("CompleteTerms: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lots (-want +got):\n%s", diff)
	}
}

// TestCompleteTermsReadsOnce reads the published data once for a batch of
// lots and returns it, returns a failure to read it as it came when a lot
// needs its rates, and otherwise no data, so the lots wait for their
// registry ISIN (D-214); a batch of deletions alone reads nothing.
func TestCompleteTermsReadsOnce(t *testing.T) {
	t.Parallel()
	edo, tos := lot(t, "EDO1136", "10", "2026-11-03", "0"), lot(t, "TOS1129", "5", "2026-11-04", "0")
	edo.PublishedRates, tos.PublishedRates = true, true
	src := &published{}
	e := ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{edo, tos}}
	p, err := treasurybonds.CompleteTerms(context.Background(), src, &e)
	if err != nil || p == nil {
		t.Fatalf("CompleteTerms = %v, %v, want the data read", p, err)
	}
	if src.reads != 1 {
		t.Errorf("read the published data %d times, want once", src.reads)
	}
	if got := e.Lots[1].Terms.FirstPeriodRate; !got.Equal(kernel.NewDecimal(433, -4)) {
		t.Errorf("TOS1129 first-period rate = %s, want 0.0433", got)
	}
	down := &published{err: errOutage}
	e = ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{edo}}
	_, err = treasurybonds.CompleteTerms(context.Background(), down, &e)
	if diff := cmp.Diff(errOutage, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("CompleteTerms error (-want +got):\n%s", diff)
	}
}

// TestCompleteTermsWithoutData returns data that lists nothing, so the
// lots wait for their registry ISIN (D-214), for lots whose terms the owner
// entered while the data can't be read, and for a batch of deletions
// alone, which reads nothing.
func TestCompleteTermsWithoutData(t *testing.T) {
	t.Parallel()
	down := &published{err: errOutage}
	e := ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{lot(t, "EDO1136", "10", "2026-11-03", "0.05")}}
	p, err := treasurybonds.CompleteTerms(context.Background(), down, &e)
	if err != nil || p == nil || len(p.ISINs)+len(p.Rates) != 0 {
		t.Errorf("CompleteTerms of entered terms while the data is down = %v, %v, want data listing nothing", p, err)
	}
	none := &published{}
	e = ingest.BondLotEntries{Account: account, Deleted: []ingest.BondLotKey{key(t, "EDO1136", "2026-11-03")}}
	p, err = treasurybonds.CompleteTerms(context.Background(), none, &e)
	if err != nil || p == nil || len(p.ISINs)+len(p.Rates) != 0 || none.reads != 0 {
		t.Errorf("CompleteTerms of deletions = %v, %v after %d reads, want data listing nothing and no read", p, err, none.reads)
	}
}

// held is one account holding the instruments of codes, the first as a
// holding, the rest as open lots.
func held(t *testing.T, codes ...string) []ingest.HeldAccount {
	t.Helper()
	a := ingest.HeldAccount{Account: account}
	for i, code := range codes {
		in := instrument(t, code)
		if i == 0 {
			a.Holdings = append(a.Holdings, ledger.Holding{Account: account, Instrument: in, Quantity: kernel.NewQuantity(kernel.NewDecimal(1, 0))})
			continue
		}
		a.Lots = append(a.Lots, ledger.Lot{Instrument: in, Quantity: kernel.NewQuantity(kernel.NewDecimal(1, 0))})
	}
	return []ingest.HeldAccount{a}
}

// TestMissingTerms lists the held bond instruments without stored terms,
// under their provisional ISIN or the registry's (D-214), each once in
// order, and never one whose terms are stored or an instrument that is no
// treasury bond.
func TestMissingTerms(t *testing.T) {
	t.Parallel()
	etf := ledger.Instrument{ISIN: instrument(t, "EDO1136").ISIN, Symbol: "SYNT"}
	foreign := registry(t, "EDO0536", "DE0000000009")
	stored := []ingest.BondTerms{{Instrument: instrument(t, "EDO1136")}}
	accounts := held(t, "TOS1129", "EDO1136", "EDO1236", "TOS1129")
	accounts[0].Holdings = append(accounts[0].Holdings,
		ledger.Holding{Instrument: etf}, ledger.Holding{Instrument: foreign},
		ledger.Holding{Instrument: registry(t, "COI1130", "PL9999000034")})
	var got []string
	for _, in := range treasurybonds.MissingTerms(accounts, stored) {
		got = append(got, in.String())
	}
	want := []string{"PL9999000034/COI1130", instrument(t, "EDO1236").String(), instrument(t, "TOS1129").String()}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("MissingTerms (-want +got):\n%s", diff)
	}
	if got := treasurybonds.MissingTerms(held(t, "EDO1136"), stored); got != nil {
		t.Errorf("MissingTerms with every series stored = %v, want none", got)
	}
}

// TestFillMissing fills in the series the data lists, the fee assumed past
// the reviewed month, keyed by its registry ISIN when it was held under its
// provisional one (D-214), and leaves out one whose rates it doesn't list.
// SyncPublished reads the data only when something is missing or
// provisional, once, and returns a failure to read it.
func TestFillMissing(t *testing.T) {
	t.Parallel()
	src := &published{}
	if terms, isins, err := treasurybonds.SyncPublished(t.Context(), src, nil, nil); terms != nil || isins != nil || err != nil || src.reads != 0 {
		t.Errorf("SyncPublished(none) = %v, %v, %v after %d reads, want nothing and no read", terms, isins, err, src.reads)
	}
	missing := []ledger.Instrument{instrument(t, "COI1130"), instrument(t, "EDO1236")}
	got, isins, err := treasurybonds.SyncPublished(t.Context(), src, missing, nil)
	if err != nil || isins != nil {
		t.Fatalf("SyncPublished = %v, %v", isins, err)
	}
	assumed := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief assumed the latest fee of its reviewed table of the issue letters, for the owner to confirm")
	want := []ingest.BondTerms{{Instrument: registry(t, "EDO1236", "PL9999000026"), Terms: pl.Terms{
		Series: series(t, "EDO1236"), FirstPeriodRate: kernel.NewDecimal(655, -4), Margin: kernel.NewDecimal(200, -4),
		EarlyRedemptionFee: pln(t, "3.00"), Attribution: &assumed, FeeSource: pl.FeeAssumed,
	}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FillMissing (-want +got):\n%s", diff)
	}
	if src.reads != 1 {
		t.Errorf("read the data %d times, want once", src.reads)
	}
	_, _, err = treasurybonds.SyncPublished(t.Context(), &published{err: errOutage}, missing, nil)
	if diff := cmp.Diff(errOutage, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("SyncPublished error (-want +got):\n%s", diff)
	}
}

// TestRegistryISINs: a sync delivers the registry ISIN of each provisional
// instrument whose series the data lists, an ISIN listed before the
// series' rates too, and nothing for a series the data doesn't list, an
// instrument under its registry ISIN already, or one whose ISIN isn't its
// series' provisional one (D-214).
func TestRegistryISINs(t *testing.T) {
	t.Parallel()
	src := &published{}
	mismatched := ledger.Instrument{ISIN: instrument(t, "EDO1136").ISIN, Symbol: "EDO1236"}
	provisional := []ledger.Instrument{
		instrument(t, "EDO1136"), instrument(t, "COI1130"), instrument(t, "TOS1129"),
		registry(t, "EDO1236", "PL9999000026"), mismatched,
	}
	terms, got, err := treasurybonds.SyncPublished(t.Context(), src, nil, provisional)
	if err != nil || terms != nil {
		t.Fatalf("SyncPublished = %v, %v", terms, err)
	}
	isin := func(s string) ledger.ISIN { return registry(t, "X", s).ISIN }
	want := []ingest.RegistryISIN{
		{Instrument: instrument(t, "EDO1136"), ISIN: isin("PL9999000018")},
		{Instrument: instrument(t, "COI1130"), ISIN: isin("PL9999000034")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SyncPublished registry ISINs (-want +got):\n%s", diff)
	}
	if src.reads != 1 {
		t.Errorf("read the data %d times, want once", src.reads)
	}
}

// series parses code.
func series(t *testing.T, code string) pl.Series {
	t.Helper()
	s, err := pl.ParseSeries(code)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
