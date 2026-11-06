package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"testing"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errSyncRuns is the fake core's refusal of a retry while a sync runs.
var errSyncRuns = errors.New("a sync of the instance is already running")

// fakeSyncs is an in-memory SyncService: it wants the owner's session,
// refuses a retry of busy-main as running, and records the instances it
// started.
type fakeSyncs struct {
	ingestv1connect.UnimplementedSyncServiceHandler

	mu      sync.Mutex
	retried []string
}

func (f *fakeSyncs) RetrySync(_ context.Context, r *connect.Request[ingestv1.RetrySyncRequest]) (*connect.Response[ingestv1.RetrySyncResponse], error) {
	if r.Header().Get("Authorization") != "Bearer "+pluginToken {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if r.Msg.GetInstance() == "busy-main" {
		return nil, codeErr(connect.CodeFailedPrecondition, errSyncRuns)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retried = append(f.retried, r.Msg.GetInstance())
	return connect.NewResponse(&ingestv1.RetrySyncResponse{}), nil
}

// ListSyncAttempts answers xtb-main's history: three timeouts in a row
// since a sync that went through; any other instance was never synced.
func (f *fakeRead) ListSyncAttempts(_ context.Context, r *connect.Request[viewsv1.ListSyncAttemptsRequest]) (*connect.Response[viewsv1.ListSyncAttemptsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	if r.Msg.GetConnectorInstance() != "xtb-main" {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	timeout := func(start, next string, n int32) *viewsv1.SyncAttempt {
		return &viewsv1.SyncAttempt{
			ConnectorInstance: "xtb-main", BatchId: "b-2", Mode: "transactions", StartedAt: start, EndedAt: start, Failed: true,
			ErrorClass: viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_TIMEOUT, Failures: n, NextSyncAt: next,
		}
	}
	return connect.NewResponse(&viewsv1.ListSyncAttemptsResponse{
		Status: &viewsv1.SyncStatus{
			ConnectorInstance: "xtb-main", LastSuccessAt: "2026-03-02T21:40:00Z", Failures: 3, FailingSince: "2026-03-03T06:00:00Z",
			LastErrorClass: viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_TIMEOUT, NextSyncAt: "2026-03-03T12:00:00Z", KeepsFailing: true,
		},
		Attempts: []*viewsv1.SyncAttempt{
			timeout("2026-03-03T07:00:00Z", "2026-03-03T12:00:00Z", 3), timeout("2026-03-03T06:30:00Z", "2026-03-03T07:00:00Z", 2),
			timeout("2026-03-03T06:00:00Z", "2026-03-03T06:30:00Z", 1),
			{ConnectorInstance: "xtb-main", BatchId: "b-1", Mode: "transactions", StartedAt: "2026-03-02T21:40:00Z", EndedAt: "2026-03-02T21:40:00Z"},
		},
	}), nil
}

// syncItem adds the open item of a sync of instance that keeps failing.
func syncItem(h *harness, id int64, instance string) {
	h.attention.add(&viewsv1.AttentionItem{
		Id: id, Kind: "connector-sync-failing", Subject: instance, Detail: "failing since 2026-03-03T06:00:00Z",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN,
		ConnectorInstance: instance,
	})
}

// retryPost posts the owner's "Retry now" on instance.
func retryPost(instance string) *http.Request {
	return form("/inbox/retry-sync", url.Values{"instance": {instance}}.Encode())
}

// TestInboxFailingSync checks a sync that keeps failing in the Inbox
// (D-196): its timeline of attempts, error classes and next try, and the
// owner's "Retry now", which reaches the core and comes back saying the
// sync started; a retry the core refuses because a sync runs shows its
// reason; a read session sees the timeline without the button.
func TestInboxFailingSync(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	syncItem(h, 40, "xtb-main")
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The xtb-main sync keeps failing", `<span class="pill pill-danger">Failing</span>`,
		"What happened: 3 failed attempt(s) in a row", "Last good sync: <span class=\"mono\">2026-03-02T21:40:00Z</span>",
		"failed, timeout; next try at", "went through", `action="/inbox/retry-sync"`, `name="instance" value="xtb-main"`, "Retry now")
	read := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), readSession))
	wantBody(t, read, "What happened: 3 failed attempt(s) in a row")
	wantNoBody(t, read, `action="/inbox/retry-sync"`)

	rec = h.serve(t, withSession(retryPost("xtb-main"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox?retried=xtb-main" {
		t.Errorf("Location = %q, want /inbox?retried=xtb-main", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox?retried=xtb-main", ""), ownerSession))
	wantBody(t, rec, "A sync of xtb-main has started.")
	rec = h.serve(t, withSession(retryPost("busy-main"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this: "+errSyncRuns.Error())
	rec = h.serve(t, withSession(retryPost(""), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	rec = h.serve(t, withSession(retryPost("xtb-main"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	h.syncs.mu.Lock()
	defer h.syncs.mu.Unlock()
	if !slices.Equal(h.syncs.retried, []string{"xtb-main"}) {
		t.Errorf("retries started = %v, want xtb-main once", h.syncs.retried)
	}
}

// TestRemovedCheckPagesRedirect checks that the pages D-192 removed send
// the browser to the Inbox for good, with or without a session.
func TestRemovedCheckPagesRedirect(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, path := range []string{"/attention", "/tickers", "/assertions", "/price-checks", "/transfers"} {
		rec := h.serve(t, request(http.MethodGet, path, ""))
		wantStatus(t, rec, http.StatusMovedPermanently)
		if got := rec.Header().Get("Location"); got != "/inbox" {
			t.Errorf("GET %s: Location = %q, want /inbox", path, got)
		}
	}
}

// TestNavigation checks D-192's four places: Overview, Accounts, Inbox
// with its open count and Settings, which the owner opens on the general
// settings and a read session on the sleeves; the sleeves and plugins sit
// under Settings.
func TestNavigation(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, rec, `<a href="/" aria-current="page">Overview</a>`, `<a href="/accounts">Accounts</a>`, `<a href="/inbox">Inbox&nbsp;`,
		`<a href="/settings">Settings</a>`)
	wantNoBody(t, rec, `href="/attention"`, `href="/tickers"`, `href="/price-checks"`, `href="/transfers"`, `href="/assertions"`,
		`<a href="/sleeves">`, `<a href="/plugins">`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), readSession))
	wantBody(t, rec, `<a href="/sleeves" aria-current="page">Settings</a>`, `<a href="/sleeves" aria-current="page">Sleeves</a>`,
		`<a href="/plugins">Plugins</a>`)
	wantNoBody(t, rec, `href="/settings"`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/plugins", ""), ownerSession))
	wantBody(t, rec, `<a href="/settings" aria-current="page">Settings</a>`, `<a href="/settings">General</a>`,
		`<a href="/plugins" aria-current="page">Plugins</a>`)
}

// TestAccountBanner checks that an account's page shows its own open
// items at the top, linked to the Inbox, a broken transfer's included
// through its second account, and nothing of another account's (D-196).
func TestAccountBanner(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, `class="banner"`)
	h.attention.add(&viewsv1.AttentionItem{
		Id: 50, Kind: "balance-assertion-flagged", Subject: accountID + "/2026-03-15/cash:EUR", Detail: "reported 100.00, derived 90.00",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK,
		AccountId: accountID, AccountIds: []string{accountID}, OwnerResolves: true,
	}, &viewsv1.AttentionItem{
		Id: 51, Kind: "transfer-broken", Subject: "test/w9 -> test/d9", Detail: "its legs no longer match",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS,
		AccountId: "acc-other", AccountIds: []string{"acc-other", accountID},
	})
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantBody(t, rec, `class="banner"`, "Needs you · 2", `<a href="/inbox#item-50">A balance in acc-synthetic-1 doesn&#39;t match</a>`,
		`<a href="/inbox#item-51">A confirmed transfer no longer matches</a>`, `<a href="/inbox">Open the Inbox</a>`)
	wantNoBody(t, rec, "Map a broker account")

	inbox := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantBody(t, inbox, `id="item-50"`, "2026-03-15", "cash EUR", "100,00", "90,00", `action="/inbox/resolve"`)
}

// TestInboxMapsBrokerAccount checks that the owner maps an unmapped broker
// account in place (D-72, D-196): its item offers the instance's unmapped
// broker accounts and the accounts, and the form goes back to the Inbox.
func TestInboxMapsBrokerAccount(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(&viewsv1.AttentionItem{
		Id: 60, Kind: "broker-account-unmapped", Subject: "ibkr-main", Detail: "last seen 2026-03-31",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN,
		ConnectorInstance: "ibkr-main",
	})
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantBody(t, rec, `action="/accounts/bind"`, `name="next" value="/inbox"`, `<option value="ibkr-main/U0000001">ibkr-main/U0000001</option>`,
		`<option value="`+accountID+`">`)
	wantNoBody(t, rec, "ibkr-main/U0000002")
	post := form("/accounts/bind", url.Values{"broker": {"ibkr-main/U0000001"}, "account": {accountID}, "next": {"/inbox"}}.Encode())
	rec = h.serve(t, withSession(post, ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox" {
		t.Errorf("Location = %q, want /inbox", got)
	}
}

// TestInboxEmpty checks the open view with nothing open and no transfer
// to decide: Kasztan asleep and one line.
func TestInboxEmpty(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.mu.Lock()
	h.attention.items = nil
	h.attention.mu.Unlock()
	h.transfers.mu.Lock()
	h.transfers.decision = ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED
	h.transfers.mu.Unlock()
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Nothing needs you.", "kasztan-asleep")
}
