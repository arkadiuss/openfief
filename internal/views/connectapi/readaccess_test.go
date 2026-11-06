package connectapi_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// readAccess maps every method of the read API to what of D-137's read
// access it serves. A method missing here fails TestReadAccessMethods, so
// nothing joins the read API without being placed against D-137. The
// methods marked beyondD137 serve what D-137's list doesn't name; the
// author is asked about them (implementation questions). Every method
// needs a place, a deprecated one too: the read API keeps none that serves
// nothing.
func readAccess() map[protoreflect.Name]string {
	return map[protoreflect.Name]string{
		"ListSleeves":           "sleeves",
		"ListSleeveKinds":       "sleeves",
		"ListAccounts":          "accounts",
		"GetAccount":            "accounts",
		"ListConnectorBindings": "accounts",
		"ListTransactions":      "transactions",
		"ListLots":              "lots",
		"ListHoldings":          "holdings",
		"ListCash":              "cash",
		"ListAccountValuations": "valuations",
		"GetValuationStatus":    "valuations",
		"ListValuationTotals":   "valuations",
		"ListValuationStatuses": "valuations",
		"GetAllocation":         "allocation",
		"GetPerformance":        "performance",
		"GetModelPortfolio":     "model portfolio",
		"ListPriceChecks":       "price checks",
		"ListTransfers":         "transfers",
		"ListAttentionItems":    "attention items",
		"ListPlugins":           "installed plugins",
		"GetSettings":           beyondD137,
		"ListYearlyTaxes":       beyondD137,
		"ListAssertionChecks":   beyondD137,
		// D-193's recent activity: coupons and corrections are
		// transactions, but import batches and sync attempts are not on
		// D-137's list.
		"ListActivity": beyondD137,
		// D-196's failing sync: the connectors' sync history.
		"ListSyncStatuses": beyondD137,
		"ListSyncAttempts": beyondD137,
		// D-195's instrument page: its listings, prices and exemption
		// classes, then its holdings, lots and transactions across accounts.
		"GetInstrument":              beyondD137,
		"ListInstrumentPrices":       beyondD137,
		"ListInstrumentHoldings":     "holdings",
		"ListInstrumentTransactions": "transactions",
	}
}

const beyondD137 = "beyond D-137's list"

// removed names what an author decision took off the read API, with the
// decision: D-157 removed ListBrokerAccounts and its messages rather than
// keep them deprecated, since the read API never serves a broker account
// (D-121).
func removed() map[protoreflect.FullName]string {
	return map[protoreflect.FullName]string{
		"openfief.views.v1.ReadService.ListBrokerAccounts": "D-157",
		"openfief.views.v1.ListBrokerAccountsRequest":      "D-157",
		"openfief.views.v1.ListBrokerAccountsResponse":     "D-157",
		"openfief.views.v1.BrokerAccount":                  "D-157",
	}
}

// TestReadAccessRemovedMethods checks that what removed names is absent
// from the read API's contract, as a method and as a message alike, and
// that no method left is deprecated (D-157).
func TestReadAccessRemovedMethods(t *testing.T) {
	t.Parallel()
	file := viewsv1.File_openfief_views_v1_read_service_proto
	svc := file.Services().ByName("ReadService")
	if svc == nil {
		t.Fatal("no ReadService")
	}
	for name, decision := range removed() {
		if d, err := protoregistry.GlobalFiles.FindDescriptorByName(name); err == nil {
			t.Errorf("%s removed it, but the read API still has %s", decision, d.FullName())
		}
	}
	for i := range svc.Methods().Len() {
		m := svc.Methods().Get(i)
		if opts, ok := m.Options().(*descriptorpb.MethodOptions); ok && opts.GetDeprecated() {
			t.Errorf("%s is deprecated: remove it under an author decision instead (D-157)", m.Name())
		}
	}
}

// TestReadAccessMethods enumerates the read API's methods and checks that
// each is placed against D-137's read access (readAccess), that every kind
// of record D-137 names is served by some method (checkServed), and that
// no method changes state: each declares no side effects and is a GET
// without a body (D-137).
func TestReadAccessMethods(t *testing.T) {
	t.Parallel()
	svc := viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService")
	if svc == nil || svc.Methods().Len() == 0 {
		t.Fatal("ReadService has no methods")
	}
	access := readAccess()
	served := map[string]bool{}
	for i := range svc.Methods().Len() {
		m := svc.Methods().Get(i)
		what, ok := access[m.Name()]
		if !ok {
			t.Errorf("%s is not placed against D-137's read access: add it to readAccess", m.Name())
			continue
		}
		served[what] = true
		checkReadOnly(t, m)
	}
	for name := range access {
		if svc.Methods().ByName(name) == nil {
			t.Errorf("readAccess names %s, which the read API doesn't have", name)
		}
	}
	checkServed(t, served)
}

