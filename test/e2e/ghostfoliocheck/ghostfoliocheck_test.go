package ghostfoliocheck_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/test/e2e/ghostfoliocheck"
)

func TestSyncs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		log     string
		want    []ghostfoliocheck.Sync
		wantErr bool
	}{
		{name: "nothing logged", log: "", want: nil},
		{
			name: "a start, a finished pass and a failed one",
			log: `{"time":"2026-11-08T10:00:00Z","level":"INFO","msg":"ghostfolio syncing","interval":"1h0m0s"}
{"time":"2026-11-08T10:00:05Z","level":"INFO","msg":"synced","accounts_created":5,"imported":40,"deleted":1,"accounts_cleared":2,"balances_set":30,"balances_deleted":3,"left_out":12}
{"time":"2026-11-08T11:00:05Z","level":"ERROR","msg":"sync failed","error":"logging in to Ghostfolio: HTTP 403"}
`,
			want: []ghostfoliocheck.Sync{
				{LeftOut: 12, Report: ghostfoliocheck.Report{
					AccountsCreated: 5, Imported: 40, Deleted: 1, AccountsCleared: 2, BalancesSet: 30, BalancesDeleted: 3,
				}},
				{Err: "logging in to Ghostfolio: HTTP 403"},
			},
		},
		{
			name: "a pass that changed nothing",
			log:  `{"msg":"synced","accounts_created":0,"imported":0,"deleted":0,"accounts_cleared":0,"balances_set":0,"balances_deleted":0,"left_out":0}`,
			want: []ghostfoliocheck.Sync{{}},
		},
		{name: "other lines, JSON or not, are skipped", log: "plain text\n{\"msg\":\"other\"}\n", want: nil},
		{name: "a pass line that isn't JSON", log: `msg="synced" imported=1`, wantErr: true},
		{name: "a finished pass without a count", log: `{"msg":"synced","imported":1}`, wantErr: true},
		{name: "a failed pass without its error", log: `{"msg":"sync failed"}`, wantErr: true},
		{name: "a line naming a pass without a message", log: `{"note":"synced"}`, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ghostfoliocheck.Syncs([]byte(tt.log))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Syncs error = %v, want error %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Syncs (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCommentKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		comment string
		want    string
		ok      bool
	}{
		{comment: "openfief:ibkr/T-1", want: "ibkr/T-1", ok: true},
		{comment: "openfief:ibkr/T-1#fee:USD", want: "ibkr/T-1", ok: true},
		{comment: "openfief:demo/a#fee:b", want: "demo/a#fee:b", ok: true},
		{comment: "openfief:", ok: false},
		{comment: "openfief:#fee:EUR", ok: false},
		{comment: "a fee of my own", ok: false},
		{comment: "", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			t.Parallel()
			got, ok := ghostfoliocheck.CommentKey(tt.comment)
			if got != tt.want || ok != tt.ok {
				t.Errorf("CommentKey(%q) = %q, %v; want %q, %v", tt.comment, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNetUnits(t *testing.T) {
	t.Parallel()
	acts := []ghostfoliocheck.Activity{
		act(t, "1", "acc", ghostfoliocheck.TypeBuy, "BOND", "10"),
		act(t, "2", "acc", ghostfoliocheck.TypeBuy, "BOND", "2.5"),
		act(t, "3", "acc", ghostfoliocheck.TypeSell, "BOND", "4"),
		act(t, "4", "acc", ghostfoliocheck.TypeInterest, "BOND", "1"),
		act(t, "5", "other", ghostfoliocheck.TypeBuy, "BOND", "100"),
		act(t, "6", "acc", ghostfoliocheck.TypeBuy, "OTHER", "7"),
	}
	tests := []struct {
		name, account, symbol, want string
	}{
		{name: "buys less sells, income left out", account: "acc", symbol: "BOND", want: "8.5"},
		{name: "another account", account: "other", symbol: "BOND", want: "100"},
		{name: "nothing held", account: "acc", symbol: "NONE", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ghostfoliocheck.NetUnits(acts, tt.account, tt.symbol); !got.Equal(dec(t, tt.want)) {
				t.Errorf("NetUnits(%s, %s) = %s, want %s", tt.account, tt.symbol, got, tt.want)
			}
		})
	}
}

func TestLatest(t *testing.T) {
	t.Parallel()
	bs := []ghostfoliocheck.Balance{
		{AccountID: "acc", Date: date(t, "2026-01-05"), Value: dec(t, "30")},
		{AccountID: "acc", Date: date(t, "2026-01-02"), Value: dec(t, "10")},
		{AccountID: "acc", Date: date(t, "2026-01-10"), Value: dec(t, "-5")},
		{AccountID: "other", Date: date(t, "2026-01-03"), Value: dec(t, "99")},
	}
	tests := []struct {
		name, account, on, want string
	}{
		{name: "the latest on or before the date", account: "acc", on: "2026-01-07", want: "30"},
		{name: "on the balance's own date", account: "acc", on: "2026-01-05", want: "30"},
		{name: "after every balance", account: "acc", on: "2026-02-01", want: "-5"},
		{name: "before every balance", account: "acc", on: "2026-01-01", want: "0"},
		{name: "another account", account: "other", on: "2026-01-07", want: "99"},
		{name: "an account with none", account: "none", on: "2026-01-07", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ghostfoliocheck.Latest(bs, tt.account, date(t, tt.on)); !got.Equal(dec(t, tt.want)) {
				t.Errorf("Latest(%s, %s) = %s, want %s", tt.account, tt.on, got, tt.want)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	t.Parallel()
	b := act(t, "b", "acc", ghostfoliocheck.TypeBuy, "BOND", "10")
	a := act(t, "a", "acc", ghostfoliocheck.TypeFee, "FEES", "0")
	a.Fee = dec(t, "1.5")
	got := ghostfoliocheck.Describe([]ghostfoliocheck.Activity{b, a})
	want := []string{
		"a acc openfief:x/a FEE 2026-01-02 PLN MANUAL FEES 0 100 1.5",
		"b acc openfief:x/b BUY 2026-01-02 PLN MANUAL BOND 10 100 0",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Describe (-want +got):\n%s", diff)
	}
	balances := ghostfoliocheck.DescribeBalances([]ghostfoliocheck.Balance{
		{ID: "2", AccountID: "acc", Date: date(t, "2026-01-05"), Value: dec(t, "30.5")},
		{ID: "1", AccountID: "acc", Date: date(t, "2026-01-02"), Value: dec(t, "10")},
	})
	if diff := cmp.Diff([]string{"1 acc 2026-01-02 10", "2 acc 2026-01-05 30.5"}, balances); diff != "" {
		t.Errorf("DescribeBalances (-want +got):\n%s", diff)
	}
}

func TestEditable(t *testing.T) {
	t.Parallel()
	fee := act(t, "a", "acc", ghostfoliocheck.TypeFee, "FEES", "0")
	buy := act(t, "c", "acc", ghostfoliocheck.TypeBuy, "BOND", "10")
	interest := act(t, "b", "acc", ghostfoliocheck.TypeInterest, "BOND", "1")
	tests := []struct {
		name   string
		acts   []ghostfoliocheck.Activity
		wantID string
		ok     bool
	}{
		{name: "the first by ID holding units", acts: []ghostfoliocheck.Activity{buy, fee, interest}, wantID: "b", ok: true},
		{name: "fees only", acts: []ghostfoliocheck.Activity{fee}, ok: false},
		{name: "none", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ghostfoliocheck.Editable(tt.acts)
			if got.ID != tt.wantID || ok != tt.ok {
				t.Errorf("Editable = %q, %v; want %q, %v", got.ID, ok, tt.wantID, tt.ok)
			}
		})
	}
}

func TestEdited(t *testing.T) {
	t.Parallel()
	a := act(t, "id-1", "acc", ghostfoliocheck.TypeBuy, "BOND", "10")
	a.Fee = dec(t, "0.7")
	data, err := json.Marshal(ghostfoliocheck.Edited(&a))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"accountId":"acc","comment":"openfief:x/id-1","currency":"PLN","dataSource":"MANUAL",` +
		`"date":"2026-01-02T00:00:00.000Z","fee":0.7,"id":"id-1","quantity":11,"symbol":"BOND","type":"BUY","unitPrice":100}`
	if string(data) != want {
		t.Errorf("Edited = %s\nwant      %s", data, want)
	}
}

// act is an activity of account on a fixed day, in PLN, at 100 a unit.
func act(t *testing.T, id, account, typ, symbol, quantity string) ghostfoliocheck.Activity {
	t.Helper()
	return ghostfoliocheck.Activity{
		ID: id, AccountID: account, Comment: "openfief:x/" + id, Type: typ, Date: date(t, "2026-01-02"),
		Currency: cur(t, "PLN"), DataSource: "MANUAL", Symbol: symbol,
		Quantity: dec(t, quantity), UnitPrice: dec(t, "100"),
	}
}

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func cur(t *testing.T, s string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
