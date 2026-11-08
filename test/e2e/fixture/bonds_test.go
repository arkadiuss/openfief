package main

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/gus"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// fetchedAt is the clock the ministry client dates its fetch with here.
func fetchedAt() time.Time { return time.Date(2026, time.November, 8, 9, 0, 0, 0, time.UTC) }

// publishMinistry publishes the fixture's ministry data, as the suite does.
func publishMinistry(t *testing.T, client *http.Client, base string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, base+MinistryControlPath, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", MinistryControlPath, err)
	}
	if err := resp.Body.Close(); err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST %s: status %d (%v), want 204", MinistryControlPath, resp.StatusCode, err)
	}
}

// describePublished writes what the published data lists: each series'
// ISIN, rates, and the attribution's source.
func describePublished(p *pl.Published) []string {
	var out []string
	for _, r := range p.Rates {
		out = append(out, r.Series.String()+" rates "+r.FirstPeriodRate.String()+" "+r.Margin.String())
	}
	for _, i := range p.ISINs {
		out = append(out, i.Series.String()+" ISIN "+i.ISIN)
	}
	slices.Sort(out)
	a := p.Attribution
	return append(out, "from "+a.Publisher+", "+a.Dataset+", "+a.License+", made "+a.Made.String())
}

// TestFixtureServesMinistryData checks that the connector's own ministry
// client finds nothing while the fixture withholds the data, and, once the
// suite publishes it, reads the synthetic spreadsheet the gov.pl page
// links, and the same file through dane.gov.pl's resource list when gov.pl
// fails: each series' registry ISIN, first-period rate and margin.
func TestFixtureServesMinistryData(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	client := func(gov string) *treasurybonds.MinistryClient {
		return treasurybonds.NewMinistryClient(treasurybonds.MinistryConfig{
			GovPL: srv.Client(), GovPLBaseURL: gov, Dane: srv.Client(), DaneBaseURL: srv.URL + DanePath, Now: fetchedAt,
		})
	}
	if _, err := client(srv.URL + GovPLPath).Published(t.Context()); !errors.Is(err, treasurybonds.ErrUnexpectedStatus) {
		t.Errorf("Published before the data is published: %v, want %v", err, treasurybonds.ErrUnexpectedStatus)
	}
	publishMinistry(t, srv.Client(), srv.URL)
	want := []string{
		"COI1029 ISIN " + ISINCOI, "COI1029 rates 0.056 0.015",
		"EDO0534 ISIN " + ISINEDO, "EDO0534 rates 0.065 0.0175",
		"TOS0329 ISIN " + ISINTOS, "TOS0329 rates 0.0565 0",
	}
	tests := []struct {
		name, gov, url string
	}{
		{name: "gov.pl", gov: srv.URL + GovPLPath, url: srv.URL + GovPLPath + govPLFile},
		{name: "dane.gov.pl", gov: srv.URL + "/nowhere", url: srv.URL + DanePath + daneFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := client(tt.gov).Published(t.Context())
			if err != nil {
				t.Fatalf("Published: %v", err)
			}
			got := describePublished(p)
			wantAll := append(slices.Clone(want), "from "+treasurybonds.Publisher+", "+ministryTitle+", "+treasurybonds.License+", made "+ministryMade)
			if diff := cmp.Diff(wantAll, got); diff != "" || p.Attribution.URL != tt.url {
				t.Errorf("published data (-want +got):\n%s\nURL %s, want %s", diff, p.Attribution.URL, tt.url)
			}
		})
	}
}

// TestFixtureServesBondIndices checks that the connector's own GUS client
// reads exactly the demo dataset's CPI from the fixture, year by year and
// with the indicator of each year's classification, and its NBP archive
// reader the synthetic reference rate changes.
func TestFixtureServesBondIndices(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	cpi := gus.New(srv.Client(), srv.URL+GUSPath)
	var got []string
	for _, year := range []int{2022, 2023, 2024, 2025, 2026, 2027} {
		readings, err := cpi.CPI(t.Context(), year)
		if err != nil {
			t.Fatalf("CPI(%d): %v", year, err)
		}
		for _, r := range readings {
			got = append(got, describeReading(r))
		}
	}
	demoCPI := demo.CPIReadings()
	want := make([]string, 0, len(demoCPI))
	for _, r := range demoCPI {
		want = append(want, describeReading(r))
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CPI (-want +got):\n%s", diff)
	}
	changes, err := nbp.NewArchive(srv.Client(), srv.URL+NBPArchivePath).ReferenceRates(t.Context())
	if err != nil {
		t.Fatalf("ReferenceRates: %v", err)
	}
	rates := make([]string, 0, len(changes))
	for _, c := range changes {
		rates = append(rates, c.From.String()+" "+c.Rate.String())
	}
	if diff := cmp.Diff([]string{"2022-01-05 0.05", "2025-07-03 0.045"}, rates); diff != "" {
		t.Errorf("reference rates (-want +got):\n%s", diff)
	}
}

// describeReading writes a CPI reading's month and value.
func describeReading(r valuation.IndexReading) string {
	return time.Date(r.Year, r.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01") + " " + r.Value.String()
}
