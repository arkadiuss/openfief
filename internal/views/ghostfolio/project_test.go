package ghostfolio_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// TestDerivedUUIDs pins the UUIDs the view derives: an owned account's from
// the instance and the account, a bond's from the instance, the Ghostfolio
// user and its series code, a fee asset's from the instance, the user and
// its currency (D-64, D-108). They are UUID version 5 in OpenFief's
// namespace, checked against Python's uuid.uuid5, and must never change: a
// new UUID would orphan every owned account, and move every activity to a
// new asset.
func TestDerivedUUIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, got, want string
	}{
		{"owned account", ghostfolio.AccountUUID("ghostfolio", "acct-ibkr"), "a013b834-4bd6-5aaf-9784-769143502838"},
		{"bond series", bondSym("EDO0536"), "28c758ef-4bec-5db3-9fc5-6fd61c14f623"},
		{"fee asset", feeSym(t, "PLN"), "e48bff40-0f25-5948-b8a8-7a49ea3193ac"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %s, want %s", tt.got, tt.want)
			}
		})
	}
	if ghostfolio.AccountUUID("other", "acct-ibkr") == ghostfolio.AccountUUID("ghostfolio", "acct-ibkr") {
		t.Error("two instances share an owned account")
	}
	const otherUser = "5a1e0f6c-0000-4000-8000-00000000beef"
	for _, other := range []string{
		ghostfolio.BondSymbol("other", gfUser, "EDO0536"), ghostfolio.BondSymbol(instance, otherUser, "EDO0536"),
		ghostfolio.FeeSymbol("other", gfUser, cur(t, "PLN")), ghostfolio.FeeSymbol(instance, otherUser, cur(t, "PLN")),
	} {
		if other == bondSym("EDO0536") || other == feeSym(t, "PLN") {
			t.Errorf("another instance or user shares the asset %s", other)
		}
	}
}

