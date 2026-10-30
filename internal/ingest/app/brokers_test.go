package app_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// fakeBrokers is an in-memory app.BrokerAccountStore.
type fakeBrokers struct {
	accounts []ingest.BrokerAccount
	err      error
}

func (f *fakeBrokers) RecordBrokerAccounts(_ context.Context, instance string, institution ingest.InstitutionID, ids []ingest.BrokerAccountID, on kernel.Date) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.accounts {
		if b := &f.accounts[i]; b.Instance == instance {
			b.InLastDelivery = slices.Contains(ids, b.ID)
			if b.InLastDelivery {
				b.LastSeen = on
			}
		}
	}
	for _, id := range ids {
		if !slices.ContainsFunc(f.accounts, func(b ingest.BrokerAccount) bool { return b.Instance == instance && b.ID == id }) {
			f.accounts = append(f.accounts, ingest.BrokerAccount{Instance: instance, ID: id, Institution: institution, LastSeen: on, InLastDelivery: true})
		}
	}
	return nil
}

func (f *fakeBrokers) BrokerAccounts(_ context.Context, instance string) ([]ingest.BrokerAccount, error) {
	var out []ingest.BrokerAccount
	for _, b := range f.accounts {
		if b.Instance == instance {
			out = append(out, b)
		}
	}
	slices.SortFunc(out, func(a, b ingest.BrokerAccount) int { return strings.Compare(string(a.ID), string(b.ID)) })
	return out, f.err
}

func (f *fakeBrokers) MarkSeeded(_ context.Context, instance string, ids []ingest.BrokerAccountID) error {
	for i := range f.accounts {
		if b := &f.accounts[i]; b.Instance == instance && slices.Contains(ids, b.ID) {
			b.Seeded = true
		}
	}
	return f.err
}

func (f *fakeBrokers) MapBrokerAccount(_ context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, _ time.Time) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.accounts {
		b := &f.accounts[i]
		if b.Instance != instance || b.ID != id {
			continue
		}
		if b.Mapped() && b.Account != account {
			return app.ErrBrokerAccountMapped
		}
		b.Account = account
		return nil
	}
	return app.ErrUnknownBrokerAccount
}

// byBroker returns m with its transactions and assertions named by broker
// account id instead of their account, keyed by the source "test".
func byBroker(m *ingest.Mapping, id ingest.BrokerAccountID, others ...ingest.BrokerAccountID) ingest.Mapping {
	e := ingest.BrokerEntries{Account: id}
	for i := range m.Transactions {
		t := m.Transactions[i]
		t.Account = ""
		e.Transactions = append(e.Transactions, t)
	}
	for _, a := range m.Assertions {
		a.Account = ""
		e.Assertions = append(e.Assertions, a)
	}
	return ingest.Mapping{BrokerAccounts: append([]ingest.BrokerAccountID{id}, others...), ByBrokerAccount: []ingest.BrokerEntries{e}}
}

func TestConnectorsResolveBrokerAccounts(t *testing.T) {
	t.Parallel()
	const main, side ingest.BrokerAccountID = "U0000001", "U0000002"
	mappedMain := func(seeded bool) []ingest.BrokerAccount {
		return []ingest.BrokerAccount{{Instance: "ibkr-main", ID: main, Institution: "ibkr", Account: "acc-1", Seeded: seeded}}
	}
	tests := []struct {
		name       string
		known      []ingest.BrokerAccount
		mode       ingest.Mode
		mapping    func(t *testing.T) ingest.Mapping
		wantErr    error
		wantKeys   []string
		wantSeeded bool
	}{
		{
			name: "unmapped broker accounts book nothing, but are recorded", mode: ingest.ModeOpening, mapping: openingMapping,
		},
		{
			name: "a mapped account is booked in its account with the account in the key, and seeded", known: mappedMain(false),
			mode: ingest.ModeOpening, mapping: openingMapping,
			wantKeys: []string{"plugin:ibkr-main:test:acc-1"}, wantSeeded: true,
		},
		{
			name: "an opening batch doesn't seed an account seeded already", known: mappedMain(true),
			mode: ingest.ModeOpening, mapping: openingMapping, wantSeeded: true,
		},
		{
			name: "a transactions batch books a mapped account", known: mappedMain(false),
			mode: ingest.ModeTransactions, mapping: mapping, wantKeys: []string{"plugin:ibkr-main:test:acc-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			m := tt.mapping(t)
			d.Institution, d.Mapping = "ibkr", byBroker(&m, main, side)
			l := &fakeLedger{}
			brokers := &fakeBrokers{accounts: slices.Clone(tt.known)}
			c := connectors(t, &fakeCaller{delivery: d}, l, &fakeStores{})
			c.Brokers = brokers
			_, err := c.Sync(t.Context(), "ibkr-main", &app.SyncCall{ID: "b-1", Mode: tt.mode})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Sync error (-want +got):\n%s", diff)
			}
			checkBooked(t, l.txs, tt.wantKeys)
			checkRecorded(t, brokers, side, tt.wantSeeded)
		})
	}
}

