package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// fakeLedger is app.Ledger over fixed records.
type fakeLedger struct {
	accounts []ledger.Account
	checks   map[ledger.AccountID][]ledgerapp.AssertionCheck
	txs      map[ledger.AccountID][]ledger.Transaction
	links    []ledger.TransferLink
}

func (f *fakeLedger) Accounts(context.Context) ([]ledger.Account, error) { return f.accounts, nil }

func (f *fakeLedger) AssertionChecks(_ context.Context, a ledger.AccountID) ([]ledgerapp.AssertionCheck, error) {
	return f.checks[a], nil
}

func (f *fakeLedger) CurrentTransactions(_ context.Context, a ledger.AccountID) ([]ledger.Transaction, error) {
	return f.txs[a], nil
}

func (f *fakeLedger) TransferLinks(context.Context) ([]ledger.TransferLink, error) {
	return f.links, nil
}

// fakeImports is app.Imports over fixed records.
type fakeImports struct {
	brokers []ingest.BrokerAccount
	tickers []ingest.PendingTicker
	batches []ingest.BatchProblems
}

func (f *fakeImports) AllBrokerAccounts(context.Context) ([]ingest.BrokerAccount, error) {
	return f.brokers, nil
}

func (f *fakeImports) PendingTickers(context.Context) ([]ingest.PendingTicker, error) {
	return f.tickers, nil
}

func (f *fakeImports) UnsupportedBatches(context.Context) ([]ingest.BatchProblems, error) {
	return f.batches, nil
}

// fakePriceChecks is app.PriceChecks over fixed checks.
type fakePriceChecks []valuation.PriceCheck

func (f fakePriceChecks) PriceChecks(context.Context) ([]valuation.PriceCheck, error) { return f, nil }

// fakeValues is app.Values with one wrapper per account.
type fakeValues struct {
	wrappers map[ledger.AccountID]jurisdiction.Wrapper
	settings jurisdiction.Settings
}

func (f *fakeValues) TaxWrapper(_ context.Context, a ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	w, ok := f.wrappers[a]
	return w, ok, nil
}

func (f *fakeValues) ValueSettings(context.Context) (jurisdiction.Settings, error) {
	return f.settings, nil
}

// day parses a date or fails t.
func day(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// pln returns amount in PLN.
func pln(t *testing.T, amount string) kernel.Money {
	t.Helper()
	d, err := kernel.ParseDecimal(amount)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", amount, err)
	}
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return kernel.NewMoney(d, c)
}

// key returns the test source's key id.
func key(id string) ledger.TransactionKey {
	return ledger.TransactionKey{Source: "test", ExternalID: id}
}

// cash returns a deposit, or a withdrawal for a negative amount, in account
// on date.
func cash(t *testing.T, account, id, date, amount string) ledger.Transaction {
	t.Helper()
	h := ledger.TransactionHeader{Key: key(id), Account: ledger.AccountID(account), Date: day(t, date)}
	m := pln(t, amount)
	var (
		tx  ledger.Transaction
		err error
	)
	if m.Sign() < 0 {
		tx, err = ledger.NewWithdrawal(h, m.Neg())
	} else {
		tx, err = ledger.NewDeposit(h, m)
	}
	if err != nil {
		t.Fatalf("building %s: %v", id, err)
	}
	return tx
}

