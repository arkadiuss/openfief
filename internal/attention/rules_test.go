package attention_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
)

// t0 and t1 return the times of the changes the tests make.
func t0() time.Time { return time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC) }

func t1() time.Time { return time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC) }

// itemIn returns an item of kind whose history ends in state and reason,
// with detail "d".
func itemIn(kind attention.Kind, state attention.State, reason attention.Reason) *attention.Item {
	return &attention.Item{ID: 7, Kind: kind, Subject: "s", Changes: []attention.Change{
		{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "first", At: t0()},
		{Seq: 2, State: state, Reason: reason, Detail: "d", At: t0()},
	}}
}

func TestObserve(t *testing.T) {
	t.Parallel()
	const kind = attention.KindPriceCheckFlagged
	tests := []struct {
		name    string
		item    *attention.Item
		present bool
		detail  string
		want    attention.Change
		wantOK  bool
	}{
		{
			name: "new condition raises an item", present: true, detail: "d",
			want: attention.Change{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "d", At: t1()}, wantOK: true,
		},
		{name: "nothing found and no item changes nothing"},
		{name: "open item with the same detail stays", item: itemIn(kind, attention.StateOpen, attention.ReasonChanged), present: true, detail: "d"},
		{
			name: "open item with another detail changes", item: itemIn(kind, attention.StateOpen, attention.ReasonChanged), present: true, detail: "e",
			want: attention.Change{Seq: 3, State: attention.StateOpen, Reason: attention.ReasonChanged, Detail: "e", At: t1()}, wantOK: true,
		},
		{name: "acknowledged item with the same detail stays", item: itemIn(kind, attention.StateAcknowledged, attention.ReasonAcknowledged), present: true, detail: "d"},
		{
			name: "acknowledged item with another detail opens again", item: itemIn(kind, attention.StateAcknowledged, attention.ReasonAcknowledged), present: true, detail: "e",
			want: attention.Change{Seq: 3, State: attention.StateOpen, Reason: attention.ReasonChanged, Detail: "e", At: t1()}, wantOK: true,
		},
		{
			name: "open item clears", item: itemIn(kind, attention.StateOpen, attention.ReasonChanged),
			want: attention.Change{Seq: 3, State: attention.StateResolved, Reason: attention.ReasonCleared, Detail: "d", At: t1()}, wantOK: true,
		},
		{
			name: "acknowledged item clears", item: itemIn(kind, attention.StateAcknowledged, attention.ReasonAcknowledged),
			want: attention.Change{Seq: 3, State: attention.StateResolved, Reason: attention.ReasonCleared, Detail: "d", At: t1()}, wantOK: true,
		},
		{
			name: "cleared item comes back", item: itemIn(kind, attention.StateResolved, attention.ReasonCleared), present: true, detail: "d",
			want: attention.Change{Seq: 3, State: attention.StateOpen, Reason: attention.ReasonReopened, Detail: "d", At: t1()}, wantOK: true,
		},
		{name: "resolved item stays resolved while not found", item: itemIn(kind, attention.StateResolved, attention.ReasonCleared)},
		{name: "owner-resolved item with the same detail stays resolved", item: itemIn(kind, attention.StateResolved, attention.ReasonResolved), present: true, detail: "d"},
		{
			name: "owner-resolved item with another detail reopens", item: itemIn(kind, attention.StateResolved, attention.ReasonResolved), present: true, detail: "e",
			want: attention.Change{Seq: 3, State: attention.StateOpen, Reason: attention.ReasonReopened, Detail: "e", At: t1()}, wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cond := attention.Condition{Kind: kind, Subject: "s", Detail: tt.detail}
			got, ok := attention.Observe(tt.item, &cond, tt.present, t1())
			if ok != tt.wantOK {
				t.Fatalf("Observe ok = %v, want %v", ok, tt.wantOK)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Observe (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReconcile(t *testing.T) {
	t.Parallel()
	const kind = attention.KindBrokerAccountUnmapped
	stored := []attention.Item{
		{ID: 1, Kind: kind, Subject: "ibkr/a", Changes: []attention.Change{{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, At: t0()}}},
		{ID: 2, Kind: kind, Subject: "ibkr/b", Changes: []attention.Change{{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, At: t0()}}},
	}
	tests := []struct {
		name    string
		conds   []attention.Condition
		want    []attention.Step
		wantErr error
	}{
		{
			name:  "one stays, one clears, one is raised, in subject order",
			conds: []attention.Condition{{Kind: kind, Subject: "ibkr/c"}, {Kind: kind, Subject: "ibkr/a"}},
			want: []attention.Step{
				{Kind: kind, Subject: "ibkr/b", ItemID: 2, Change: attention.Change{Seq: 2, State: attention.StateResolved, Reason: attention.ReasonCleared, At: t1()}},
				{Kind: kind, Subject: "ibkr/c", Change: attention.Change{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, At: t1()}},
			},
		},
		{name: "same conditions change nothing", conds: []attention.Condition{{Kind: kind, Subject: "ibkr/a"}, {Kind: kind, Subject: "ibkr/b"}}},
		{name: "a repeated subject is refused", conds: []attention.Condition{{Kind: kind, Subject: "x"}, {Kind: kind, Subject: "x"}}, wantErr: attention.ErrInvalidCondition},
		{name: "another kind is refused", conds: []attention.Condition{{Kind: attention.KindTickerUnconfirmed, Subject: "x"}}, wantErr: attention.ErrInvalidCondition},
		{name: "an empty subject is refused", conds: []attention.Condition{{Kind: kind}}, wantErr: attention.ErrInvalidCondition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := attention.Reconcile(kind, stored, tt.conds, t1())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Reconcile error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Reconcile (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOwnerActions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		resolve bool
		item    *attention.Item
		want    attention.Change
		wantOK  bool
		wantErr error
	}{
		{
			name: "acknowledging an open item", item: itemIn(attention.KindSettingMissing, attention.StateOpen, attention.ReasonRaised),
			want: attention.Change{Seq: 3, State: attention.StateAcknowledged, Reason: attention.ReasonAcknowledged, Detail: "d", At: t1()}, wantOK: true,
		},
		{name: "acknowledging again changes nothing", item: itemIn(attention.KindSettingMissing, attention.StateAcknowledged, attention.ReasonAcknowledged)},
		{name: "acknowledging a resolved item is refused", item: itemIn(attention.KindSettingMissing, attention.StateResolved, attention.ReasonCleared), wantErr: attention.ErrInvalidTransition},
		{
			name: "resolving an acknowledged flagged assertion", resolve: true, item: itemIn(attention.KindAssertionFlagged, attention.StateAcknowledged, attention.ReasonAcknowledged),
			want: attention.Change{Seq: 3, State: attention.StateResolved, Reason: attention.ReasonResolved, Detail: "d", At: t1()}, wantOK: true,
		},
		{name: "resolving import problems again changes nothing", resolve: true, item: itemIn(attention.KindImportProblems, attention.StateResolved, attention.ReasonResolved)},
		{name: "resolving an unmapped broker account by hand is refused", resolve: true, item: itemIn(attention.KindBrokerAccountUnmapped, attention.StateOpen, attention.ReasonRaised), wantErr: attention.ErrNotOwnerResolvable},
		{name: "resolving an item with no history is refused", resolve: true, item: &attention.Item{Kind: attention.KindImportProblems}, wantErr: attention.ErrInvalidTransition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			action := attention.Acknowledge
			if tt.resolve {
				action = attention.Resolve
			}
			got, ok, err := action(tt.item, t1())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("change (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConditionValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cond attention.Condition
		want error
	}{
		{name: "valid", cond: attention.Condition{Kind: "instrument-unclassified", Subject: "x", Detail: "y"}},
		{name: "kind with capitals", cond: attention.Condition{Kind: "Price", Subject: "x"}, want: attention.ErrInvalidCondition},
		{name: "kind with a trailing hyphen", cond: attention.Condition{Kind: "price-", Subject: "x"}, want: attention.ErrInvalidCondition},
		{name: "subject too long", cond: attention.Condition{Kind: "k", Subject: strings.Repeat("x", attention.MaxSubjectLen+1)}, want: attention.ErrInvalidCondition},
		{name: "detail too long", cond: attention.Condition{Kind: "k", Subject: "x", Detail: strings.Repeat("x", attention.MaxDetailLen+1)}, want: attention.ErrInvalidCondition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.cond.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "open", got: attention.StateOpen.String(), want: "open"},
		{name: "acknowledged", got: attention.StateAcknowledged.String(), want: "acknowledged"},
		{name: "resolved", got: attention.StateResolved.String(), want: "resolved"},
		{name: "invalid state", got: attention.State(0).String(), want: "State(0)"},
		{name: "raised", got: attention.ReasonRaised.String(), want: "raised"},
		{name: "changed", got: attention.ReasonChanged.String(), want: "changed"},
		{name: "reopened", got: attention.ReasonReopened.String(), want: "reopened"},
		{name: "cleared", got: attention.ReasonCleared.String(), want: "cleared"},
		{name: "acknowledged reason", got: attention.ReasonAcknowledged.String(), want: "acknowledged"},
		{name: "resolved reason", got: attention.ReasonResolved.String(), want: "resolved"},
		{name: "invalid reason", got: attention.Reason(0).String(), want: "Reason(0)"},
		{name: "acknowledging is the owner's", got: strconv.FormatBool(attention.ReasonAcknowledged.ByOwner()), want: "true"},
		{name: "resolving is the owner's", got: strconv.FormatBool(attention.ReasonResolved.ByOwner()), want: "true"},
		{name: "clearing is a check's", got: strconv.FormatBool(attention.ReasonCleared.ByOwner()), want: "false"},
		{name: "item without history has no state", got: (&attention.Item{}).State().String(), want: "State(0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
