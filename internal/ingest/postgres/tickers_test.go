package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// tickerBuy is account's buy of 4 units of symbol for 400 PLN, keyed id.
func tickerBuy(t *testing.T, account ledger.AccountID, symbol, id string) ingest.TickerEntry {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	day, err := kernel.ParseDate("2026-01-05")
	if err != nil {
		t.Fatal(err)
	}
	return ingest.TickerEntry{
		Symbol: symbol, Kind: ledger.KindBuy, Quantity: kernel.NewQuantity(kernel.NewDecimal(4, 0)),
		Amount: kernel.NewMoney(kernel.NewDecimal(400, 0), pln), Description: "buy " + symbol, Ref: "Cash Operations#7",
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb:" + string(account), ExternalID: id}, Account: account, Date: day},
	}
}

// tickerHolding is account's holding of 4 units of symbol at the end of
// 2026-06-30 (D-155).
func tickerHolding(t *testing.T, account ledger.AccountID, symbol string) ingest.TickerAssertion {
	t.Helper()
	day, err := kernel.ParseDate("2026-06-30")
	if err != nil {
		t.Fatal(err)
	}
	return ingest.TickerAssertion{Symbol: symbol, Account: account, Date: day, Quantity: kernel.NewQuantity(kernel.NewDecimal(4, 0)), Ref: "Open Positions#10"}
}

// rowKeys names a booking's rows: each entry's instance and key, each
// assertion's instance, account and ticker.
func rowKeys(rows []ingest.TickerRows) []string {
	var out []string
	for i := range rows {
		for j := range rows[i].Entries {
			out = append(out, rows[i].Instance+" "+rows[i].Entries[j].Header.Key.String())
		}
		for _, a := range rows[i].Assertions {
			out = append(out, rows[i].Instance+" holding "+string(a.Account)+" "+a.Symbol)
		}
	}
	return out
}

// TestTickersPerInstitution checks D-153 and D-155 over PostgreSQL: two
// XTB instances report TESTA.PL with a buy and a holding each; one
// confirmation, through either, holds for both and returns every waiting
// row of both, which then count as booked; the pending list drops the
// ticker for both; a second confirmation with another ISIN is refused; a
// correction records a new version and returns every row, booked or not,
// with the old listing; and the confirmed list shows the current ISIN.
func TestTickersPerInstitution(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	holdAtTwoInstances(t, s)
	pending, err := s.PendingTickers(ctx)
	if err != nil || len(pending) != 2 || pending[0].Waiting != 1 {
		t.Fatalf("PendingTickers = %+v, %v; want TESTA.PL waiting at both instances", pending, err)
	}
	checkAccounts(t, s, false, "acc-1", "acc-2")
	isinA, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	booking, err := s.ConfirmTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-ike", Symbol: "TESTA.PL", ISIN: isinA})
	if err != nil {
		t.Fatalf("ConfirmTicker: %v", err)
	}
	want := []string{"xtb-ike xtb:acc-2/9000000005", "xtb-ike holding acc-2 TESTA.PL", "xtb-main xtb:acc-1/9000000002", "xtb-main holding acc-1 TESTA.PL"}
	if diff := cmp.Diff(want, rowKeys(booking.Rows)); diff != "" || booking.Listing.ISIN != isinA || booking.Listing.Currency.String() != "PLN" {
		t.Errorf("ConfirmTicker booked %v in %+v (-want +got):\n%s", booking.Rows, booking.Listing, diff)
	}
	if err := s.BookedTickerRows(ctx, booking.Rows); err != nil {
		t.Fatalf("BookedTickerRows: %v", err)
	}
	checkConfirmedOnce(t, s, isinA)
	checkCorrection(t, s, isinA, want)
}

