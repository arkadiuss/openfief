package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// minute is minute m of a synthetic morning.
func minute(m int) time.Time {
	return time.Date(2026, time.November, 4, 6, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute)
}

// fakeHistory is the ingest store's sync history: an upload, a sync and
// three failed attempts of xtb-main.
type fakeHistory struct{}

func (fakeHistory) CompletedBatches(_ context.Context, q *ingest.BatchQuery) ([]ingest.BatchRecord, error) {
	all := []ingest.BatchRecord{
		{
			ID: "u-1", Instance: "files", Source: "upload", Mode: ingest.ModeTransactions, Finished: minute(25), Institution: "xtb", Transactions: 42, New: 38, Problems: 4, Unsupported: 4, Accounts: []ledger.AccountID{"acc-1", "acc-2"},
			PerAccount: []ingest.AccountCounts{{Account: "acc-1", Transactions: 40, New: 36, NewVersions: 1}, {Account: "acc-2", Transactions: 2, New: 2, FailedAssertions: 1}},
		},
		{ID: "b-1", Instance: "xtb-main", Source: "sync", Mode: ingest.ModeTransactions, Finished: minute(10), Institution: "xtb", Transactions: 2, New: 2},
	}
	var out []ingest.BatchRecord
	for i := range all {
		if b := &all[i]; (q.Before.IsZero() || b.Finished.Before(q.Before)) && len(out) < q.Limit {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (fakeHistory) SyncAttempts(_ context.Context, q *ingest.AttemptQuery) ([]ingest.SyncAttempt, error) {
	fail := func(id int64, m, n int) ingest.SyncAttempt {
		return ingest.SyncAttempt{
			ID: id, Instance: "xtb-main", Batch: "b-2", Mode: ingest.ModeTransactions, Started: minute(m - 1), Ended: minute(m),
			Class: ingest.SyncErrorTimeout, Failures: n, NextSync: minute(m + 5*n),
		}
	}
	all := []ingest.SyncAttempt{
		fail(4, 40, 3), fail(3, 30, 2), fail(2, 20, 1),
		{ID: 1, Instance: "xtb-main", Batch: "b-1", Mode: ingest.ModeTransactions, Started: minute(9), Ended: minute(10), NextSync: minute(370)},
	}
	var out []ingest.SyncAttempt
	for i := range all {
		a := &all[i]
		keep := (q.Instance == "" || a.Instance == q.Instance) && (!q.FailedOnly || a.Failed()) &&
			(q.Before.IsZero() || a.Ended.Before(q.Before))
		if keep && len(out) < q.Limit {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (fakeHistory) SyncStatuses(context.Context) ([]ingest.SyncStatus, error) {
	return []ingest.SyncStatus{{
		Instance: "xtb-main", LastAttempt: minute(40), LastSuccess: minute(10), Failures: 3, FailingSince: minute(19),
		LastClass: ingest.SyncErrorTimeout, NextSync: minute(55),
	}}, nil
}

// fakeChanges is one coupon on acc-1, booked at minute 15.
type fakeChanges struct {
	coupon ledger.Transaction
}

func (f *fakeChanges) RecordedChanges(_ context.Context, before time.Time, _ int64, _ int) ([]ledgerapp.RecordedVersion, error) {
	if !before.IsZero() && !minute(15).Before(before) {
		return nil, nil
	}
	return []ledgerapp.RecordedVersion{{VersionedTransaction: ledger.VersionedTransaction{Transaction: f.coupon, Version: 1}, ID: 7, Recorded: minute(15)}}, nil
}

// fakePrices prices ETF1 from two sources, test-prices being its own.
type fakePrices struct {
	prices []valuation.Price
}

func (f fakePrices) PriceHistory(ctx context.Context, isin string, _, _ kernel.Date) ([]valuation.Price, error) {
	return f.LatestPrices(ctx, isin)
}

func (f fakePrices) LatestPrices(_ context.Context, isin string) ([]valuation.Price, error) {
	var out []valuation.Price
	for _, p := range f.prices {
		if in, ok := views.ListingOf(p.Instrument); ok && in.ISIN.String() == isin {
			out = append(out, p)
		}
	}
	return out, nil
}

func (fakePrices) SourceOf(_ context.Context, s valuation.Series) (valuation.SeriesSource, bool, error) {
	return valuation.SeriesSource{Series: s, Source: "test-prices"}, true, nil
}

// uiv2API serves the read API over newLedger's acc-1, with a coupon, the
// sync history, two prices of ETF1, two attention items and EDO0334's
// stored terms.
func uiv2API(t *testing.T) api {
	t.Helper()
	l := newLedger(t)
	coupon, err := ledger.NewInterest(ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "coupon-1"}, Account: "acc-1", Date: date(t, "2026-03-02")},
		instrument(t, isinPL, "EDO0334"), money(t, "12.50", "PLN"), money(t, "2.38", "PLN"))
	if err != nil {
		t.Fatalf("NewInterest: %v", err)
	}
	etf := instrument(t, isinIE, "ETF1").String()
	prices := fakePrices{prices: []valuation.Price{
		{Instrument: etf, Date: date(t, "2026-03-02"), Close: money(t, "10.50", "USD"), Source: "ibkr"},
		{Instrument: etf, Date: date(t, "2026-03-02"), Close: money(t, "10.49", "USD"), Source: "test-prices"},
	}}
	opened := attention.Change{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "failing since 2026-11-04T06:19Z", At: minute(40)}
	items := fakeItems{
		{ID: 1, Kind: attention.KindSyncFailing, Subject: "xtb-main", Changes: []attention.Change{opened}},
		{ID: 2, Kind: attention.KindAssertionFlagged, Subject: "acc-1/2026-03-02/" + isinIE + "/ETF1", Changes: []attention.Change{
			opened, {Seq: 2, State: attention.StateAcknowledged, Reason: attention.ReasonAcknowledged, At: minute(41)},
		}},
	}
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, l, app.WithJurisdiction(pl.NewPack(), fakeTerms{}), app.WithAttention(items),
		app.WithSyncHistory(fakeHistory{}), app.WithRecordedChanges(&fakeChanges{coupon: coupon}), app.WithInstrumentPrices(prices),
		app.WithBondTerms(fakeSeriesTerms{})),
		fakeVerifier{}, connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return api{handler: h, log: &log}
}

// serveJSON serves target, checks a 200 answer against the OpenAPI
// document, and decodes it into out unless out is nil; it fails t unless
// the status is want.
func serveJSON(t *testing.T, a api, target string, want int, out any) {
	t.Helper()
	req, rec := a.serve(t, http.MethodGet, target, goodToken)
	if rec.Code != want {
		t.Fatalf("GET %s = %d, want %d\n%s", target, rec.Code, want, rec.Body)
	}
	loadContract(t).check(t, req, rec)
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("decoding %s: %v", rec.Body, err)
		}
	}
}

func TestRESTAttentionInbox(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	type itemBody struct {
		ID                string `json:"id"`
		Group             string `json:"group"`
		AccountID         string `json:"accountId"`
		ISIN              string `json:"isin"`
		ConnectorInstance string `json:"connectorInstance"`
	}
	var got struct {
		Items        []itemBody `json:"items"`
		Open         int        `json:"open"`
		Acknowledged int        `json:"acknowledged"`
	}
	serveJSON(t, a, "/v1/attention-items", http.StatusOK, &got)
	want := []itemBody{
		{ID: "1", Group: "ATTENTION_ITEM_GROUP_DATA_COMING_IN", ConnectorInstance: "xtb-main"},
		{ID: "2", Group: "ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK", AccountID: "acc-1", ISIN: isinIE},
	}
	if diff := cmp.Diff(want, got.Items); diff != "" || got.Open != 1 || got.Acknowledged != 1 {
		t.Errorf("items (-want +got):\n%s\nopen %d, acknowledged %d", diff, got.Open, got.Acknowledged)
	}
	var acked struct {
		Items []itemBody `json:"items"`
	}
	serveJSON(t, a, "/v1/attention-items?view=ATTENTION_VIEW_ACKNOWLEDGED&accountId=acc-1", http.StatusOK, &acked)
	if len(acked.Items) != 1 || acked.Items[0].ID != "2" {
		t.Errorf("acc-1's acknowledged items = %+v", acked.Items)
	}
	serveJSON(t, a, "/v1/attention-items?view=7", http.StatusBadRequest, nil)
}

func TestRESTActivity(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var kinds []string
	target := "/v1/activity?pageSize=2"
	for range 10 {
		var page struct {
			Activities []struct {
				Kind string `json:"kind"`
				At   string `json:"at"`
			} `json:"activities"`
			NextPageToken string `json:"nextPageToken"`
		}
		serveJSON(t, a, target, http.StatusOK, &page)
		for _, act := range page.Activities {
			kinds = append(kinds, act.Kind+" "+act.At)
		}
		if page.NextPageToken == "" {
			break
		}
		target = "/v1/activity?pageSize=2&pageToken=" + url.QueryEscape(page.NextPageToken)
	}
	want := []string{
		"ACTIVITY_KIND_SYNC_FAILED 2026-11-04T06:40:00Z", "ACTIVITY_KIND_SYNC_FAILED 2026-11-04T06:30:00Z",
		"ACTIVITY_KIND_IMPORT 2026-11-04T06:25:00Z", "ACTIVITY_KIND_SYNC_FAILED 2026-11-04T06:20:00Z",
		"ACTIVITY_KIND_COUPON 2026-11-04T06:15:00Z", "ACTIVITY_KIND_SYNC 2026-11-04T06:10:00Z",
	}
	if diff := cmp.Diff(want, kinds); diff != "" {
		t.Errorf("activity (-want +got):\n%s", diff)
	}
	serveJSON(t, a, "/v1/activity?pageToken=not-a-token", http.StatusBadRequest, nil)
}

// TestConnectActivityDetails reads the activity with a generated client
// and checks what each kind carries: a batch's counts, a failed attempt's
// class, a coupon's instrument and cash.
func TestConnectActivityDetails(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(uiv2API(t).handler)
	t.Cleanup(srv.Close)
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	page, err := c.ListActivity(t.Context(), authed(goodToken, &viewsv1.ListActivityRequest{}))
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	acts := page.Msg.GetActivities()
	if len(acts) != 6 {
		t.Fatalf("activities = %d, want 6", len(acts))
	}
	type detail struct {
		Instance, Source                  string
		Accounts                          []string
		Counts                            []string
		Transactions, Unsupported, Failed int32
		Class                             viewsv1.SyncErrorClass
		Account, ISIN, Cash               string
		Kind                              viewsv1.TransactionKind
	}
	upload, failed, coupon := acts[2].GetBatch(), acts[0].GetAttempt(), acts[4].GetChange()
	counts := make([]string, 0, len(upload.GetAccounts()))
	for _, c := range upload.GetAccounts() {
		counts = append(counts, fmt.Sprintf("%s %d %d %d %d", c.GetAccountId(), c.GetTransactions(), c.GetNewTransactions(), c.GetCorrected(), c.GetFailedAssertions()))
	}
	got := []detail{
		{Instance: acts[2].GetConnectorInstance(), Source: upload.GetSource(), Accounts: upload.GetAccountIds(), Counts: counts, Transactions: upload.GetTransactions(), Unsupported: upload.GetUnsupported()},
		{Instance: acts[0].GetConnectorInstance(), Failed: failed.GetFailures(), Class: failed.GetErrorClass()},
		{Account: coupon.GetAccountId(), ISIN: coupon.GetInstrument().GetIsin(), Cash: coupon.GetAmount().GetAmount(), Kind: coupon.GetKind()},
	}
	want := []detail{
		{Instance: "files", Source: "upload", Accounts: []string{"acc-1", "acc-2"}, Counts: []string{"acc-1 40 36 1 0", "acc-2 2 2 0 1"}, Transactions: 42, Unsupported: 4},
		{Instance: "xtb-main", Failed: 3, Class: viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_TIMEOUT},
		{Account: "acc-1", ISIN: isinPL, Cash: "10.12", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_INTEREST},
	}
	if diff := cmp.Diff(want, got); diff != "" || !failed.GetFailed() {
		t.Errorf("the upload, the failed attempt and the coupon (-want +got):\n%s\nfailed %v", diff, failed.GetFailed())
	}
}

func TestRESTSyncs(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var statuses struct {
		Statuses []struct {
			ConnectorInstance string `json:"connectorInstance"`
			Failures          int    `json:"failures"`
			FailingSince      string `json:"failingSince"`
			LastErrorClass    string `json:"lastErrorClass"`
			KeepsFailing      bool   `json:"keepsFailing"`
		} `json:"statuses"`
	}
	serveJSON(t, a, "/v1/sync-status", http.StatusOK, &statuses)
	if len(statuses.Statuses) != 1 || !statuses.Statuses[0].KeepsFailing || statuses.Statuses[0].LastErrorClass != "SYNC_ERROR_CLASS_TIMEOUT" ||
		statuses.Statuses[0].FailingSince != "2026-11-04T06:19:00Z" {
		t.Errorf("statuses = %+v", statuses)
	}
}

func TestRESTSyncAttempts(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var attempts struct {
		Attempts []struct {
			EndedAt    string `json:"endedAt"`
			ErrorClass string `json:"errorClass"`
		} `json:"attempts"`
		NextPageToken string `json:"nextPageToken"`
	}
	serveJSON(t, a, "/v1/connector-instances/xtb-main/sync-attempts?pageSize=3", http.StatusOK, &attempts)
	if len(attempts.Attempts) != 3 || attempts.NextPageToken == "" || attempts.Attempts[0].EndedAt != "2026-11-04T06:40:00Z" {
		t.Errorf("attempts = %+v", attempts)
	}
	var rest struct {
		Attempts []struct {
			ErrorClass string `json:"errorClass"`
		} `json:"attempts"`
		NextPageToken string `json:"nextPageToken"`
	}
	serveJSON(t, a, "/v1/connector-instances/xtb-main/sync-attempts?pageToken="+url.QueryEscape(attempts.NextPageToken), http.StatusOK, &rest)
	if len(rest.Attempts) != 1 || rest.Attempts[0].ErrorClass != "" || rest.NextPageToken != "" {
		t.Errorf("second page = %+v", rest)
	}
	serveJSON(t, a, "/v1/connector-instances/nbp-fx/sync-attempts", http.StatusNotFound, nil)
}

func TestRESTInstrument(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var got struct {
		Instrument struct {
			ISIN     string `json:"isin"`
			Listings []struct {
				Instrument  struct{ Symbol string } `json:"instrument"`
				PriceSource string                  `json:"priceSource"`
				LatestPrice *struct {
					Close struct{ Amount string } `json:"close"`
				} `json:"latestPrice"`
			} `json:"listings"`
			AssetClass       string `json:"assetClass"`
			ExemptionClasses []struct {
				Wrapper string `json:"wrapper"`
			} `json:"exemptionClasses"`
		} `json:"instrument"`
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE, http.StatusOK, &got)
	in := got.Instrument
	if in.ISIN != isinIE || len(in.Listings) != 1 || in.Listings[0].PriceSource != "test-prices" || in.Listings[0].LatestPrice == nil ||
		in.Listings[0].LatestPrice.Close.Amount != "10.49" || in.AssetClass != "listed" || len(in.ExemptionClasses) == 0 {
		t.Errorf("instrument = %+v", in)
	}
	serveJSON(t, a, "/v1/instruments/XX0000000001", http.StatusNotFound, nil)
}

func TestRESTInstrumentPrices(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var prices struct {
		Prices []struct {
			Source  string `json:"source"`
			Primary bool   `json:"primary"`
		} `json:"prices"`
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE+"/prices?from=2026-03-01&to=2026-03-02", http.StatusOK, &prices)
	if len(prices.Prices) != 2 || prices.Prices[0].Primary || !prices.Prices[1].Primary {
		t.Errorf("prices = %+v", prices)
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE+"/prices?from=2026-03-02&to=2026-03-01", http.StatusBadRequest, nil)
	var checks struct {
		Checks []json.RawMessage `json:"checks"`
	}
	serveJSON(t, a, "/v1/price-checks?isin="+isinIE, http.StatusOK, &checks)
}

func TestRESTInstrumentHoldings(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var holdings struct {
		Holdings []struct {
			AccountID string            `json:"accountId"`
			Quantity  string            `json:"quantity"`
			Lots      []json.RawMessage `json:"lots"`
		} `json:"holdings"`
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE+"/holdings?date=2026-03-02", http.StatusOK, &holdings)
	if len(holdings.Holdings) != 1 || holdings.Holdings[0].AccountID != "acc-1" || holdings.Holdings[0].Quantity != "3" || len(holdings.Holdings[0].Lots) != 1 {
		t.Errorf("holdings = %+v", holdings)
	}
	var txs struct {
		Transactions []struct {
			ExternalID string `json:"externalId"`
			AccountID  string `json:"accountId"`
		} `json:"transactions"`
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE+"/transactions", http.StatusOK, &txs)
	if len(txs.Transactions) != 1 || txs.Transactions[0].ExternalID != "buy-1" || txs.Transactions[0].AccountID != "acc-1" {
		t.Errorf("transactions = %+v", txs)
	}
}

func TestUIv2EnumsMatch(t *testing.T) {
	t.Parallel()
	for k := views.ActivityImport; k <= views.ActivityCorrection; k++ {
		name := map[views.ActivityKind]string{
			views.ActivityImport: "import", views.ActivitySync: "sync", views.ActivitySyncFailed: "sync_failed",
			views.ActivityCoupon: "coupon", views.ActivityCorrection: "correction",
		}[k]
		checkEnum(t, viewsv1.ActivityKind(0).Descriptor(), "ACTIVITY_KIND_", int(k), name)
	}
	for c := ingest.SyncErrorTimeout; c <= ingest.SyncErrorOther; c++ {
		checkEnum(t, viewsv1.SyncErrorClass(0).Descriptor(), "SYNC_ERROR_CLASS_", int(c), c.String())
	}
	for g := attention.GroupDataComingIn; g <= attention.GroupDecisions; g++ {
		checkEnum(t, viewsv1.AttentionItemGroup(0).Descriptor(), "ATTENTION_ITEM_GROUP_", int(g), g.String())
	}
}
