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
// and TOS1129 at 4.33%.
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
	if err := errors.Join(err1, err2, err3); err != nil {
		return nil, err
	}
	return &pl.Published{
		Rates: []pl.PublishedRates{
			{Series: edo, FirstPeriodRate: kernel.NewDecimal(611, -4), Margin: kernel.NewDecimal(222, -4)},
			{Series: later, FirstPeriodRate: kernel.NewDecimal(655, -4), Margin: kernel.NewDecimal(200, -4)},
			{Series: tos, FirstPeriodRate: kernel.NewDecimal(433, -4)},
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
// doesn't list yet stays pending.
func TestCompleteTerms(t *testing.T) {
	t.Parallel()
	table := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief took it from its reviewed table of the issue letters")
	assumed := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief assumed the latest fee of its reviewed table of the issue letters, for the owner to confirm")
	entered := syntheticAttribution("read; the early-redemption fee is not in the data: the owner entered it")
	tests := []struct {
		name       string
		lot        func(l *ingest.BondLot)
		want       func(l *ingest.BondLot)
		wantReads  int
		wantErrors []error
	}{
		{
			name: "nothing left out: the owner's terms, no read",
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
			wantReads: 1,
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
			wantReads: 1,
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
			wantReads: 1,
		},
		{
			name: "the rates left out, the fee entered",
			lot:  func(l *ingest.BondLot) { l.PublishedRates = true },
			want: func(l *ingest.BondLot) {
				l.Terms.FirstPeriodRate, l.Terms.Margin, l.Terms.Attribution = kernel.NewDecimal(611, -4), kernel.NewDecimal(222, -4), &entered
			},
			wantReads: 1,
		},
		{
			name: "only the fee left out: the table's, no read, no attribution",
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
			wantReads: 1, wantErrors: []error{pl.ErrNotPublished},
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
			err := treasurybonds.CompleteTerms(context.Background(), src, &e)
			if src.reads != tt.wantReads {
				t.Errorf("read the published data %d times, want %d", src.reads, tt.wantReads)
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
// lots, and returns a failure to read it as it came.
func TestCompleteTermsReadsOnce(t *testing.T) {
	t.Parallel()
	edo, tos := lot(t, "EDO1136", "10", "2026-11-03", "0"), lot(t, "TOS1129", "5", "2026-11-04", "0")
	edo.PublishedRates, tos.PublishedRates = true, true
	src := &published{}
	e := ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{edo, tos}}
	if err := treasurybonds.CompleteTerms(context.Background(), src, &e); err != nil {
		t.Fatalf("CompleteTerms: %v", err)
	}
	if src.reads != 1 {
		t.Errorf("read the published data %d times, want once", src.reads)
	}
	if got := e.Lots[1].Terms.FirstPeriodRate; !got.Equal(kernel.NewDecimal(433, -4)) {
		t.Errorf("TOS1129 first-period rate = %s, want 0.0433", got)
	}
	down := &published{err: errOutage}
	e = ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{edo}}
	err := treasurybonds.CompleteTerms(context.Background(), down, &e)
	if diff := cmp.Diff(errOutage, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("CompleteTerms error (-want +got):\n%s", diff)
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

// TestMissingTerms lists the held series without stored terms, each once
// in series order, and never one whose terms are stored or an instrument
// that is no bond of this connector.
func TestMissingTerms(t *testing.T) {
	t.Parallel()
	etf := ledger.Instrument{ISIN: instrument(t, "EDO1136").ISIN, Symbol: "SYNT"}
	stored := []ingest.BondTerms{{Instrument: instrument(t, "EDO1136")}}
	accounts := held(t, "TOS1129", "EDO1136", "EDO1236", "TOS1129")
	accounts[0].Holdings = append(accounts[0].Holdings, ledger.Holding{Instrument: etf})
	var got []string
	for _, s := range treasurybonds.MissingTerms(accounts, stored) {
		got = append(got, s.String())
	}
	if diff := cmp.Diff([]string{"EDO1236", "TOS1129"}, got); diff != "" {
		t.Errorf("MissingTerms (-want +got):\n%s", diff)
	}
	if got := treasurybonds.MissingTerms(held(t, "EDO1136"), stored); got != nil {
		t.Errorf("MissingTerms with every series stored = %v, want none", got)
	}
}

// TestFillMissing fills in the series the data lists, the fee assumed past
// the reviewed month, leaves out one it doesn't list, reads the data only
// when something is missing and returns a failure to read it.
func TestFillMissing(t *testing.T) {
	t.Parallel()
	src := &published{}
	if got, err := treasurybonds.FillMissing(t.Context(), src, nil); got != nil || err != nil || src.reads != 0 {
		t.Errorf("FillMissing(none) = %v, %v after %d reads, want nothing and no read", got, err, src.reads)
	}
	missing := []pl.Series{series(t, "COI1130"), series(t, "EDO1236")}
	got, err := treasurybonds.FillMissing(t.Context(), src, missing)
	if err != nil {
		t.Fatalf("FillMissing: %v", err)
	}
	assumed := syntheticAttribution("read; the early-redemption fee is not in the data: OpenFief assumed the latest fee of its reviewed table of the issue letters, for the owner to confirm")
	want := []ingest.BondTerms{{Instrument: instrument(t, "EDO1236"), Terms: pl.Terms{
		Series: series(t, "EDO1236"), FirstPeriodRate: kernel.NewDecimal(655, -4), Margin: kernel.NewDecimal(200, -4),
		EarlyRedemptionFee: pln(t, "3.00"), Attribution: &assumed, FeeSource: pl.FeeAssumed,
	}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FillMissing (-want +got):\n%s", diff)
	}
	if src.reads != 1 {
		t.Errorf("read the data %d times, want once", src.reads)
	}
	_, err = treasurybonds.FillMissing(t.Context(), &published{err: errOutage}, missing)
	if diff := cmp.Diff(errOutage, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("FillMissing error (-want +got):\n%s", diff)
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