// TestProject checks how each transaction kind becomes activities (D-4,
// D-64), on one account with synthetic instruments.
func TestProject(t *testing.T) {
	t.Parallel()
	must := mustTx(t)
	etf, other, bond := instrument(t, isinIE, "ETF1"), instrument(t, isinXS, "ETF2"), instrument(t, isinPL, "EDO0136")
	zero := money(t, "0", "USD")
	edo, feeEUR, feePLN := bondSym("EDO0136"), feeSym(t, "EUR"), feeSym(t, "PLN")
	tests := []struct {
		name         string
		txs          []ledger.Transaction
		vals         []ghostfolio.Valuation
		want         []ghostfolio.Activity
		wantAssets   []ghostfolio.AssetProfile
		wantProblems []ghostfolio.Problem
	}{
		{
			name: "a buy with a fee in its currency",
			txs:  []ledger.Transaction{must(ledger.NewBuy(header(t, "a1", "b1", "2025-01-03"), etf, qty(t, "3"), money(t, "1261.50", "USD"), money(t, "1.25", "USD")))},
			want: []ghostfolio.Activity{activity(t, "a1", "b1", ghostfolio.TypeBuy, "2025-01-03", "USD", yahoo, "ETF1.L", "3", "420.5", "1.25")},
		},
		{
			name: "a fee in another currency becomes a FEE of its own",
			txs:  []ledger.Transaction{must(ledger.NewBuy(header(t, "a1", "b2", "2025-01-03"), etf, qty(t, "2"), money(t, "100", "USD"), money(t, "2", "EUR")))},
			want: []ghostfolio.Activity{
				activity(t, "a1", "b2", ghostfolio.TypeBuy, "2025-01-03", "USD", yahoo, "ETF1.L", "2", "50", "0"),
				func() ghostfolio.Activity {
					a := activity(t, "a1", "b2", ghostfolio.TypeFee, "2025-01-03", "EUR", ghostfolio.DataSourceManual, feeEUR, "0", "0", "2")
					a.Comment += "#fee:EUR"
					return a
				}(),
			},
			wantAssets: []ghostfolio.AssetProfile{{Symbol: feeEUR, Name: "OpenFief fees (EUR)", Currency: cur(t, "EUR")}},
		},
		{
			name: "a sell at the proceeds' unit price",
			txs:  []ledger.Transaction{must(ledger.NewSell(header(t, "a1", "s1", "2025-02-03"), etf, qty(t, "2"), money(t, "900", "USD"), money(t, "1", "USD")))},
			want: []ghostfolio.Activity{activity(t, "a1", "s1", ghostfolio.TypeSell, "2025-02-03", "USD", yahoo, "ETF1.L", "2", "450", "1")},
		},
		{
			name: "an opening lot is a buy dated when the lot was opened",
			txs:  []ledger.Transaction{must(ledger.NewOpeningLot(header(t, "a1", "o1", "2025-01-02"), etf, qty(t, "10"), money(t, "1000", "USD"), date(t, "2023-05-04")))},
			want: []ghostfolio.Activity{activity(t, "a1", "o1", ghostfolio.TypeBuy, "2023-05-04", "USD", yahoo, "ETF1.L", "10", "100", "0")},
		},
		{
			name: "a dividend of quantity 1 at the gross amount, the tax withheld as its fee",
			txs:  []ledger.Transaction{must(ledger.NewDividend(header(t, "a1", "d1", "2025-03-03"), etf, money(t, "12.50", "USD"), money(t, "1.88", "USD")))},
			want: []ghostfolio.Activity{activity(t, "a1", "d1", ghostfolio.TypeDividend, "2025-03-03", "USD", yahoo, "ETF1.L", "1", "12.5", "1.88")},
		},
		{
			name:       "a fee on its own, against the fee asset of its currency",
			txs:        []ledger.Transaction{must(ledger.NewFee(header(t, "a1", "f1", "2025-03-04"), money(t, "5", "PLN")))},
			want:       []ghostfolio.Activity{activity(t, "a1", "f1", ghostfolio.TypeFee, "2025-03-04", "PLN", ghostfolio.DataSourceManual, feePLN, "0", "0", "5")},
			wantAssets: []ghostfolio.AssetProfile{{Symbol: feePLN, Name: "OpenFief fees (PLN)", Currency: cur(t, "PLN")}},
		},
		{
			name: "cash movements have no activity",
			txs: []ledger.Transaction{
				must(ledger.NewDeposit(header(t, "a1", "c1", "2025-01-02"), money(t, "1000", "USD"))),
				must(ledger.NewWithdrawal(header(t, "a1", "c2", "2025-01-03"), money(t, "10", "USD"))),
				must(ledger.NewFXConversion(header(t, "a1", "c3", "2025-01-04"), money(t, "10", "USD"), money(t, "9", "EUR"), zero)),
			},
		},
		{
			name: "an instrument with no listing is a problem",
			txs:  []ledger.Transaction{must(ledger.NewBuy(header(t, "a1", "u1", "2025-01-03"), other, qty(t, "1"), money(t, "10", "USD"), zero))},
			wantProblems: []ghostfolio.Problem{
				{Transaction: ledger.TransactionKey{Source: "test/a1", ExternalID: "u1"}, Reason: ghostfolio.ProblemUnlisted},
			},
		},
		{
			name: "a unit price rounded to 15 significant digits",
			txs:  []ledger.Transaction{must(ledger.NewBuy(header(t, "a1", "r1", "2025-01-03"), etf, qty(t, "3"), money(t, "100", "USD"), zero))},
			want: []ghostfolio.Activity{activity(t, "a1", "r1", ghostfolio.TypeBuy, "2025-01-03", "USD", yahoo, "ETF1.L", "3", "33.3333333333333", "0")},
		},
		{
			name: "a bond is a MANUAL asset priced by its valuations",
			txs: []ledger.Transaction{
				must(ledger.NewBuy(header(t, "a1", "e1", "2025-01-02"), bond, qty(t, "10"), money(t, "1000", "PLN"), money(t, "0", "PLN"))),
				must(ledger.NewBuy(header(t, "a1", "e2", "2025-01-04"), bond, qty(t, "10"), money(t, "1000", "PLN"), money(t, "0", "PLN"))),
			},
			vals: []ghostfolio.Valuation{
				{Date: date(t, "2025-01-03"), Instrument: bond, Value: money(t, "1000.27", "PLN")},
				{Date: date(t, "2025-01-02"), Instrument: bond, Value: money(t, "1000.00", "PLN")},
				{Date: date(t, "2025-01-04"), Instrument: bond, Value: money(t, "2000.60", "PLN")},
			},
			want: []ghostfolio.Activity{
				activity(t, "a1", "e1", ghostfolio.TypeBuy, "2025-01-02", "PLN", ghostfolio.DataSourceManual, edo, "10", "100", "0"),
				activity(t, "a1", "e2", ghostfolio.TypeBuy, "2025-01-04", "PLN", ghostfolio.DataSourceManual, edo, "10", "100", "0"),
			},
			wantAssets: []ghostfolio.AssetProfile{{
				Symbol: edo, Name: "EDO0136", ISIN: isinPL, Currency: cur(t, "PLN"), AssetClass: "FIXED_INCOME", AssetSubClass: "BOND",
				MarketData: []ghostfolio.Price{
					{Date: date(t, "2025-01-02"), Price: dec(t, "100")},
					{Date: date(t, "2025-01-03"), Price: dec(t, "100.027")},
					{Date: date(t, "2025-01-04"), Price: dec(t, "100.03")},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := ghostfolio.Source{
				Instance: instance, User: gfUser, Currency: cur(t, "PLN"),
				Accounts: []ghostfolio.SourceAccount{{ID: "a1", Name: "Synthetic broker", Transactions: tt.txs, Valuations: tt.vals}},
				Listings: map[ledger.ISIN]ghostfolio.Listing{etf.ISIN: {DataSource: yahoo, Symbol: "ETF1.L"}},
			}
			got, problems := ghostfolio.Project(&src)
			want := ghostfolio.Desired{
				Accounts:   []ghostfolio.Account{{ID: ghostfolio.AccountUUID(instance, "a1"), Name: "Synthetic broker", Currency: cur(t, "PLN")}},
				Activities: tt.want,
				Assets:     tt.wantAssets,
			}
			// TestProjectBalances checks the cash balances.
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(ghostfolio.Desired{}, "Balances")); diff != "" {
				t.Errorf("Project (-want +got):\n%s", diff)
			}
			problems = slices.DeleteFunc(problems, func(p ghostfolio.Problem) bool { return p.Reason == ghostfolio.ProblemCashCurrency })
			if diff := cmp.Diff(tt.wantProblems, problems, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("problems (-want +got):\n%s", diff)
			}
		})
	}
}

// TestProjectAccounts checks that every account of the record gets its own
// owned account, in ID order, even with no activity, and that activities
// come in account, then comment order.
func TestProjectAccounts(t *testing.T) {
	t.Parallel()
	must := mustTx(t)
	etf := instrument(t, isinIE, "ETF1")
	zero := money(t, "0", "EUR")
	src := ghostfolio.Source{
		Instance: instance, Currency: cur(t, "EUR"),
		Accounts: []ghostfolio.SourceAccount{
			{ID: "b", Name: "Second", Transactions: []ledger.Transaction{
				must(ledger.NewBuy(header(t, "b", "2", "2025-01-02"), etf, qty(t, "1"), money(t, "1", "EUR"), zero)),
				must(ledger.NewBuy(header(t, "b", "1", "2025-01-03"), etf, qty(t, "1"), money(t, "1", "EUR"), zero)),
			}},
			{ID: "a", Name: "First"},
		},
		Listings: map[ledger.ISIN]ghostfolio.Listing{etf.ISIN: {DataSource: yahoo, Symbol: "ETF1.L"}},
	}
	got, _ := ghostfolio.Project(&src)
	wantAccounts := []ghostfolio.Account{
		{ID: ghostfolio.AccountUUID(instance, "a"), Name: "First", Currency: cur(t, "EUR")},
		{ID: ghostfolio.AccountUUID(instance, "b"), Name: "Second", Currency: cur(t, "EUR")},
	}
	if diff := cmp.Diff(wantAccounts, got.Accounts); diff != "" {
		t.Errorf("accounts (-want +got):\n%s", diff)
	}
	comments := make([]string, 0, len(got.Activities))
	for _, a := range got.Activities {
		comments = append(comments, a.Comment)
	}
	if diff := cmp.Diff([]string{comment("b", "1"), comment("b", "2")}, comments); diff != "" {
		t.Errorf("activity order (-want +got):\n%s", diff)
	}
}

// TestProjectBalances checks the owned accounts' cash balances (D-108): the
// account's cash in the owned account's currency at the end of each date it
// changes, cash in another currency left out as a problem, and a balance
// too large for a double leaving out the account's balances.
func TestProjectBalances(t *testing.T) {
	t.Parallel()
	must := mustTx(t)
	etf, unlisted := instrument(t, isinIE, "ETF1"), instrument(t, isinXS, "ETF2")
	a1 := ghostfolio.AccountUUID(instance, "a1")
	bal := func(day, value string) ghostfolio.Balance {
		return ghostfolio.Balance{AccountID: a1, Date: date(t, day), Value: dec(t, value)}
	}
	tests := []struct {
		name         string
		txs          []ledger.Transaction
		want         []ghostfolio.Balance
		wantProblems []ghostfolio.Problem
	}{
		{name: "no cash, no balance"},
		{
			name: "one balance per date the cash changes, in date order",
			txs: []ledger.Transaction{
				must(ledger.NewBuy(header(t, "a1", "b1", "2025-01-03"), etf, qty(t, "2"), money(t, "200", "PLN"), money(t, "1.5", "PLN"))),
				must(ledger.NewDeposit(header(t, "a1", "d1", "2025-01-02"), money(t, "1000", "PLN"))),
				must(ledger.NewDividend(header(t, "a1", "v1", "2025-01-07"), etf, money(t, "10", "PLN"), money(t, "1.9", "PLN"))),
			},
			want: []ghostfolio.Balance{bal("2025-01-02", "1000"), bal("2025-01-03", "798.5"), bal("2025-01-07", "806.6")},
		},
		{
			name: "two transactions on one date make one balance; a date that nets to nothing makes none",
			txs: []ledger.Transaction{
				must(ledger.NewDeposit(header(t, "a1", "d1", "2025-01-02"), money(t, "100", "PLN"))),
				must(ledger.NewDeposit(header(t, "a1", "d2", "2025-01-02"), money(t, "50", "PLN"))),
				must(ledger.NewDeposit(header(t, "a1", "d3", "2025-01-03"), money(t, "20", "PLN"))),
				must(ledger.NewWithdrawal(header(t, "a1", "w1", "2025-01-03"), money(t, "20", "PLN"))),
				must(ledger.NewWithdrawal(header(t, "a1", "w2", "2025-01-04"), money(t, "150", "PLN"))),
			},
			want: []ghostfolio.Balance{bal("2025-01-02", "150"), bal("2025-01-04", "0")},
		},
		{
			name: "cash in another currency is left out as a problem",
			txs: []ledger.Transaction{
				must(ledger.NewDeposit(header(t, "a1", "d1", "2025-01-02"), money(t, "100", "PLN"))),
				must(ledger.NewFXConversion(header(t, "a1", "x1", "2025-01-03"), money(t, "40", "PLN"), money(t, "10", "USD"), money(t, "0", "PLN"))),
				must(ledger.NewDeposit(header(t, "a1", "d2", "2025-01-04"), money(t, "5", "EUR"))),
			},
			want: []ghostfolio.Balance{bal("2025-01-02", "100"), bal("2025-01-03", "60")},
			wantProblems: []ghostfolio.Problem{
				{Account: "a1", Currency: cur(t, "EUR"), Reason: ghostfolio.ProblemCashCurrency},
				{Account: "a1", Currency: cur(t, "USD"), Reason: ghostfolio.ProblemCashCurrency},
			},
		},
		{
			name: "a transaction left out still moves the cash",
			txs: []ledger.Transaction{
				must(ledger.NewBuy(header(t, "a1", "u1", "2025-01-03"), unlisted, qty(t, "1"), money(t, "10", "PLN"), money(t, "0", "PLN"))),
			},
			want: []ghostfolio.Balance{bal("2025-01-03", "-10")},
			wantProblems: []ghostfolio.Problem{
				{Transaction: ledger.TransactionKey{Source: "test/a1", ExternalID: "u1"}, Reason: ghostfolio.ProblemUnlisted},
			},
		},
		{
			name: "a balance too large for a double leaves out the account's balances",
			txs: []ledger.Transaction{
				must(ledger.NewDeposit(header(t, "a1", "d1", "2025-01-02"), money(t, "1", "PLN"))),
				must(ledger.NewDeposit(header(t, "a1", "d2", "2025-01-03"), money(t, "1234567890123456789", "PLN"))),
			},
			wantProblems: []ghostfolio.Problem{{Account: "a1", Reason: ghostfolio.ProblemTooLarge}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := ghostfolio.Source{
				Instance: instance, User: gfUser, Currency: cur(t, "PLN"),
				Accounts: []ghostfolio.SourceAccount{{ID: "a1", Transactions: tt.txs}},
				Listings: map[ledger.ISIN]ghostfolio.Listing{etf.ISIN: {DataSource: yahoo, Symbol: "ETF1.L"}},
			}
			got, problems := ghostfolio.Project(&src)
			if diff := cmp.Diff(tt.want, got.Balances, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("balances (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantProblems, problems, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("problems (-want +got):\n%s", diff)
			}
		})
	}
}

// TestProjectRejects checks the transactions Project leaves out: a number
// too large for a double and a key too long for a comment.
func TestProjectRejects(t *testing.T) {
	t.Parallel()
	must := mustTx(t)
	etf := instrument(t, isinIE, "ETF1")
	zero := money(t, "0", "USD")
	huge := kernel.NewMoney(dec(t, "1234567890123456789"), cur(t, "USD"))
	long := header(t, "a1", strings.Repeat("x", 2000), "2025-01-02")
	tests := []struct {
		name string
		tx   ledger.Transaction
		want ghostfolio.ProblemReason
	}{
		{"a cost of 19 digits", must(ledger.NewBuy(header(t, "a1", "h1", "2025-01-02"), etf, qty(t, "1"), huge, zero)), ghostfolio.ProblemTooLarge},
		{"a key of 2000 characters", must(ledger.NewBuy(long, etf, qty(t, "1"), money(t, "1", "USD"), zero)), ghostfolio.ProblemCommentTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := ghostfolio.Source{
				Instance: instance, Currency: cur(t, "USD"),
				Accounts: []ghostfolio.SourceAccount{{ID: "a1", Transactions: []ledger.Transaction{tt.tx}}},
				Listings: map[ledger.ISIN]ghostfolio.Listing{etf.ISIN: {DataSource: yahoo, Symbol: "ETF1.L"}},
			}
			got, problems := ghostfolio.Project(&src)
			problems = slices.DeleteFunc(problems, func(p ghostfolio.Problem) bool { return p.Transaction == ledger.TransactionKey{} })
			if len(got.Activities) != 0 || len(problems) != 1 || problems[0].Reason != tt.want {
				t.Errorf("Project = %d activities, problems %v; want none and %s", len(got.Activities), problems, tt.want)
			}
		})
	}
}