// checkBooked fails t unless txs are in acc-1 with the key sources want.
func checkBooked(t *testing.T, txs []ledger.Transaction, want []string) {
	t.Helper()
	keys := make([]string, 0, len(txs))
	for i := range txs {
		keys = append(keys, txs[i].Key.Source)
		if txs[i].Account != "acc-1" {
			t.Errorf("transaction %v booked in %q, want acc-1", txs[i].Key, txs[i].Account)
		}
	}
	if diff := cmp.Diff(want, keys, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("booked key sources (-want +got):\n%s", diff)
	}
}

// checkRecorded fails t unless brokers holds ibkr-main's first broker
// account seeded as wantSeeded and side recorded, unmapped, in the last
// delivery.
func checkRecorded(t *testing.T, brokers *fakeBrokers, side ingest.BrokerAccountID, wantSeeded bool) {
	t.Helper()
	got, err := brokers.BrokerAccounts(t.Context(), "ibkr-main")
	if err != nil {
		t.Fatalf("BrokerAccounts: %v", err)
	}
	if len(got) != 2 || got[1].ID != side || got[1].Mapped() || !got[1].InLastDelivery || got[1].Institution != "ibkr" {
		t.Errorf("broker accounts = %+v, want %s recorded, unmapped, in the last delivery", got, side)
	}
	if len(got) > 0 && got[0].Seeded != wantSeeded {
		t.Errorf("%s seeded = %v, want %v", got[0].ID, got[0].Seeded, wantSeeded)
	}
}

func TestConnectorsBrokerAccountsNeedAStore(t *testing.T) {
	t.Parallel()
	d := delivery(t)
	m := mapping(t)
	d.Mapping = byBroker(&m, "U0000001")
	l := &fakeLedger{}
	_, err := connectors(t, &fakeCaller{delivery: d}, l, &fakeStores{}).Sync(t.Context(), "ibkr-main", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
	if diff := cmp.Diff(app.ErrNoBrokerAccountStore, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Sync error (-want +got):\n%s", diff)
	}
	if l.calls != 0 {
		t.Errorf("%d ledger imports, want none", l.calls)
	}
}

// TestConnectorsImportFileByBrokerAccount checks that an uploaded statement
// of linked accounts books every mapped broker account bound to the
// instance, lists the unmapped one and books nothing of it (D-109), and
// that a broker account mapped to an account not bound to the instance
// refuses the statement whole.
func TestConnectorsImportFileByBrokerAccount(t *testing.T) {
	t.Parallel()
	d := delivery(t)
	m := mapping(t)
	one, two, three := byBroker(&m, "U0000001"), byBroker(&m, "U0000002"), byBroker(&m, "U0000003")
	d.Mapping = ingest.Mapping{
		BrokerAccounts:  []ingest.BrokerAccountID{"U0000001", "U0000002", "U0000003"},
		ByBrokerAccount: append(append(one.ByBrokerAccount, two.ByBrokerAccount...), three.ByBrokerAccount...),
	}
	l := &fakeLedger{}
	c := connectors(t, &fakeCaller{delivery: d}, l, &fakeStores{})
	c.Bindings = &fakeBindings{accounts: []ledger.AccountID{"acc-1", "acc-2"}}
	brokers := &fakeBrokers{accounts: []ingest.BrokerAccount{
		{Instance: "ibkr-main", ID: "U0000001", Account: "acc-1"},
		{Instance: "ibkr-main", ID: "U0000002", Account: "acc-2"},
		{Instance: "ibkr-main", ID: "U0000003"},
	}}
	c.Brokers = brokers
	b, err := c.ImportFile(t.Context(), "ibkr-main", "b-1", ingest.ModeTransactions, app.NamedFile{Name: "s.xml", Contents: strings.NewReader("x")})
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	got := []any{b.Transactions, b.Accounts, b.Unmapped, len(l.txs)}
	want := []any{2, []ledger.AccountID{"acc-1", "acc-2"}, []ingest.BrokerAccountID{"U0000003"}, 2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("transactions, accounts, unmapped, booked (-want +got):\n%s", diff)
	}
	brokers.accounts[1].Account = "acc-9"
	_, err = c.ImportFile(t.Context(), "ibkr-main", "b-2", ingest.ModeTransactions, app.NamedFile{Name: "s.xml", Contents: strings.NewReader("x")})
	if diff := cmp.Diff(ingest.ErrAccountNotBound, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("ImportFile into an account not bound to the instance error (-want +got):\n%s", diff)
	}
}