func TestChecks(t *testing.T) {
	t.Parallel()
	const a1, a2 = ledger.AccountID("acc-1"), ledger.AccountID("acc-2")
	tests := []struct {
		name  string
		check func(t *testing.T) app.Check
		want  []attention.Condition
	}{
		{
			name: "flagged balance assertions",
			check: func(t *testing.T) app.Check {
				t.Helper()
				flagged := ledgerapp.AssertionCheck{
					Assertion: ledger.CashAssertion(a1, day(t, "2026-03-02"), pln(t, "100")), Status: ledgerapp.AssertionFlagged,
					Derived: pln(t, "90").Amount(),
				}
				matched := flagged
				matched.Status = ledgerapp.AssertionMatched
				entered := flagged
				entered.Assertion = ledger.CashCheckpoint(a1, day(t, "2026-03-03"), pln(t, "100"))
				return app.FlaggedAssertions(&fakeLedger{
					accounts: []ledger.Account{{ID: a1}},
					checks:   map[ledger.AccountID][]ledgerapp.AssertionCheck{a1: {flagged, matched, entered}},
				})
			},
			want: []attention.Condition{{Kind: attention.KindAssertionFlagged, Subject: "acc-1/2026-03-02/cash:PLN", Detail: "derived 90, reported 100"}},
		},
		{
			name: "cash checkpoints: only the owner's, flagged",
			check: func(t *testing.T) app.Check {
				t.Helper()
				entered := ledgerapp.AssertionCheck{
					Assertion: ledger.CashCheckpoint(a1, day(t, "2026-03-31"), pln(t, "105")), Status: ledgerapp.AssertionFlagged,
					Derived: pln(t, "100").Amount(),
				}
				matched := entered
				matched.Assertion.Date, matched.Status = day(t, "2026-04-30"), ledgerapp.AssertionMatched
				reported := entered
				reported.Assertion.Entered = false
				reported.Assertion.Date = day(t, "2026-03-02")
				return app.CashCheckpoints(&fakeLedger{
					accounts: []ledger.Account{{ID: a1}},
					checks:   map[ledger.AccountID][]ledgerapp.AssertionCheck{a1: {reported, entered, matched}},
				})
			},
			want: []attention.Condition{{Kind: attention.KindCashCheckpointMismatch, Subject: "acc-1/2026-03-31/PLN", Detail: "shown 105 PLN, derived 100 PLN, difference 5 PLN"}},
		},
		{
			name: "broken transfers",
			check: func(t *testing.T) app.Check {
				t.Helper()
				return app.BrokenTransfers(&fakeLedger{
					accounts: []ledger.Account{{ID: a1}, {ID: a2}},
					txs: map[ledger.AccountID][]ledger.Transaction{
						a1: {cash(t, "acc-1", "w", "2026-03-02", "-100"), cash(t, "acc-1", "w2", "2026-03-02", "-50")},
						a2: {cash(t, "acc-2", "d", "2026-03-03", "99.50"), cash(t, "acc-2", "d2", "2026-03-03", "50")},
					},
					links: []ledger.TransferLink{
						{Pair: ledger.TransferPair{Withdrawal: key("w"), Deposit: key("d")}, Decision: ledger.TransferConfirmed, Version: 1},
						{Pair: ledger.TransferPair{Withdrawal: key("w2"), Deposit: key("d2")}, Decision: ledger.TransferConfirmed, Version: 1},
					},
				})
			},
			want: []attention.Condition{{Kind: attention.KindTransferBroken, Subject: "test/w -> test/d", Detail: "confirmed as version 1; its legs no longer match"}},
		},
		{
			name: "flagged price checks: only an instrument's latest, unacknowledged",
			check: func(t *testing.T) app.Check {
				t.Helper()
				check := func(inst, date, source string, status valuation.PriceCheckStatus, acked bool) valuation.PriceCheck {
					return valuation.PriceCheck{
						CrossCheck: valuation.CrossCheckPrice{Instrument: inst, Date: day(t, date), Source: source},
						Difference: pln(t, "0.05").Amount(), Status: status, Acknowledged: acked,
					}
				}
				return app.FlaggedPriceChecks(fakePriceChecks{
					check("AAA", "2026-03-02", "ibkr", valuation.PriceFlagged, false),
					check("BBB", "2026-03-02", "ibkr", valuation.PriceFlagged, false),
					check("BBB", "2026-03-03", "ibkr", valuation.PriceMatched, false),
					check("CCC", "2026-03-03", "ibkr", valuation.PriceFlagged, true),
				})
			},
			want: []attention.Condition{{Kind: attention.KindPriceCheckFlagged, Subject: "AAA", Detail: "2026-03-02 ibkr: differs by 0.05"}},
		},
		{
			name: "unmapped broker accounts",
			check: func(t *testing.T) app.Check {
				t.Helper()
				return app.UnmappedBrokerAccounts(&fakeImports{brokers: []ingest.BrokerAccount{
					{Instance: "ibkr", ID: "U0000001", LastSeen: day(t, "2026-03-02")},
					{Instance: "ibkr", ID: "U0000002", Account: a1},
				}})
			},
			want: []attention.Condition{{Kind: attention.KindBrokerAccountUnmapped, Subject: "ibkr/U0000001", Detail: "last seen 2026-03-02"}},
		},
		{
			name: "unconfirmed tickers",
			check: func(*testing.T) app.Check {
				return app.UnconfirmedTickers(&fakeImports{tickers: []ingest.PendingTicker{{Instance: "xtb", Ticker: ingest.Ticker{Symbol: "AAA.US"}, Waiting: 2}}})
			},
			want: []attention.Condition{{Kind: attention.KindTickerUnconfirmed, Subject: "xtb/AAA.US", Detail: "2 rows wait"}},
		},
		{
			name: "import problems",
			check: func(*testing.T) app.Check {
				return app.ImportProblems(&fakeImports{batches: []ingest.BatchProblems{{Batch: "b-1", Instance: "ibkr", Unsupported: 3}}})
			},
			want: []attention.Condition{{Kind: attention.KindImportProblems, Subject: "b-1", Detail: "ibkr: 3 rows left out"}},
		},
		{
			name: "the marginal rate is missing while an IKZE account needs it",
			check: func(*testing.T) app.Check {
				return app.MissingSettings(&fakeLedger{accounts: []ledger.Account{{ID: a1}, {ID: a2}}},
					&fakeValues{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{a2: "ikze"}}, pl.NewPack())
			},
			want: []attention.Condition{{Kind: attention.KindSettingMissing, Subject: app.MarginalRateSetting, Detail: "the liquidation value of an account's tax wrapper needs it"}},
		},
		{
			name: "the marginal rate set",
			check: func(t *testing.T) app.Check {
				t.Helper()
				return app.MissingSettings(&fakeLedger{accounts: []ledger.Account{{ID: a2}}},
					&fakeValues{
						wrappers: map[ledger.AccountID]jurisdiction.Wrapper{a2: "ikze"},
						settings: jurisdiction.Settings{MarginalRate: pln(t, "0.32").Amount(), HasMarginalRate: true},
					}, pl.NewPack())
			},
		},
		{
			name: "no IKZE account, no missing setting",
			check: func(*testing.T) app.Check {
				return app.MissingSettings(&fakeLedger{accounts: []ledger.Account{{ID: a1}}}, &fakeValues{}, pl.NewPack())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.check(t)
			got, err := c.Find(t.Context())
			if err != nil {
				t.Fatalf("Find: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Find (-want +got):\n%s", diff)
			}
			for _, cond := range got {
				if cond.Kind != c.Kind || cond.Validate() != nil {
					t.Errorf("condition %+v doesn't fit the check's kind %s", cond, c.Kind)
				}
			}
		})
	}
}
