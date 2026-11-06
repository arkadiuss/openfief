package ui

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

func TestStatusSentence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		st   statusLine
		want string
	}{
		{name: "nothing valued, nothing open", want: "Nothing is valued yet. Nothing needs you."},
		{name: "one item", st: statusLine{ValuedThrough: "2026-03-31", Open: 1}, want: "Everything is valued through 2026-03-31. 1 thing needs you."},
		{
			name: "items and failing syncs",
			st:   statusLine{ValuedThrough: "2026-03-31", Open: 3, Failing: []string{"xtb-main", "nbp-fx"}},
			want: "Everything is valued through 2026-03-31. 3 things need you. The xtb-main sync keeps failing. The nbp-fx sync keeps failing.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.st.Sentence(); got != tt.want {
				t.Errorf("Sentence = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMonthChange(t *testing.T) {
	t.Parallel()
	f := numberFormat{Group: narrowSpace, Decimal: ","}
	day := mustDate(t, "2026-03-31")
	total := func(date, gross string) *viewsv1.DateTotal {
		return &viewsv1.DateTotal{Date: date, ValueReporting: pln(gross)}
	}
	tests := []struct {
		name    string
		totals  []*viewsv1.DateTotal
		change  *viewsv1.Money
		percent string
	}{
		{
			name:   "from the last valued day before the month",
			totals: []*viewsv1.DateTotal{total("2026-02-26", "900"), total("2026-02-27", "1000"), total("2026-03-02", "1010"), total("2026-03-31", "1050")},
			change: pln("50"), percent: "+5,0%",
		},
		{name: "a fall", totals: []*viewsv1.DateTotal{total("2026-02-28", "1000"), total("2026-03-31", "968.5")}, change: pln("-31.5"), percent: "-3,2%"},
		{name: "no day before the month", totals: []*viewsv1.DateTotal{total("2026-03-02", "1000"), total("2026-03-31", "1050")}},
		{name: "no total on the day", totals: []*viewsv1.DateTotal{total("2026-02-27", "1000"), total("2026-03-30", "1050")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := monthChange(tt.totals, day, f)
			if diff := cmp.Diff(tt.change, got.Money, protocmp.Transform()); diff != "" || got.Percent != tt.percent {
				t.Errorf("monthChange (-want +got):\n%s\npercent %q, want %q", diff, got.Percent, tt.percent)
			}
		})
	}
}

func TestYearReturn(t *testing.T) {
	t.Parallel()
	f := numberFormat{Group: narrowSpace, Decimal: ","}
	days := []*viewsv1.PerformanceDay{{Date: "2025-12-30", Index: "100"}, {Date: "2025-12-31", Index: "110"}, {Date: "2026-03-31", Index: "121"}}
	if got := yearReturn(days, mustDate(t, "2026-03-31"), f); got != "+10,0%" {
		t.Errorf("yearReturn = %q, want +10,0%%", got)
	}
	if got := yearReturn(days[2:], mustDate(t, "2026-03-31"), f); got != "+0,0%" {
		t.Errorf("yearReturn of one day = %q", got)
	}
	if got := yearReturn(nil, mustDate(t, "2026-03-31"), f); got != "" {
		t.Errorf("yearReturn without days = %q", got)
	}
}

func TestGroupBySleeve(t *testing.T) {
	t.Parallel()
	row := func(id, sleeve, gross string) accountRow {
		r := accountRow{Account: &viewsv1.Account{Id: id, Name: id, Sleeve: sleeve}}
		if gross != "" {
			r.Total = &viewsv1.DateTotal{ValueReporting: pln(gross)}
		}
		return r
	}
	sleeves := []*viewsv1.Sleeve{{Id: "bonds", Name: "Bonds"}, {Id: "core", Name: "Core"}, {Id: "empty", Name: "Empty"}}
	got := groupBySleeve([]accountRow{row("a", "core", "100.50"), row("b", "bonds", "20"), row("c", "core", "9.50"), row("d", "lost", "")}, sleeves)
	type summary struct {
		ID, Name, Total string
		Accounts        []string
	}
	sums := make([]summary, 0, len(got))
	for _, g := range got {
		s := summary{ID: g.ID, Name: g.Name, Total: g.Total.GetAmount()}
		for _, a := range g.Accounts {
			s.Accounts = append(s.Accounts, a.Account.GetId())
		}
		sums = append(sums, s)
	}
	want := []summary{
		{ID: "bonds", Name: "Bonds", Total: "20", Accounts: []string{"b"}},
		{ID: "core", Name: "Core", Total: "110", Accounts: []string{"a", "c"}},
		{ID: "lost", Name: "lost", Accounts: []string{"d"}},
	}
	if diff := cmp.Diff(want, sums); diff != "" {
		t.Errorf("groupBySleeve (-want +got):\n%s", diff)
	}
}

func TestActivityLine(t *testing.T) {
	t.Parallel()
	f := numberFormat{Group: narrowSpace, Decimal: ","}
	names := map[string]string{accountID: "Synthetic brokerage"}
	tests := []struct {
		name string
		a    *viewsv1.Activity
		want activityEntry
	}{
		{
			name: "an import with rows left out",
			a: &viewsv1.Activity{
				Kind: viewsv1.ActivityKind_ACTIVITY_KIND_IMPORT, At: "2026-03-31T10:00:00Z", ConnectorInstance: "files",
				Batch: &viewsv1.ActivityBatch{NewTransactions: 1, Unsupported: 2, AccountIds: []string{accountID}},
			},
			want: activityEntry{Day: "2026-03-31", Text: "Imported · files · 1 new transaction, 2 left out", Links: []inboxLink{{Href: "/accounts/" + accountID, Text: "Synthetic brokerage"}}},
		},
		{
			name: "a sync into two accounts, counted per account",
			a: &viewsv1.Activity{
				Kind: viewsv1.ActivityKind_ACTIVITY_KIND_SYNC, At: "2026-03-31T11:00:00Z", ConnectorInstance: "ibkr-main",
				Batch: &viewsv1.ActivityBatch{NewTransactions: 3, AccountIds: []string{accountID, "acc-2"}, Accounts: []*viewsv1.BatchAccount{
					{AccountId: accountID, NewTransactions: 1}, {AccountId: "acc-2", NewTransactions: 2},
				}},
			},
			want: activityEntry{Day: "2026-03-31", Text: "Synced · ibkr-main · 3 new transactions", Links: []inboxLink{
				{Href: "/accounts/" + accountID, Text: "Synthetic brokerage: 1 new"}, {Href: "/accounts/acc-2", Text: "acc-2: 2 new"},
			}},
		},
		{
			name: "a failed sync",
			a: &viewsv1.Activity{
				Kind: viewsv1.ActivityKind_ACTIVITY_KIND_SYNC_FAILED, At: "2026-03-30T06:00:00Z", ConnectorInstance: "xtb-main",
				Attempt: &viewsv1.SyncAttempt{ErrorClass: viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_TIMEOUT},
			},
			want: activityEntry{Day: "2026-03-30", Text: "Sync failed · xtb-main · timeout"},
		},
		{
			name: "a deletion",
			a: &viewsv1.Activity{
				Kind: viewsv1.ActivityKind_ACTIVITY_KIND_CORRECTION, At: "2026-03-29T06:00:00Z",
				Change: &viewsv1.ActivityChange{AccountId: "acc-gone", Deleted: true, Kind: viewsv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-03-02"},
			},
			want: activityEntry{Day: "2026-03-29", Text: "Deleted · buy of 2026-03-02", Links: []inboxLink{{Href: "/accounts/acc-gone", Text: "acc-gone"}}},
		},
		{
			name: "a coupon",
			a: &viewsv1.Activity{
				Kind: viewsv1.ActivityKind_ACTIVITY_KIND_COUPON, At: "2026-03-28T06:00:00Z",
				Change: &viewsv1.ActivityChange{AccountId: accountID, Instrument: syntheticInstrument(), Amount: pln("12.50")},
			},
			want: activityEntry{Day: "2026-03-28", Text: "Coupon booked · Synthetic Fund · SYN (XX0000000001) · 12,50 PLN", Links: []inboxLink{{Href: "/accounts/" + accountID, Text: "Synthetic brokerage"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, activityLine(tt.a, names, f)); diff != "" {
				t.Errorf("activityLine (-want +got):\n%s", diff)
			}
		})
	}
}
