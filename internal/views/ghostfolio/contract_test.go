package ghostfolio_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/areknoster/hypert"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// recordGhostfolio switches TestGhostfolioContract to record against a
// running Ghostfolio. Committed code always replays: turn it on only
// locally, following .claude/skills/hypert-recording/SKILL.md, then turn it
// off.
//
// The recordings under testdata/TestGhostfolioContract come from a fresh,
// local Ghostfolio 3.93.0 (the docker.io/ghostfolio/ghostfolio:3.93.0 image
// on 127.0.0.1:3333, with its own PostgreSQL and Redis) and a new anonymous
// user, holding only the synthetic record of fakes_test.go. To record
// again, start such an instance, create a user with
// POST /api/v1/user, and run this test with the user's security token in
// GHOSTFOLIO_SECURITY_TOKEN. The token never reaches a file: the login
// body is sanitized, and the bearer token in the login response is swapped
// for a placeholder before hypert writes it (sessionSwap).
const recordGhostfolio = false

// The recorded Ghostfolio's address and the placeholder for every secret.
const (
	recordedBaseURL = "http://127.0.0.1:3333"
	sanitized       = "SANITIZED"
	loginPath       = "/api/v1/auth/anonymous"
)

// contractClient is the hypert client of the contract test: it sanitizes
// the login body as well as hypert's default headers, and validates every
// request's JSON body against the recording as well as its method, path,
// query and headers.
func contractClient(t *testing.T) *http.Client {
	t.Helper()
	parent := &http.Client{}
	if recordGhostfolio {
		parent.Transport = &sessionSwap{next: http.DefaultTransport}
	}
	return hypert.TestClient(t, recordGhostfolio,
		hypert.WithParentHTTPClient(parent),
		hypert.WithRequestSanitizer(hypert.ComposedRequestSanitizer(hypert.DefaultRequestSanitizer(), loginBodySanitizer())),
		hypert.WithRequestValidator(hypert.ComposedRequestValidator(
			hypert.PathValidator(), hypert.MethodValidator(), hypert.QueryParamsValidator(),
			hypert.SchemeValidator(), hypert.HeadersValidator(), jsonBodyValidator(),
		)))
}

// loginBodySanitizer replaces the login body, which carries the security
// token, with a placeholder.
func loginBodySanitizer() hypert.RequestSanitizerFunc {
	return hypert.RequestSanitizerFunc(func(req *http.Request) *http.Request {
		if req.URL.Path == loginPath {
			// While recording, hypert sends the body this one is read
			// through, so it must be read to the end first.
			if _, err := io.Copy(io.Discard, req.Body); err != nil {
				return req
			}
			body := `{"accessToken":"` + sanitized + `"}`
			req.Body = io.NopCloser(strings.NewReader(body))
			req.ContentLength = int64(len(body))
		}
		return req
	})
}

// jsonBodyValidator checks that a request's body is the recorded JSON, so
// the contract covers what the view sends, not only where.
func jsonBodyValidator() hypert.RequestValidatorFunc {
	return hypert.RequestValidatorFunc(func(t hypert.T, recorded, got hypert.RequestData) error {
		if len(recorded.BodyBytes) == 0 && len(got.BodyBytes) == 0 {
			return nil
		}
		var want, have any
		if err := json.Unmarshal(recorded.BodyBytes, &want); err != nil {
			return fmt.Errorf("recorded body of %s: %w", recorded, err)
		}
		if err := json.Unmarshal(got.BodyBytes, &have); err != nil {
			return fmt.Errorf("body of %s: %w", got, err)
		}
		if diff := cmp.Diff(want, have); diff != "" {
			t.Errorf("body of %s (-recorded +got):\n%s", got, diff)
		}
		return nil
	})
}

// sessionSwap keeps Ghostfolio's bearer token out of the recordings: it
// hands the client a placeholder instead of the token from the login
// response, and puts the real one back on every request. Used only while
// recording.
type sessionSwap struct {
	next  http.RoundTripper
	token string
}

func (s *sessionSwap) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "Bearer "+sanitized {
		req = req.Clone(req.Context())
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.next.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("recording: %w", err)
	}
	if req.URL.Path != loginPath {
		return resp, nil
	}
	var out struct {
		AuthToken string `json:"authToken"`
	}
	data, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); err != nil || closeErr != nil {
		return nil, errors.Join(err, closeErr)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("login response: %w", err)
	}
	s.token = out.AuthToken
	body := []byte(`{"authToken":"` + sanitized + `"}`)
	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))
	resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
	return resp, nil
}

// securityToken is the user's security token while recording, and the
// placeholder while replaying.
func securityToken() (string, error) {
	if recordGhostfolio {
		return os.Getenv("GHOSTFOLIO_SECURITY_TOKEN"), nil
	}
	return sanitized, nil
}