// holdAtTwoInstances holds a buy and a holding of TESTA.PL at xtb-main in
// acc-1 and at xtb-ike in acc-2, both XTB's.
func holdAtTwoInstances(t *testing.T, s *postgres.Store) {
	t.Helper()
	testa := ingest.Ticker{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL", Name: "Test Alpha SA"}
	for _, inst := range []struct {
		instance string
		account  ledger.AccountID
		id       string
	}{{"xtb-main", "acc-1", "9000000002"}, {"xtb-ike", "acc-2", "9000000005"}} {
		added, err := s.HoldTickers(t.Context(), &ingest.TickerHold{
			Institution: "xtb", Tickers: []ingest.Ticker{testa},
			Waiting: ingest.TickerRows{
				Instance: inst.instance, Entries: []ingest.TickerEntry{tickerBuy(t, inst.account, "TESTA.PL", inst.id)},
				Assertions: []ingest.TickerAssertion{tickerHolding(t, inst.account, "TESTA.PL")},
			},
		})
		if err != nil || len(added) != 1 {
			t.Fatalf("HoldTickers(%s) = %v, %v; want TESTA.PL added", inst.instance, added, err)
		}
	}
}

// checkConfirmedOnce checks that TESTA.PL, confirmed as isinA through
// xtb-ike, is confirmed at xtb for both instances, waits for nothing, and
// refuses another ISIN.
func checkConfirmedOnce(t *testing.T, s *postgres.Store, isinA ledger.ISIN) {
	t.Helper()
	ctx := t.Context()
	pending, err := s.PendingTickers(ctx)
	if err != nil || len(pending) != 0 {
		t.Errorf("PendingTickers after the confirmation = %+v, %v; want none", pending, err)
	}
	confirmed, err := s.ConfirmedTickers(ctx, "xtb")
	if err != nil || confirmed["TESTA.PL"].ISIN != isinA {
		t.Errorf("ConfirmedTickers(xtb) = %v, %v; want TESTA.PL as %s", confirmed, err, isinA)
	}
	again, err := s.ConfirmTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isinA})
	if err != nil || len(again.Rows) != 0 {
		t.Errorf("confirming again = %+v, %v; want nothing left to book", again, err)
	}
	isinC, err := ledger.ParseISIN("PLTESTC00008")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConfirmTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isinC}); !errors.Is(err, app.ErrTickerConfirmed) {
		t.Errorf("confirming another ISIN = %v, want ErrTickerConfirmed", err)
	}
	if _, err := s.ConfirmTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTZ.PL", ISIN: isinC}); !errors.Is(err, app.ErrUnknownTicker) {
		t.Errorf("confirming an unknown ticker = %v, want ErrUnknownTicker", err)
	}
}

// checkCorrection corrects TESTA.PL from isinA to PLTESTC00008 and checks
// that every row is returned to book anew, with the old listing, and that
// the confirmed list shows the corrected ISIN with both rows booked.
func checkCorrection(t *testing.T, s *postgres.Store, isinA ledger.ISIN, rows []string) {
	t.Helper()
	ctx := t.Context()
	isinC, err := ledger.ParseISIN("PLTESTC00008")
	if err != nil {
		t.Fatal(err)
	}
	if _, unknownErr := s.CorrectTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTB.PL", ISIN: isinC}); !errors.Is(unknownErr, app.ErrUnknownTicker) {
		t.Errorf("correcting an unknown ticker = %v, want ErrUnknownTicker", unknownErr)
	}
	booking, err := s.CorrectTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isinC})
	if err != nil {
		t.Fatalf("CorrectTicker: %v", err)
	}
	if diff := cmp.Diff(rows, rowKeys(booking.Rows)); diff != "" || booking.Listing.ISIN != isinC || booking.Previous.ISIN != isinA {
		t.Errorf("CorrectTicker booked %+v, %+v (-want +got):\n%s", booking.Listing, booking.Previous, diff)
	}
	checkConfirmedList(t, s, isinC)
	checkAccounts(t, s, true, "acc-1", "acc-2")
	checkAccounts(t, s, false)
}

// checkAccounts checks the accounts of TESTA.PL's rows at XTB, all or
// those waiting.
func checkAccounts(t *testing.T, s *postgres.Store, all bool, want ...ledger.AccountID) {
	t.Helper()
	got, err := s.TickerAccounts(t.Context(), "xtb-main", "TESTA.PL", all)
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); err != nil || diff != "" {
		t.Errorf("TickerAccounts(all %t) = %v, %v (-want +got):\n%s", all, got, err, diff)
	}
}

// checkConfirmedList checks that the confirmed list shows TESTA.PL once,
// at XTB, as isin, with both its rows booked.
func checkConfirmedList(t *testing.T, s *postgres.Store, isin ledger.ISIN) {
	t.Helper()
	list, err := s.ConfirmedTickerList(t.Context())
	if err != nil || len(list) != 1 {
		t.Fatalf("ConfirmedTickerList = %+v, %v; want TESTA.PL once", list, err)
	}
	want := ingest.ConfirmedTicker{
		Institution: "xtb", Instance: "xtb-ike", ISIN: isin, TradingCurrency: "PLN", Booked: 2,
		Ticker: ingest.Ticker{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL", Name: "Test Alpha SA"},
	}
	if diff := cmp.Diff(want, list[0]); diff != "" {
		t.Errorf("confirmed ticker (-want +got):\n%s", diff)
	}
}

// TestCorrectingAnUnconfirmedTicker checks that a ticker never confirmed
// can't be corrected.
func TestCorrectingAnUnconfirmedTicker(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	if _, err := s.HoldTickers(ctx, &ingest.TickerHold{
		Institution: "xtb", Tickers: []ingest.Ticker{{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL"}},
		Waiting: ingest.TickerRows{Instance: "xtb-main"},
	}); err != nil {
		t.Fatalf("HoldTickers: %v", err)
	}
	isinA, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CorrectTicker(ctx, ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isinA}); !errors.Is(err, ingest.ErrTickerNotConfirmed) {
		t.Errorf("CorrectTicker = %v, want ErrTickerNotConfirmed", err)
	}
}
