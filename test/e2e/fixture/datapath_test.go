package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// dataPathAccount is the ledger account the data path's statements are
// mapped into here; the core names its own.
const dataPathAccount = "acc-datapath"

// TestFixtureServesDataPath checks that the connectors' own Flex Web
// Service client gets each data path statement for its query ID, and that
// the IBKR Flex mapper books it as the end-to-end data path expects: the
// opening statement seeds three opening lots, with their open dates and
// costs, and the cash of two currencies, and gives two cross-check prices;
// the week books a deposit, a buy, a dividend with tax withheld and a fee
// after it; the corrected week changes only the deposit's amount, under the
// same key.
func TestFixtureServesDataPath(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	tests := []struct {
		query      string
		mode       ingest.Mode
		want       []string
		wantChecks []string
	}{
		{
			query: DataPathOpening, mode: ingest.ModeOpening,
			want: []string{
				"deposit opening-cash:CHF:2026-03-06 2026-03-06: cash 100 CHF; equity -100 CHF",
				"deposit opening-cash:GBP:2026-03-06 2026-03-06: cash 250 GBP; equity -250 GBP",
				"opening-lot opening-lot:CHTEST000052:2026-01-20:9100103 2026-03-06 opened 2026-01-20: holding CHTEST000052/TSTH 800 CHF 8; equity -800 CHF",
				"opening-lot opening-lot:GBTEST000047:2026-02-10:9100101 2026-03-06 opened 2026-02-10: holding GBTEST000047/TSTG 1000 GBP 10; equity -1000 GBP",
				"opening-lot opening-lot:GBTEST000047:2026-02-16:9100102 2026-03-06 opened 2026-02-16: holding GBTEST000047/TSTG 520 GBP 5; equity -520 GBP",
			},
			wantChecks: []string{"GBTEST000047/TSTG 2026-03-06 105.5 GBP", "CHTEST000052/TSTH 2026-03-06 104 CHF"},
		},
		{
			query: DataPathWeek, mode: ingest.ModeTransactions,
			want:       weekBooked("1000"),
			wantChecks: []string{"GBTEST000047/TSTG 2026-03-13 105 GBP", "CHTEST000052/TSTH 2026-03-13 102 CHF"},
		},
		{
			query: DataPathCorrected, mode: ingest.ModeTransactions,
			want:       weekBooked("1100"),
			wantChecks: []string{"GBTEST000047/TSTG 2026-03-13 105 GBP", "CHTEST000052/TSTH 2026-03-13 102 CHF"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			t.Parallel()
			m := mapDataPath(t, srv, tt.query, tt.mode)
			if len(m.Problems) > 0 {
				t.Errorf("problems: %+v", m.Problems)
			}
			if diff := cmp.Diff(tt.want, describeTransactions(m.Transactions)); diff != "" {
				t.Errorf("transactions (-want +got):\n%s", diff)
			}
			checks := make([]string, 0, len(m.CrossCheckPrices))
			for _, p := range m.CrossCheckPrices {
				checks = append(checks, fmt.Sprintf("%s %s %s", p.Instrument, p.Date, p.Close))
			}
			if diff := cmp.Diff(tt.wantChecks, checks); diff != "" {
				t.Errorf("cross-check prices (-want +got):\n%s", diff)
			}
		})
	}
}

// mapDataPath fetches the statement of query from srv with the Flex Web
// Service client and maps it in mode into dataPathAccount.
func mapDataPath(t *testing.T, srv *httptest.Server, query string, mode ingest.Mode) ingest.Mapping {
	t.Helper()
	c := ibkrflex.Client{
		HTTP: srv.Client(), Token: flexAccepted, QueryID: query,
		SendRequestURL:  srv.URL + FlexPath + "/SendRequest",
		GetStatementURL: srv.URL + FlexPath + "/GetStatement",
		Wait:            func(context.Context, int) error { return nil },
	}
	statement, err := c.FetchStatement(t.Context())
	if err != nil {
		t.Fatalf("FetchStatement: %v", err)
	}
	m, err := ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{"U0000002": dataPathAccount}}.Map(bytes.NewReader(statement), mode)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	return m
}

