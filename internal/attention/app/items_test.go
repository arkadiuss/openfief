package app_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
)

// clock returns a now func that starts at 2026-03-02 10:00 UTC and moves a
// minute per call.
func clock() func() time.Time {
	t := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	return func() time.Time {
		t = t.Add(time.Minute)
		return t
	}
}

// found is a check of kind ticker-unconfirmed that finds what *conds holds.
func found(conds *[]attention.Condition) app.Check {
	return app.Check{Kind: attention.KindTickerUnconfirmed, Find: func(context.Context) ([]attention.Condition, error) {
		return *conds, nil
	}}
}

// states returns each item's subject and state, in ID order.
func states(t *testing.T, items *app.Items) map[string]string {
	t.Helper()
	list, err := items.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	out := map[string]string{}
	for i := range list {
		out[list[i].Subject] = list[i].State().String()
	}
	return out
}

func TestSweepRaisesDeduplicatesAndClears(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	conds := []attention.Condition{
		{Kind: attention.KindTickerUnconfirmed, Subject: "xtb/AAA.US", Detail: "2 rows wait"},
		{Kind: attention.KindTickerUnconfirmed, Subject: "xtb/BBB.PL", Detail: "1 rows wait"},
	}
	failing := app.Check{Kind: attention.KindSettingMissing, Find: func(context.Context) ([]attention.Condition, error) { return nil, errFake }}
	items := app.New(store, fakeOwners{}, clock(), failing, found(&conds))
	for range 2 {
		if err := items.Sweep(t.Context()); !cmp.Equal(err, errFake, cmpopts.EquateErrors()) {
			t.Fatalf("Sweep error = %v, want the failing check's", err)
		}
	}
	if diff := cmp.Diff(map[string]string{"xtb/AAA.US": "open", "xtb/BBB.PL": "open"}, states(t, items)); diff != "" {
		t.Fatalf("after two sweeps (-want +got):\n%s", diff)
	}
	if len(store.events) != 2 {
		t.Errorf("events = %d, want one per raised item", len(store.events))
	}
	conds = conds[:1]
	if err := items.Sweep(t.Context()); !cmp.Equal(err, errFake, cmpopts.EquateErrors()) {
		t.Fatalf("Sweep error = %v, want the failing check's", err)
	}
	if diff := cmp.Diff(map[string]string{"xtb/AAA.US": "open", "xtb/BBB.PL": "resolved"}, states(t, items)); diff != "" {
		t.Errorf("after the condition cleared (-want +got):\n%s", diff)
	}
	if len(store.events) != 3 {
		t.Errorf("events = %d, want one more for the cleared item", len(store.events))
	}
}

func TestRaiseAndClear(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	items := app.New(store, fakeOwners{}, clock())
	cond := attention.Condition{Kind: "instrument-unclassified", Subject: "XS0000000001", Detail: "fund policy unreadable"}
	for range 2 {
		if err := items.Raise(t.Context(), cond); err != nil {
			t.Fatalf("Raise: %v", err)
		}
	}
	if err := items.Clear(t.Context(), cond.Kind, cond.Subject); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if err := items.Clear(t.Context(), cond.Kind, "never-raised"); err != nil {
		t.Fatalf("Clear of nothing: %v", err)
	}
	list, err := items.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || len(list[0].Changes) != 2 || list[0].State() != attention.StateResolved {
		t.Errorf("items = %+v, want one item raised once and cleared", list)
	}
	err = items.Raise(t.Context(), attention.Condition{Kind: "Bad Kind", Subject: "x"})
	if !cmp.Equal(err, attention.ErrInvalidCondition, cmpopts.EquateErrors()) {
		t.Errorf("Raise of an invalid kind error = %v, want ErrInvalidCondition", err)
	}
}

