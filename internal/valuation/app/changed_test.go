package app_test

import (
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// changedRecorder returns the input recorder over the fake store, which counts
// every value it is given as new.
func changedRecorder(t *testing.T) *app.InputRecorder {
	t.Helper()
	l := twoAccounts(t)
	return app.NewInputRecorder(app.NewPlanner(l, l), &fakeRecorder{}, pl.NewPack())
}

// TestRecordFXRatesChangedReportsTheStoresCount checks that the recorder hands
// the store's count of the FX rates that were new or different back to its
// caller (D-211), after the checks of the recording without a count.
func TestRecordFXRatesChangedReportsTheStoresCount(t *testing.T) {
	t.Parallel()
	r := changedRecorder(t)
	rates := []valuation.FXRate{rate(t, "USD", "2026-03-05", "4"), rate(t, "EUR", "2026-03-04", "4.3")}
	if got, err := r.RecordFXRatesChanged(t.Context(), rates); err != nil || got != 2 {
		t.Errorf("RecordFXRatesChanged = %d, %v; want the store's 2", got, err)
	}
	if got, err := r.RecordFXRatesChanged(t.Context(), []valuation.FXRate{{Date: date(t, "2026-03-05")}}); err == nil || got != 0 {
		t.Errorf("RecordFXRatesChanged of a rate without a currency = %d, %v; want no count and an error", got, err)
	}
}

// TestRecordIndexReadingsChangedReportsTheStoresCount is the same for index
// readings.
func TestRecordIndexReadingsChangedReportsTheStoresCount(t *testing.T) {
	t.Parallel()
	r := changedRecorder(t)
	readings := []valuation.IndexReading{{Kind: valuation.CPI, Year: 2026, Month: time.February, Value: dec(t, "0.03"), Source: "bonds"}}
	if got, err := r.RecordIndexReadingsChanged(t.Context(), readings); err != nil || got != 1 {
		t.Errorf("RecordIndexReadingsChanged = %d, %v; want the store's 1", got, err)
	}
	if got, err := r.RecordIndexReadingsChanged(t.Context(), []valuation.IndexReading{{Year: 2026, Month: time.March}}); err == nil || got != 0 {
		t.Errorf("RecordIndexReadingsChanged of a reading without a kind = %d, %v; want no count and an error", got, err)
	}
}

// TestRecordIndexLevelsChangedReportsTheStoresCount is the same for market
// index levels.
func TestRecordIndexLevelsChangedReportsTheStoresCount(t *testing.T) {
	t.Parallel()
	r := changedRecorder(t)
	levels := []valuation.IndexLevel{
		{Index: "FAKE20", Date: date(t, "2026-03-02"), Level: dec(t, "1522.6"), Source: "prices"},
		{Index: "FAKE20", Date: date(t, "2026-03-03"), Level: dec(t, "1530"), Source: "prices"},
	}
	if got, err := r.RecordIndexLevelsChanged(t.Context(), levels); err != nil || got != 2 {
		t.Errorf("RecordIndexLevelsChanged = %d, %v; want the store's 2", got, err)
	}
	if got, err := r.RecordIndexLevelsChanged(t.Context(), []valuation.IndexLevel{{Index: "FAKE20", Date: date(t, "2026-03-02")}}); err == nil || got != 0 {
		t.Errorf("RecordIndexLevelsChanged of a level without a value or a source = %d, %v; want no count and an error", got, err)
	}
}
