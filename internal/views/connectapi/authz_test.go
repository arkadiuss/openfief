package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
)

// acc1Only lets the caller fakeVerifier accepts view acc-1 alone, and
// nobody anything else, as a SpiceDB that granted that caller acc-1 only
// would (D-163).
type acc1Only struct{}

func (acc1Only) MayAct(_ context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	return c == identity.Caller{Holder: identity.HolderScriptToken, ID: "test"} && a == identity.ActionView && account == "acc-1", nil
}

// TestRESTServesOnlyViewableAccounts checks, through the REST transcoding,
// that the read API asks the authorizer about the request's caller: it
// lists only the accounts the caller may view, and answers for another one
// exactly as for an account not stored (D-163).
func TestRESTServesOnlyViewableAccounts(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	l.accounts = append(l.accounts, ledger.Account{ID: "acc-2", Institution: "broker-b", Name: "Synthetic B"})
	h, err := connectapi.New(newViews(t, l, app.WithAuthorizer(acc1Only{})), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	a := api{handler: h, log: &bytes.Buffer{}}
	_, rec := a.serve(t, http.MethodGet, "/v1/accounts", goodToken)
	var listed struct {
		Accounts []struct {
			ID string `json:"id"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/accounts = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(listed.Accounts) != 1 || listed.Accounts[0].ID != "acc-1" {
		t.Errorf("GET /v1/accounts listed %s, want acc-1 alone", rec.Body)
	}
	for _, path := range []string{"", "/transactions", "/holdings?date=2026-03-02", "/lots?date=2026-03-02", "/cash?date=2026-03-02", "/balance-assertions", "/valuations?from=2026-03-01&to=2026-03-02", "/valuation-status"} {
		checkHidden(t, a, path)
	}
}

// checkHidden checks that path under acc-1 answers, and under acc-2 answers
// exactly as under an account not stored.
func checkHidden(t *testing.T, a api, path string) {
	t.Helper()
	_, viewable := a.serve(t, http.MethodGet, "/v1/accounts/acc-1"+path, goodToken)
	_, hidden := a.serve(t, http.MethodGet, "/v1/accounts/acc-2"+path, goodToken)
	_, missing := a.serve(t, http.MethodGet, "/v1/accounts/nope"+path, goodToken)
	if viewable.Code != http.StatusOK {
		t.Errorf("GET acc-1%s = %d %s, want 200", path, viewable.Code, viewable.Body)
	}
	if hidden.Code != http.StatusNotFound || missing.Code != http.StatusNotFound {
		t.Errorf("GET acc-2%s = %d, nope%s = %d; want both 404", path, hidden.Code, path, missing.Code)
	}
	if diff := cmp.Diff(missing.Body.String(), hidden.Body.String()); diff != "" {
		t.Errorf("GET acc-2%s tells a hidden account from a missing one (-missing +hidden):\n%s", path, diff)
	}
}
