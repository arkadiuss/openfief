package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestConnectorKeySource(t *testing.T) {
	t.Parallel()
	if got, want := ingest.ConnectorKeySource("xtb-main", "acc-1"), "plugin:xtb-main:acc-1"; got != want {
		t.Errorf("ConnectorKeySource = %q, want %q", got, want)
	}
}

func TestScopeToConnector(t *testing.T) {
	t.Parallel()
	key := func(source, id string) ledger.Transaction {
		return ledger.Transaction{Key: ledger.TransactionKey{Source: source, ExternalID: id}, Account: "acc-1"}
	}
	in := []ledger.Transaction{key("acc-1", "T1"), key("", "T2"), key("plugin:other:acc-1", "T3")}
	got := ingest.ScopeToConnector("xtb-main", in)
	want := []ledger.Transaction{
		key("plugin:xtb-main:acc-1", "T1"),
		key("", "T2"), // still rejected by the ledger: no source
		key("plugin:xtb-main:plugin:other:acc-1", "T3"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ScopeToConnector (-want +got):\n%s", diff)
	}
	if in[0].Key.Source != "acc-1" {
		t.Errorf("ScopeToConnector changed its input: %q", in[0].Key.Source)
	}
	if got := ingest.ScopeToConnector("xtb-main", nil); len(got) != 0 {
		t.Errorf("ScopeToConnector(nil) = %v, want empty", got)
	}
}

func TestBookedBy(t *testing.T) {
	t.Parallel()
	tx := func(source, id, day string) ledger.Transaction {
		d, err := kernel.ParseDate(day)
		if err != nil {
			t.Fatalf("ParseDate: %v", err)
		}
		return ledger.Transaction{Key: ledger.TransactionKey{Source: source, ExternalID: id}, Account: "acc-1", Date: d, Kind: ledger.KindInterest}
	}
	current := []ledger.Transaction{
		tx("plugin:bonds-main:treasury-bonds:acc-1", "lot:COI1028:2024-10-10:buy", "2024-10-10"),
		tx("plugin:bonds-main-2:treasury-bonds:acc-1", "lot:COI1028:2024-10-10:buy", "2024-10-10"), // another instance
		tx("manual", "deposit-1", "2025-01-02"),
		tx("plugin:bonds-main:treasury-bonds:acc-1", "lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
	}
	header := func(id, day string) ledger.TransactionHeader {
		x := tx("treasury-bonds:acc-1", id, day)
		return ledger.TransactionHeader{Key: x.Key, Account: x.Account, Date: x.Date}
	}
	want := []ledger.TransactionHeader{
		header("lot:COI1028:2024-10-10:buy", "2024-10-10"),
		header("lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
	}
	if diff := cmp.Diff(want, ingest.BookedBy("bonds-main", current)); diff != "" {
		t.Errorf("BookedBy (-want +got):\n%s", diff)
	}
	if got := ingest.BookedBy("bonds-main", nil); len(got) != 0 {
		t.Errorf("BookedBy(nil) = %v, want none", got)
	}
}