// TestGhostfolioContract is the contract test of the supported Ghostfolio
// versions (D-64): it replays the reconcile loop against recordings of a
// real Ghostfolio 3.93.0. The first sync creates the owned accounts and
// imports; a second sync against what Ghostfolio stored changes nothing,
// which proves every field survives Ghostfolio's round trip; a manual edit
// is reverted; a removed transaction is deleted; an emptied account is
// cleared with a filtered bulk delete; a sync with the restricted view on
// is refused; and the owner's own account and activity are never touched.
func TestGhostfolioContract(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	hc := contractClient(t)
	c := ghostfolio.NewClient(hc, recordedBaseURL, securityToken)
	own := addOwnActivity(t, c)
	record := syntheticRecord()
	s := ghostfolio.NewSyncer(record, c, instance, nil)
	runSteps(t, s, []syncStep{
		{name: "first sync", want: ghostfolio.Report{AccountsCreated: 2, Imported: 3}},
		{name: "nothing changed", want: ghostfolio.Report{}},
		{name: "a manual edit is reverted", before: func(t *testing.T) {
			t.Helper()
			editBuy(t, hc, c)
		}, want: ghostfolio.Report{Imported: 1, Deleted: 1}},
		{name: "nothing changed after the revert", want: ghostfolio.Report{}},
		{name: "a removed transaction is deleted", before: func(*testing.T) { record.withoutBond("fee-1") }, want: ghostfolio.Report{Deleted: 1}},
		{name: "an emptied account is cleared", before: func(*testing.T) {
			record.withoutBond("open-1")
			record.withoutBond("buy-1")
		}, want: ghostfolio.Report{AccountsCleared: 1}},
	})
	setRestrictedView(t, hc, true)
	if _, err := s.Sync(ctx); !errors.Is(err, ghostfolio.ErrRestrictedView) {
		t.Errorf("Sync with the restricted view on = %v, want ErrRestrictedView", err)
	}
	setRestrictedView(t, hc, false)
	acts, err := c.Activities(ctx, []string{own.AccountID})
	if err != nil || len(acts) != 1 || acts[0].Comment != own.Comment || !acts[0].Fee.Equal(own.Fee) {
		t.Errorf("the owner's own activities = %+v, %v; want the one fee of their own, unchanged", acts, err)
	}
}

// addOwnActivity logs c in and gives the new user's own account an
// activity of the owner's own, which no sync may touch, and returns it.
func addOwnActivity(t *testing.T, c *ghostfolio.Client) ghostfolio.Activity {
	t.Helper()
	ctx := t.Context()
	if err := c.Login(ctx); err != nil {
		t.Fatalf("Login: %v", err)
	}
	ids, err := c.Accounts(ctx)
	if err != nil || len(ids) != 1 {
		t.Fatalf("Accounts = %v, %v; want the new user's own account", ids, err)
	}
	own := activity(t, "own", "1", ghostfolio.TypeFee, "2025-01-02", "PLN", ghostfolio.DataSourceManual, ghostfolio.FeeSymbol(cur(t, "PLN")), "0", "0", "3")
	own.AccountID, own.Comment = ids[0], "a fee of my own"
	feeAsset := ghostfolio.AssetProfile{Symbol: own.Symbol, Name: "OpenFief fees (PLN)", Currency: cur(t, "PLN")}
	if _, err := c.Import(ctx, &ghostfolio.Batch{Activities: []ghostfolio.Activity{own}, Assets: []ghostfolio.AssetProfile{feeAsset}}, false); err != nil {
		t.Fatalf("importing the owner's own activity: %v", err)
	}
	return own
}

// syncStep is one sync of a scenario: what changes before it, and what it
// should report.
type syncStep struct {
	name   string
	before func(t *testing.T)
	want   ghostfolio.Report
}

// runSteps runs each step's change, then a sync, and fails t when a sync
// fails or reports anything else.
func runSteps(t *testing.T, s *ghostfolio.Syncer, steps []syncStep) {
	t.Helper()
	for _, st := range steps {
		if st.before != nil {
			st.before(t)
		}
		got, err := s.Sync(t.Context())
		if err != nil {
			t.Fatalf("%s: Sync: %v", st.name, err)
		}
		if diff := cmp.Diff(st.want, got); diff != "" {
			t.Fatalf("%s: report (-want +got):\n%s", st.name, diff)
		}
	}
}

// editBuy edits the synthetic buy by hand, as an owner would in
// Ghostfolio's UI: one more unit, the comment kept.
func editBuy(t *testing.T, hc *http.Client, c *ghostfolio.Client) {
	t.Helper()
	owned := ghostfolio.AccountUUID(instance, "acct-bonds")
	acts, err := c.Activities(t.Context(), []string{owned})
	if err != nil {
		t.Fatalf("Activities: %v", err)
	}
	for i := range acts {
		a := &acts[i]
		if a.Comment != comment("bonds", "buy-1") {
			continue
		}
		body := map[string]any{
			"accountId": a.AccountID, "comment": a.Comment, "currency": a.Currency.String(), "dataSource": a.DataSource,
			"date": a.Date.String() + "T00:00:00.000Z", "fee": json.Number(a.Fee.String()), "id": a.ID,
			"quantity": json.Number(a.Quantity.Add(dec(t, "1")).String()), "symbol": a.Symbol, "type": string(a.Type),
			"unitPrice": json.Number(a.UnitPrice.String()),
		}
		send(t, hc, http.MethodPut, "/api/v1/activities/"+a.ID, body)
		return
	}
	t.Fatal("the synthetic buy is not in Ghostfolio")
}

// setRestrictedView turns the user's restricted view setting on or off.
func setRestrictedView(t *testing.T, hc *http.Client, on bool) {
	t.Helper()
	send(t, hc, http.MethodPut, "/api/v1/user/setting", map[string]any{"isRestrictedView": on})
}

// send makes one request of the owner's own, outside the view's client,
// and fails t unless it succeeds.
func send(t *testing.T, hc *http.Client, method, path string, body any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, recordedBaseURL+path, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sanitized)
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode/100 != 2 {
		t.Fatalf("%s %s: HTTP %d", method, path, resp.StatusCode)
	}
}
