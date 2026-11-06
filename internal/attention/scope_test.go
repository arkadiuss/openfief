package attention_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/ingest"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
)

func TestGroupOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind attention.Kind
		want attention.Group
	}{
		{attention.KindTickerUnconfirmed, attention.GroupDataComingIn},
		{attention.KindBrokerAccountUnmapped, attention.GroupDataComingIn},
		{attention.KindSyncFailing, attention.GroupDataComingIn},
		{attention.KindImportProblems, attention.GroupDataComingIn},
		{ingest.KindUndeclaredDeliveries, attention.GroupDataComingIn},
		{valuationapp.GapKind, attention.GroupDataComingIn},
		{attention.KindAssertionFlagged, attention.GroupNumbersToCheck},
		{attention.KindPriceCheckFlagged, attention.GroupNumbersToCheck},
		{attention.KindTransferBroken, attention.GroupDecisions},
		{attention.KindSettingMissing, attention.GroupDecisions},
		{valuationapp.UnclassifiedKind, attention.GroupDecisions},
		{"some-later-check", attention.GroupDecisions},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			t.Parallel()
			if got := attention.GroupOf(tt.kind); got != tt.want {
				t.Errorf("GroupOf = %v, want %v", got, tt.want)
			}
		})
	}
	names := map[attention.Group]string{
		attention.GroupDataComingIn: "data-coming-in", attention.GroupNumbersToCheck: "numbers-to-check",
		attention.GroupDecisions: "decisions", 0: "Group(0)",
	}
	for g, name := range names {
		if got := g.String(); got != name {
			t.Errorf("%d.String() = %q, want %q", int(g), got, name)
		}
	}
}

func TestScopeOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    attention.Kind
		subject string
		detail  string
		want    attention.Scope
	}{
		{
			name: "a holding's balance assertion", kind: attention.KindAssertionFlagged, subject: "acc-1/2026-03-02/XX0000000001/SYNT",
			want: attention.Scope{Account: "acc-1", Instrument: "XX0000000001"},
		},
		{
			name: "a cash balance assertion", kind: attention.KindAssertionFlagged, subject: "acc-1/2026-03-02/cash:PLN",
			want: attention.Scope{Account: "acc-1"},
		},
		{
			name: "a price check", kind: attention.KindPriceCheckFlagged, subject: "XX0000000001/SYNT",
			want: attention.Scope{Instrument: "XX0000000001"},
		},
		{
			name: "an instrument to classify", kind: valuationapp.UnclassifiedKind, subject: "XX0000000001",
			want: attention.Scope{Instrument: "XX0000000001"},
		},
		{
			name: "a price series gap", kind: valuationapp.GapKind, subject: "acc-2/prices/XX0000000001/SYNT",
			want: attention.Scope{Account: "acc-2", Instrument: "XX0000000001"},
		},
		{
			name: "an FX series gap", kind: valuationapp.GapKind, subject: "acc-2/fx_rates/USD",
			want: attention.Scope{Account: "acc-2"},
		},
		{
			name: "an unmapped broker account names its instance only", kind: attention.KindBrokerAccountUnmapped, subject: "ibkr-main/U0000001",
			want: attention.Scope{Instance: "ibkr-main"},
		},
		{
			name: "a ticker", kind: attention.KindTickerUnconfirmed, subject: "xtb-main/CDR.PL",
			want: attention.Scope{Instance: "xtb-main"},
		},
		{
			name: "a failing sync", kind: attention.KindSyncFailing, subject: "xtb-main",
			want: attention.Scope{Instance: "xtb-main"},
		},
		{
			name: "undeclared deliveries", kind: ingest.KindUndeclaredDeliveries, subject: "bonds-main",
			want: attention.Scope{Instance: "bonds-main"},
		},
		{
			name: "import problems name the instance in the detail", kind: attention.KindImportProblems, subject: "b-1",
			detail: "xtb-main: 2 rows left out", want: attention.Scope{Instance: "xtb-main"},
		},
		{name: "import problems without an instance", kind: attention.KindImportProblems, subject: "b-1", detail: "2 rows"},
		{name: "a setting", kind: attention.KindSettingMissing, subject: "marginal-rate"},
		{name: "a broken transfer", kind: attention.KindTransferBroken, subject: "test/w -> test/d"},
		{name: "a later check", kind: "some-later-check", subject: "acc-1/x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			it := attention.Item{Kind: tt.kind, Subject: tt.subject, Changes: []attention.Change{{Seq: 1, State: attention.StateOpen, Detail: tt.detail}}}
			if diff := cmp.Diff(tt.want, attention.ScopeOf(&it)); diff != "" {
				t.Errorf("ScopeOf (-want +got):\n%s", diff)
			}
		})
	}
}
