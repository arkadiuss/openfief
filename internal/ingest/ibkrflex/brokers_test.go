package ibkrflex_test

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
)

// TestMapByBrokerAccount maps the synthetic fixtures by IBKR account, as
// the plugin does (D-72), and checks that resolving U0000000 to acc-ibkr
// gives exactly what the mapper configured with that account gives, keys
// included, so an account mapped in the UI books what the plugin's own
// account map booked; and that no entry leaves the mapper in an account or
// with the IBKR account in its key.
func TestMapByBrokerAccount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fixture      string
		wantReported []ingest.BrokerAccountID
		wantUnmapped []ingest.BrokerAccountID
	}{
		{fixture: "activity.xml", wantReported: []ingest.BrokerAccountID{"U0000000", "U0000001"}, wantUnmapped: []ingest.BrokerAccountID{"U0000001"}},
		{fixture: "opening.xml", wantReported: []ingest.BrokerAccountID{"U0000000"}},
	}
	for _, tt := range tests {
		for _, mode := range []ingest.Mode{ingest.ModeOpening, ingest.ModeTransactions} {
			t.Run(tt.fixture+"/"+mode.String(), func(t *testing.T) {
				t.Parallel()
				m := mapByBrokerAccount(t, tt.fixture, mode)
				if diff := cmp.Diff(tt.wantReported, m.BrokerAccounts); diff != "" {
					t.Errorf("BrokerAccounts (-want +got):\n%s", diff)
				}
				known := []ingest.BrokerAccount{{Instance: "ibkr-main", ID: "U0000000", Account: "acc-ibkr"}}
				got := ingest.ResolveBrokerAccounts(&m, known, mode)
				if diff := cmp.Diff(tt.wantUnmapped, got.Unmapped, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unmapped (-want +got):\n%s", diff)
				}
				want := mapFixture(t, tt.fixture, mode)
				checkSameEntries(t, &want, &got.Mapping)
			})
		}
	}
}

// checkSameEntries fails t unless got has want's transactions and balance
// assertions.
func checkSameEntries(t *testing.T, want, got *ingest.Mapping) {
	t.Helper()
	if diff := cmp.Diff(want.Transactions, got.Transactions, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("resolved transactions differ from the configured mapper's (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want.Assertions, got.Assertions, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("resolved assertions differ from the configured mapper's (-want +got):\n%s", diff)
	}
}

// mapByBrokerAccount maps the synthetic fixture testdata/name in mode by
// IBKR account, and fails t unless every entry is in ByBrokerAccount, in
// no account, keyed by BrokerSource.
func mapByBrokerAccount(t *testing.T, name string, mode ingest.Mode) ingest.Mapping {
	t.Helper()
	data, err := fs.ReadFile(os.DirFS("testdata"), name)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m, err := ibkrflex.Mapper{ByBrokerAccount: true}.Map(bytes.NewReader(data), mode)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(m.Transactions) != 0 || len(m.Assertions) != 0 {
		t.Errorf("entries outside ByBrokerAccount: %d transactions, %d assertions", len(m.Transactions), len(m.Assertions))
	}
	for _, e := range m.ByBrokerAccount {
		for i := range e.Transactions {
			if tx := &e.Transactions[i]; tx.Account != "" || tx.Key.Source != ibkrflex.BrokerSource {
				t.Fatalf("%s transaction %v in account %q, want none and source %q", e.Account, tx.Key, tx.Account, ibkrflex.BrokerSource)
			}
		}
	}
	return m
}