func TestOwnerActions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		kind    attention.Kind
		resolve bool
		id      int64
		want    attention.State
		wantErr error
	}{
		{name: "the owner acknowledges", token: "owner", kind: attention.KindSettingMissing, id: 1, want: attention.StateAcknowledged},
		{name: "the owner resolves import problems", token: "owner", kind: attention.KindImportProblems, resolve: true, id: 1, want: attention.StateResolved},
		{name: "the owner can't resolve an unmapped broker account", token: "owner", kind: attention.KindBrokerAccountUnmapped, resolve: true, id: 1, wantErr: attention.ErrNotOwnerResolvable},
		{name: "a read session can't acknowledge", token: "reader", kind: attention.KindSettingMissing, id: 1, wantErr: app.ErrNotOwner},
		{name: "no session", token: "", kind: attention.KindSettingMissing, id: 1, wantErr: app.ErrNoSession},
		{name: "unknown item", token: "owner", kind: attention.KindSettingMissing, id: 9, wantErr: app.ErrUnknownItem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newMemStore()
			items := app.New(store, fakeOwners{}, clock())
			if err := items.Raise(t.Context(), attention.Condition{Kind: tt.kind, Subject: "s"}); err != nil {
				t.Fatalf("Raise: %v", err)
			}
			action := items.Acknowledge
			if tt.resolve {
				action = items.Resolve
			}
			got, err := action(t.Context(), "https://openfief.example.com", tt.token, tt.id)
			checkOwnerAction(t, store, &got, err, tt.want, tt.wantErr)
		})
	}
}

// checkOwnerAction fails t unless an owner's action left the item in want
// with one more event, or failed with wantErr and stored nothing.
func checkOwnerAction(t *testing.T, store *memStore, got *attention.Item, err error, want attention.State, wantErr error) {
	t.Helper()
	if diff := cmp.Diff(wantErr, err, cmpopts.EquateErrors()); diff != "" {
		t.Fatalf("error (-want +got):\n%s", diff)
	}
	wantEvents := 2
	if wantErr != nil {
		want, wantEvents = 0, 1
	}
	if got.State() != want || len(store.events) != wantEvents {
		t.Errorf("state = %s with %d events, want %s with %d", got.State(), len(store.events), want, wantEvents)
	}
}

func TestFailedAppendStoresNothing(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	store.failAppend = true
	items := app.New(store, fakeOwners{}, clock())
	if err := items.Raise(t.Context(), attention.Condition{Kind: attention.KindSettingMissing, Subject: "s"}); err == nil {
		t.Fatal("Raise succeeded with a failing store")
	}
	if len(store.items) != 0 || len(store.events) != 0 {
		t.Errorf("items %d, events %d: want nothing kept", len(store.items), len(store.events))
	}
}

func TestRelayPublishesInOrderAndResumes(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	items := app.New(store, fakeOwners{}, clock())
	for _, s := range []string{"a", "b", "c"} {
		if err := items.Raise(t.Context(), attention.Condition{Kind: attention.KindSettingMissing, Subject: s}); err != nil {
			t.Fatalf("Raise: %v", err)
		}
	}
	pub := &recorder{failAt: 2}
	relay := &app.Relay{Outbox: store, Publisher: pub, Now: clock()}
	if n, err := relay.Publish(t.Context()); n != 1 || err == nil {
		t.Fatalf("Publish = %d, %v; want 1 and the failure", n, err)
	}
	pub.failAt = 0
	if n, err := relay.Publish(t.Context()); n != 2 || err != nil {
		t.Fatalf("Publish = %d, %v; want the other 2", n, err)
	}
	if diff := cmp.Diff([]int64{1, 2, 3}, pub.got); diff != "" {
		t.Errorf("published (-want +got):\n%s", diff)
	}
	if n, err := relay.Publish(t.Context()); n != 0 || err != nil {
		t.Errorf("Publish with nothing left = %d, %v", n, err)
	}
}

func TestLogPublisherLogsNoSubject(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := app.LogPublisher{Log: slog.New(slog.NewTextHandler(&buf, nil))}
	e := app.Event{ID: 4, ItemID: 2, Kind: attention.KindBrokerAccountUnmapped, Subject: "ibkr/U0000000", Change: attention.Change{
		Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "last seen 2026-03-02",
	}}
	if err := p.Publish(t.Context(), &e); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"event=4", "item=2", "kind=broker-account-unmapped", "state=open", "reason=raised"} {
		if !strings.Contains(out, want) {
			t.Errorf("log %q lacks %q", out, want)
		}
	}
	if strings.Contains(out, "U0000000") || strings.Contains(out, "last seen") {
		t.Errorf("log %q names the subject or detail", out)
	}
}
