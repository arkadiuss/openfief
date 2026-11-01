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
// author is asked about them (implementation questions). A deprecated
// method answers UNIMPLEMENTED and serves nothing, so it needs no place
// (retired).
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
	}
}

const beyondD137 = "beyond D-137's list"

// retired reports whether m is deprecated: it answers UNIMPLEMENTED and
// serves nothing, as the read API's broker accounts did before their
// removal (D-121).
func retired(m protoreflect.MethodDescriptor) bool {
	opts, ok := m.Options().(*descriptorpb.MethodOptions)
	return ok && opts.GetDeprecated()
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
		if retired(m) {
			continue
		}
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
	if f.FullName() == "openfief.views.v1.ListTransactionsResponse.next_page_token" {
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
// broker account (D-121, D-137). A deprecated method, which answers
// UNIMPLEMENTED, is skipped.
func TestReadAccessReturnsNoSecrets(t *testing.T) {
	t.Parallel()
	svc := viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService")
	if svc == nil || svc.Methods().Len() == 0 {
		t.Fatal("ReadService has no methods")
	}
	for i := range svc.Methods().Len() {
		m := svc.Methods().Get(i)
		if retired(m) {
			continue
		}
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
