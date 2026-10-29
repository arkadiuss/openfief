package nbp_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// change is a reference rate change from day at rate, a fraction.
func change(t *testing.T, day, rate string) pl.RateChange {
	t.Helper()
	r, err := kernel.ParseDecimal(rate)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", rate, err)
	}
	return pl.RateChange{From: date(t, day), Rate: r}
}

// TestReferenceRatesArchive replays NBP's archive as recorded on
// 2026-10-29, which ends with the change of 2026-03-05.
func TestReferenceRatesArchive(t *testing.T) {
	t.Parallel()
	a := nbp.NewArchive(recordingClient(t), nbp.DefaultArchiveURL)
	got, err := a.ReferenceRates(context.Background())
	if err != nil {
		t.Fatalf("ReferenceRates: %v", err)
	}
	if len(got) != 96 {
		t.Errorf("got %d changes, want the archive's 96", len(got))
	}
	for i := 1; i < len(got); i++ {
		if !got[i-1].From.Before(got[i].From) {
			t.Errorf("change %d (%s) is not after change %d (%s)", i, got[i].From, i-1, got[i-1].From)
		}
	}
	// Spot checks, copied from the recorded response text.
	want := map[string]pl.RateChange{}
	for _, c := range []pl.RateChange{
		change(t, "1998-02-26", "0.24"),
		change(t, "2023-10-05", "0.0575"),
		change(t, "2025-05-08", "0.0525"),
		change(t, "2025-07-03", "0.05"),
		change(t, "2026-03-05", "0.0375"),
	} {
		want[c.From.String()] = c
	}
	for _, c := range got {
		w, ok := want[c.From.String()]
		if !ok {
			continue
		}
		delete(want, c.From.String())
		if diff := cmp.Diff(w, c); diff != "" {
			t.Errorf("change from %s (-want +got):\n%s", c.From, diff)
		}
	}
	if len(want) != 0 {
		t.Errorf("changes missing: %v", want)
	}
}

func TestReferenceRatesParsesTheArchive(t *testing.T) {
	t.Parallel()
	// Synthetic archive: a byte order mark, changes out of order, and one
	// that leaves the reference rate as it was.
	body := "\uFEFF" + `<?xml version="1.0" encoding="utf-8"?>
<stopy_procentowe_archiwum data_publikacji="2015-03-04">
  <pozycje obowiazuje_od="2025-07-03"><pozycja id="ref" oprocentowanie="5,00" /><pozycja id="lom" oprocentowanie="5,50" /></pozycje>
  <pozycje obowiazuje_od="2025-05-08"><pozycja id="lom" oprocentowanie="5,75" /><pozycja id="ref" oprocentowanie="5,25" /></pozycje>
  <pozycje obowiazuje_od="2025-06-01"><pozycja id="dep" oprocentowanie="4,50" /></pozycje>
</stopy_procentowe_archiwum>`
	a := nbp.NewArchive(stubClient(http.StatusOK, body), nbp.DefaultArchiveURL)
	got, err := a.ReferenceRates(context.Background())
	if err != nil {
		t.Fatalf("ReferenceRates: %v", err)
	}
	want := []pl.RateChange{change(t, "2025-05-08", "0.0525"), change(t, "2025-07-03", "0.05")}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestReferenceRatesRejects(t *testing.T) {
	t.Parallel()
	wrap := func(changes string) string {
		return `<stopy_procentowe_archiwum>` + changes + `</stopy_procentowe_archiwum>`
	}
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "not found", status: http.StatusNotFound, body: ""},
		{name: "not XML", status: http.StatusOK, body: "<html"},
		{name: "another document", status: http.StatusOK, body: `<stopy_procentowe></stopy_procentowe>`},
		{name: "bad date", status: http.StatusOK, body: wrap(`<pozycje obowiazuje_od="08.05.2025"><pozycja id="ref" oprocentowanie="5,25" /></pozycje>`)},
		{name: "rate not a number", status: http.StatusOK, body: wrap(`<pozycje obowiazuje_od="2025-05-08"><pozycja id="ref" oprocentowanie="n/a" /></pozycje>`)},
		{name: "two changes on one day", status: http.StatusOK, body: wrap(
			`<pozycje obowiazuje_od="2025-05-08"><pozycja id="ref" oprocentowanie="5,25" /></pozycje>` +
				`<pozycje obowiazuje_od="2025-05-08"><pozycja id="ref" oprocentowanie="5,00" /></pozycje>`,
		)},
		{name: "no reference rate", status: http.StatusOK, body: wrap(`<pozycje obowiazuje_od="2025-05-08"><pozycja id="lom" oprocentowanie="5,75" /></pozycje>`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := nbp.NewArchive(stubClient(tt.status, tt.body), nbp.DefaultArchiveURL)
			_, err := a.ReferenceRates(context.Background())
			want := nbp.ErrMalformedArchive
			if tt.status != http.StatusOK {
				want = nbp.ErrUnexpectedStatus
			}
			if diff := cmp.Diff(want, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}