// checkServed fails t unless served holds every kind of record D-137's
// read access names.
func checkServed(t *testing.T, served map[string]bool) {
	t.Helper()
	for _, what := range []string{
		"sleeves", "accounts", "transactions", "lots", "holdings", "cash", "valuations", "allocation",
		"performance", "model portfolio", "price checks", "transfers", "attention items", "installed plugins",
	} {
		if !served[what] {
			t.Errorf("no read API method serves %s (D-137)", what)
		}
	}
}

// namedLikeASecret reports whether f's name holds a word read access never
// includes: a secret, credential, passkey, session or token (D-137), or a
// broker account number (D-121). A page token is not a secret: it is an
// opaque cursor into a listing that the caller sends back for the next
// page, and it proves nothing.
func namedLikeASecret(f protoreflect.FieldDescriptor) bool {
	if f.Name() == "next_page_token" {
		return false
	}
	name := strings.ToLower(string(f.Name()))
	return slices.ContainsFunc([]string{
		"secret", "credential", "passkey", "password", "session", "token", "cookie", "api_key", "private_key",
		"broker_account", "account_number",
	}, func(w string) bool { return strings.Contains(name, w) })
}

// TestReadAccessReturnsNoSecrets walks every message each served method
// returns, nested ones included, and checks that none has a field named
// for a secret, credential, passkey, session or token, and none is a
// broker account (D-121, D-137).
func TestReadAccessReturnsNoSecrets(t *testing.T) {
	t.Parallel()
	svc := viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService")
	if svc == nil || svc.Methods().Len() == 0 {
		t.Fatal("ReadService has no methods")
	}
	for i := range svc.Methods().Len() {
		m := svc.Methods().Get(i)
		seen := map[protoreflect.FullName]bool{}
		walkMessage(t, m.Name(), m.Output(), seen)
	}
}

// walkMessage checks md and every message its fields hold.
func walkMessage(t *testing.T, method protoreflect.Name, md protoreflect.MessageDescriptor, seen map[protoreflect.FullName]bool) {
	t.Helper()
	if seen[md.FullName()] {
		return
	}
	seen[md.FullName()] = true
	if md.Name() == "BrokerAccount" {
		t.Errorf("%s returns a BrokerAccount (D-121)", method)
	}
	for i := range md.Fields().Len() {
		f := md.Fields().Get(i)
		if namedLikeASecret(f) {
			t.Errorf("%s returns %s.%s, named like a secret or a broker account number (D-137)", method, md.Name(), f.Name())
		}
		if f.Message() != nil {
			walkMessage(t, method, f.Message(), seen)
		}
	}
}

// TestReadAccessAttentionNamesNoBrokerAccount checks that the read API's
// attention items never carry a broker account number (D-121, D-137):
// an unmapped broker account's item names only its connector instance,
// and other subjects are served as they are.
func TestReadAccessAttentionNamesNoBrokerAccount(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	raised := []attention.Change{{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "last seen 2026-03-01", At: at}}
	items := fakeItems{
		{ID: 1, Kind: attention.KindBrokerAccountUnmapped, Subject: "flex-main/U0000007", Changes: raised},
		{ID: 2, Kind: attention.KindTickerUnconfirmed, Subject: "xtb-main/TESTA.PL", Changes: raised},
	}
	h, err := connectapi.New(newViews(t, newLedger(t), app.WithJurisdiction(pl.NewPack(), fakeTerms{}), app.WithAttention(items)),
		fakeVerifier{}, connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	_, rec := api{handler: h}.serve(t, http.MethodGet, "/v1/attention-items", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/attention-items = %d\n%s", rec.Code, rec.Body)
	}
	body := rec.Body.String()
	if strings.Contains(body, "U0000007") {
		t.Errorf("the read API serves a broker account number: %s", body)
	}
	if !strings.Contains(body, `"subject":"flex-main"`) || !strings.Contains(body, `"subject":"xtb-main/TESTA.PL"`) {
		t.Errorf("attention subjects = %s, want flex-main and xtb-main/TESTA.PL", body)
	}
}
