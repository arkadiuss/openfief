package pl_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// attribution is a synthetic attribution of published data.
func attribution(t *testing.T) pl.Attribution {
	t.Helper()
	return pl.Attribution{
		Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
		URL: "https://www.example.org/attachment/file", Made: date(t, "2026-10-26"),
		Fetched: time.Date(2026, time.November, 7, 6, 0, 0, 0, time.UTC), Processing: "read and rounded",
	}
}

// TestPublishedRatesOf looks series up in synthetic published rates: an
// indexed series with its margin, a fixed-rate one without, and a series the
// data doesn't list.
func TestPublishedRatesOf(t *testing.T) {
	t.Parallel()
	edo := pl.PublishedRates{Series: series(t, "EDO1136"), FirstPeriodRate: dec(t, "0.0611"), Margin: dec(t, "0.0222")}
	tos := pl.PublishedRates{Series: series(t, "TOS1129"), FirstPeriodRate: dec(t, "0.0433")}
	p := pl.Published{Rates: []pl.PublishedRates{edo, tos}, Attribution: attribution(t)}
	tests := []struct {
		name    string
		series  string
		want    pl.PublishedRates
		wantErr error
	}{
		{name: "an indexed series: rate and margin", series: "EDO1136", want: edo},
		{name: "a fixed-rate series: no margin", series: "TOS1129", want: tos},
		{name: "a series the data doesn't list", series: "COI1130", wantErr: pl.ErrNotPublished},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := p.RatesOf(series(t, tt.series))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("RatesOf error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RatesOf (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAttributionValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		edit    func(a *pl.Attribution)
		wantErr error
	}{
		{name: "complete", edit: func(*pl.Attribution) {}},
		{name: "no publisher", edit: func(a *pl.Attribution) { a.Publisher = "" }, wantErr: pl.ErrInvalidAttribution},
		{name: "no dataset", edit: func(a *pl.Attribution) { a.Dataset = "" }, wantErr: pl.ErrInvalidAttribution},
		{name: "no licence", edit: func(a *pl.Attribution) { a.License = "" }, wantErr: pl.ErrInvalidAttribution},
		{name: "no address", edit: func(a *pl.Attribution) { a.URL = "" }, wantErr: pl.ErrInvalidAttribution},
		{name: "no date made", edit: func(a *pl.Attribution) { a.Made = kernel.Date{} }, wantErr: pl.ErrInvalidAttribution},
		{name: "no time taken", edit: func(a *pl.Attribution) { a.Fetched = time.Time{} }, wantErr: pl.ErrInvalidAttribution},
		{name: "no processing", edit: func(a *pl.Attribution) { a.Processing = "" }, wantErr: pl.ErrInvalidAttribution},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := attribution(t)
			tt.edit(&a)
			if diff := cmp.Diff(tt.wantErr, a.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}
