package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1/attentionv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// attentionAPI is a generated AttentionService client.
type attentionAPI struct {
	attentionv1connect.AttentionServiceClient
}

// attentionClient returns an AttentionService client of srv that presents
// token as the view plugin's.
func attentionClient(srv *httptest.Server, token string) attentionAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return attentionAPI{attentionv1connect.NewAttentionServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// TestAttentionService checks that the core's plugin port serves the
// owner's actions on attention items (D-130, D-79): the owner's view
// reaches the items, which answer that an item they don't hold is not
// found, while another caller and an unknown session are refused first.
func TestAttentionService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	tests := []struct {
		name, caller, session string
		want                  connect.Code
	}{
		{"the owner's view, no such item", uiCaller, oc.session, connect.CodeNotFound},
		{"not a view plugin", "wrong", oc.session, connect.CodePermissionDenied},
		{"an unknown session", uiCaller, "ofsess_nope_nope", connect.CodeUnauthenticated},
	}
	for _, tt := range tests {
		_, err := attentionClient(oc.srv, tt.caller).AcknowledgeAttentionItem(t.Context(), connect.NewRequest(&attentionv1.AcknowledgeAttentionItemRequest{
			Origin: localUI, SessionToken: tt.session, ItemId: 1,
		}))
		if connect.CodeOf(err) != tt.want {
			t.Errorf("%s: AcknowledgeAttentionItem = %v, want %s", tt.name, err, tt.want)
		}
	}
}

// checkUnmappedItem fails t unless got lists one open item, the unmapped
// broker account, raised once, named by its instance alone: the read API
// never serves a broker account number (D-121, D-137).
func checkUnmappedItem(t *testing.T, got *viewsv1.ListAttentionItemsResponse) {
	t.Helper()
	if len(got.GetItems()) != 1 || got.GetOpen() != 1 {
		t.Fatalf("GET /v1/attention-items = %v, want one open item", got)
	}
	it := got.GetItems()[0]
	if it.GetKind() != "broker-account-unmapped" || it.GetSubject() != "flex-main" || len(it.GetHistory()) != 1 {
		t.Errorf("item = %v, want the unmapped broker account, raised once", it)
	}
}

// TestAttentionChecksRaiseItems checks that a tick of the core's checks
// raises an item for an unmapped broker account, which the read API lists
// once however often the checks run, and that the outbox's events are
// published.
func TestAttentionChecksRaiseItems(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	c, srv, pool := startCore(t, dsn)
	day, err := kernel.ParseDate("2026-10-27")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if err := ingestpg.New(pool).RecordBrokerAccounts(t.Context(), "flex-main", "ibkr", []ingest.BrokerAccountID{"U0000001"}, day); err != nil {
		t.Fatalf("RecordBrokerAccounts: %v", err)
	}
	for range 2 {
		c.attention.tick(t.Context(), quiet())
	}
	issued := tokenCmd(t, dsn, "issue", "--name", "attention-test", "--ttl", "1h")
	secret, _, _ := strings.Cut(issued, "\n")
	var got viewsv1.ListAttentionItemsResponse
	getJSON(t, srv, "/v1/attention-items", secret, &got)
	checkUnmappedItem(t, &got)
	if events, err := c.attention.relay.Outbox.Unpublished(t.Context(), 10); err != nil || len(events) != 0 {
		t.Errorf("unpublished events after a tick = %v, %v; want none", events, err)
	}
}