// weekBooked is how the week's transactions are booked, with the
// deposit's amount.
func weekBooked(deposit string) []string {
	return []string{
		"buy 9100201 2026-03-10: holding GBTEST000047/TSTG 208 GBP 2; cash -208 GBP; fee 1 GBP; cash -1 GBP",
		"deposit 8100201 2026-03-09: cash " + deposit + " GBP; equity -" + deposit + " GBP",
		"dividend 8100202 2026-03-11: holding CHTEST000052/TSTH -12 CHF; withholding-tax 4.2 CHF; cash 7.8 CHF",
		"fee 8100204 2026-03-12: fee 3 GBP; cash -3 GBP",
	}
}

// describeTransactions writes each transaction as its kind, external ID,
// date, open date and postings, in text order.
func describeTransactions(txs []ledger.Transaction) []string {
	out := make([]string, 0, len(txs))
	for i := range txs {
		tx := &txs[i]
		head := fmt.Sprintf("%s %s %s", tx.Kind, tx.Key.ExternalID, tx.Date)
		if !tx.LotOpened.IsZero() {
			head += " opened " + tx.LotOpened.String()
		}
		postings := make([]string, 0, len(tx.Postings))
		for _, p := range tx.Postings {
			s := describeBook(p.Book) + " " + p.Amount.String()
			if !p.Quantity.IsZero() {
				s += " " + p.Quantity.String()
			}
			postings = append(postings, s)
		}
		out = append(out, head+": "+strings.Join(postings, "; "))
	}
	slices.Sort(out)
	return out
}

// describeBook names a ledger account by its kind, and a holding's by its
// instrument.
func describeBook(b ledger.Book) string {
	if b.Kind == ledger.BookHolding {
		return "holding " + b.Instrument.String()
	}
	return b.Kind.String()
}

// TestFixtureServesDataPathRates checks that the connectors' own NBP
// client reads the data path's day-by-day rates of GBP, CHF and NOK, and
// only those, below the data path's path.
func TestFixtureServesDataPathRates(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	rates, err := nbp.New(srv.Client(), srv.URL+NBPDataPathPath).Rates(t.Context(), day(t, "2026-02-13"), day(t, "2026-02-16"))
	if err != nil {
		t.Fatalf("Rates: %v", err)
	}
	got := make([]string, 0, len(rates))
	for _, r := range rates {
		got = append(got, r.Date.String()+" "+r.Currency.String()+" "+r.Rate.String())
	}
	want := []string{
		"2026-02-13 GBP 5.3", "2026-02-13 CHF 4.8", "2026-02-13 NOK 0.39",
		"2026-02-16 GBP 5.31", "2026-02-16 CHF 4.81", "2026-02-16 NOK 0.391",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rates (-want +got):\n%s", diff)
	}
}

// TestFixtureServesDataPathListings checks that EODHD's search lists the
// data path's instruments by ISIN and by code on an exchange, with closes.
func TestFixtureServesDataPathListings(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	c := eodhd.New(srv.Client(), srv.URL+EODHDPath, eodhdAccepted)
	isin, err := ledger.ParseISIN("GBTEST000047")
	if err != nil {
		t.Fatal(err)
	}
	listed, err := c.Listings(t.Context(), isin)
	byCode, codeErr := c.Search(t.Context(), "TSTJ", eodhd.CountryExchange("NO"))
	closes, closesErr := c.Closes(t.Context(), "TSTH.SW", day(t, "2026-03-06"), day(t, "2026-03-06"))
	got := make([]string, 0, len(listed)+len(byCode)+len(closes))
	for _, l := range listed {
		got = append(got, "by ISIN: "+l.Code+"."+l.Exchange)
	}
	for _, l := range byCode {
		got = append(got, "by code: "+l.ISIN.String())
	}
	for _, cl := range closes {
		got = append(got, "close: "+cl.Date.String()+" "+cl.Close.String())
	}
	want := []string{"by ISIN: TSTG.LSE", "by code: NOTEST000068", "close: 2026-03-06 102"}
	if diff := cmp.Diff(want, got); diff != "" || errors.Join(err, codeErr, closesErr) != nil {
		t.Errorf("EODHD (-want +got):\n%s\nerrors: %v", diff, errors.Join(err, codeErr, closesErr))
	}
}
