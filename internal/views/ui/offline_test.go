package ui

import (
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// offlineAttr finds the overview the page hands the installed app.
var offlineAttr = regexp.MustCompile(`data-offline-overview="([^"]*)"`)

// keptOverview returns the overview a page carries for offline reading,
// decoded, and whether it carries one.
func keptOverview(t *testing.T, body string) (offlineOverview, bool) {
	t.Helper()
	m := offlineAttr.FindStringSubmatch(body)
	if m == nil {
		return offlineOverview{}, false
	}
	var o offlineOverview
	if err := json.Unmarshal([]byte(html.UnescapeString(m[1])), &o); err != nil {
		t.Fatalf("kept overview %q: %v", m[1], err)
	}
	return o, true
}

// TestBuildOfflineOverview checks what of an overview the installed app
// keeps (D-151): the total's three values, each account's names and three
// values, the allocation's shares, the date and the session's end, written
// in the page's number format; the allocation's total stands in for a
// missing total, an account valued on no day is marked so, and a sleeve
// is written by its name, its ID without one.
func TestBuildOfflineOverview(t *testing.T) {
	t.Parallel()
	comma := separators(" ", ",")
	account := &viewsv1.Account{Id: accountID, Name: "Synthetic brokerage", Institution: "synthetic-broker", Sleeve: "core"}
	alloc := &allocationView{
		Dimension: dimension{Label: "Asset class"}, Value: valueKind{Label: "Gross"}, Total: pln("1500.00"),
		Groups: []allocationRow{{Key: "listed", Value: pln("1000.00"), Share: "66.7%"}, {Key: "cash", Value: pln("500.00"), Share: "33.3%"}},
	}
	shares := offlineAllocation{By: "Asset class", Value: "Gross", Shares: []offlineShare{{"listed", "66.7%"}, {"cash", "33.3%"}}}
	ends := time.Date(2030, 1, 2, 3, 4, 5, 0, time.FixedZone("CET", 3600))
	for _, tc := range []struct {
		name string
		data overviewData
		ends time.Time
		want offlineOverview
	}{
		{
			name: "total and a valued account",
			data: overviewData{
				Date:     "2026-03-31",
				Total:    &viewsv1.DateTotal{ValueReporting: pln("1500.00"), NetReporting: pln("1400.00"), LiquidationReporting: pln("1300.5")},
				Accounts: []accountRow{{Account: account, Total: &viewsv1.DateTotal{ValueReporting: pln("1000.00"), NetReporting: pln("950.00")}, Flagged: 2}},
				Sleeves:  []sleeveGroup{{ID: "core", Name: "Core sleeve"}},
				Over:     &overTimeView{Allocation: alloc},
			},
			ends: ends,
			want: offlineOverview{
				Version: offlineVersion, Date: "2026-03-31", SessionEnds: "2030-01-02T02:04:05Z",
				Total: offlineValues{Gross: "1 500,00 PLN", Net: "1 400,00 PLN", Liquidation: "1 300,5 PLN"},
				Accounts: []offlineAccount{{
					Name: "Synthetic brokerage", Institution: "synthetic-broker", Sleeve: "Core sleeve", Valued: true,
					Values: offlineValues{Gross: "1 000,00 PLN", Net: "950,00 PLN"},
				}},
				Allocation: shares,
			},
		},
		{
			name: "no total, an account not valued, no session end",
			data: overviewData{Date: "2026-03-31", Accounts: []accountRow{{Account: account}}, Over: &overTimeView{Allocation: alloc}},
			want: offlineOverview{
				Version: offlineVersion, Date: "2026-03-31", Total: offlineValues{Gross: "1 500,00 PLN"},
				Accounts:   []offlineAccount{{Name: "Synthetic brokerage", Institution: "synthetic-broker", Sleeve: "core"}},
				Allocation: shares,
			},
		},
		{
			name: "nothing yet",
			data: overviewData{Date: "2026-03-31"},
			want: offlineOverview{
				Version: offlineVersion, Date: "2026-03-31", Accounts: []offlineAccount{},
				Allocation: offlineAllocation{Shares: []offlineShare{}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildOfflineOverview(&tc.data, tc.ends, comma)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("kept overview (-want +got):\n%s", diff)
			}
			encoded := got.encode()
			if strings.Contains(encoded, accountID) {
				t.Errorf("kept overview names the account's ID: %s", encoded)
			}
		})
	}
}

// TestOfflineOverviewKeys checks that the kept overview's JSON holds only
// the members D-151 allows: nothing that names an account, a transaction,
// a lot, a holding or an amount of an allocation group.
func TestOfflineOverviewKeys(t *testing.T) {
	t.Parallel()
	o := buildOfflineOverview(&overviewData{
		Date: "2026-03-31", Total: &viewsv1.DateTotal{ValueReporting: pln("1")},
		Accounts: []accountRow{{Account: &viewsv1.Account{Id: accountID, Name: "n"}, Total: &viewsv1.DateTotal{ValueReporting: pln("1")}}},
		Over:     &overTimeView{Allocation: &allocationView{Total: pln("1"), Groups: []allocationRow{{Key: "k", Value: pln("1"), Share: "100.0%"}}}},
	}, time.Time{}, plainFormat())
	var tree any
	if err := json.Unmarshal([]byte(o.encode()), &tree); err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				keys[k] = true
				walk(child)
			}
		case []any:
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(tree)
	want := map[string]bool{
		"v": true, "date": true, "sessionEnds": true, "total": true, "gross": true, "net": true, "liquidation": true,
		"accounts": true, "name": true, "institution": true, "sleeve": true, "valued": true, "values": true,
		"allocation": true, "by": true, "value": true, "shares": true, "key": true, "share": true,
	}
	if diff := cmp.Diff(want, keys); diff != "" {
		t.Errorf("kept overview's members (-want +got):\n%s", diff)
	}
}

// TestOverviewHandsOverOnlyTheOverview checks the overview page carries
// the copy to keep, with the session's end, and that no other page does:
// the account page with its transactions and lots, the accounts page with
// broker accounts, and the others never hand anything to keep. Every one
// is no-store.
func TestOverviewHandsOverOnlyTheOverview(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	o, ok := keptOverview(t, rec.Body.String())
	if !ok {
		t.Fatal("the overview hands the app nothing to keep")
	}
	got := []string{o.Date, o.SessionEnds, o.Total.Gross, o.Total.Net, o.Accounts[0].Name, o.Accounts[0].Values.Gross, o.Allocation.Shares[0].Share}
	want := []string{"2026-03-31", "2030-01-02T00:00:00Z", "1 500,00 PLN", "1 400,00 PLN", "Synthetic brokerage", "1 000,00 PLN", "66.7%"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("kept overview (-want +got):\n%s", diff)
	}
	if m := offlineAttr.FindStringSubmatch(rec.Body.String()); strings.Contains(m[1], accountID) {
		t.Errorf("kept overview names the account's ID: %s", m[1])
	}
	for _, path := range []string{"/", "/accounts", "/accounts/" + accountID, "/accounts/" + accountID + "?tab=lots", "/instruments/XX0000000001", "/holdings?by=currency&key=EUR", "/inbox", "/sleeves", "/plugins", "/settings"} {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", path, got)
		}
		if _, kept := keptOverview(t, rec.Body.String()); kept != (path == "/") {
			t.Errorf("%s: hands the app something to keep = %v", path, kept)
		}
	}
	frag := h.serve(t, htmx(withSession(request(http.MethodGet, "/allocation?by=asset_class", ""), ownerSession)))
	if _, kept := keptOverview(t, frag.Body.String()); kept {
		t.Error("the allocation fragment hands the app something to keep")
	}
}

// TestSessionEndWipesTheKeptOverview checks the server's side of wiping
// the kept overview (D-151): sign-out and a session the core rejects
// answer with Clear-Site-Data "storage", which deletes IndexedDB, and the
// login page, shown only without a session, marks itself for offline.js
// to wipe; a page with a session does neither.
func TestSessionEndWipesTheKeptOverview(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, tc := range []struct {
		name  string
		req   *http.Request
		clear string
		wipe  bool
	}{
		{name: "sign-out", req: withSession(request(http.MethodPost, "/logout", ""), ownerSession), clear: `"storage"`},
		{name: "expired or revoked session", req: withSession(request(http.MethodGet, "/", ""), "session-expired"), clear: `"storage"`},
		{name: "expired session, htmx", req: htmx(withSession(request(http.MethodGet, "/allocation", ""), "session-expired")), clear: `"storage"`},
		{name: "login page", req: request(http.MethodGet, "/login", ""), wipe: true},
		{name: "overview", req: withSession(request(http.MethodGet, "/", ""), ownerSession)},
	} {
		rec := h.serve(t, tc.req)
		if got := rec.Header().Get("Clear-Site-Data"); got != tc.clear {
			t.Errorf("%s: Clear-Site-Data = %q, want %q", tc.name, got, tc.clear)
		}
		if got := strings.Contains(rec.Body.String(), "data-offline-wipe"); got != tc.wipe {
			t.Errorf("%s: marked for wiping = %v, want %v", tc.name, got, tc.wipe)
		}
	}
}

// TestSessionEnd checks when a session ends: the earlier of its idle and
// absolute expiry, either alone, or never named.
func TestSessionEnd(t *testing.T) {
	t.Parallel()
	idle := time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC)
	absolute := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		s    *identityv1.Session
		want time.Time
	}{
		{name: "idle first", s: &identityv1.Session{IdleExpiresAt: timestamppb.New(idle), ExpiresAt: timestamppb.New(absolute)}, want: idle},
		{name: "absolute first", s: &identityv1.Session{IdleExpiresAt: timestamppb.New(absolute), ExpiresAt: timestamppb.New(idle)}, want: idle},
		{name: "absolute only", s: &identityv1.Session{ExpiresAt: timestamppb.New(absolute)}, want: absolute},
		{name: "idle only", s: &identityv1.Session{IdleExpiresAt: timestamppb.New(idle)}, want: idle},
		{name: "neither", s: &identityv1.Session{}},
		{name: "no session"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sessionEnd(tc.s); !got.Equal(tc.want) {
				t.Errorf("sessionEnd = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestOfflineScripts checks the scripts that keep and wipe the overview
// agree with the UI on the database's name, that every page loads
// offline.js, that the overview is encrypted with AES-GCM under a key
// generated as not extractable and stored as the key object, that no
// script writes the browser's other storage, and that htmx keeps no pages
// in its history cache. The browser checks
// (scripts/ui-browser.mjs) prove the same in Chromium.
func TestOfflineScripts(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if got := workerOf(t, h).cfg.OfflineDB; got != offlineDB {
		t.Errorf("worker's offlineDB = %q, want %q", got, offlineDB)
	}
	rec := h.serve(t, request(http.MethodGet, "/static/offline.js", ""))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "const dbName = '"+offlineDB+"';",
		"generateKey({ name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt'])",
		"put(key, keyEntry)", "deleteDatabase(dbName)")
	wantNoBody(t, rec, "exportKey", "localStorage", "sessionStorage", "caches.", "document.cookie")
	for _, path := range []string{"/static/layout.js", "/static/app.js", "/static/webauthn.js", workerPath} {
		rec := h.serve(t, request(http.MethodGet, path, ""))
		wantNoBody(t, rec, "localStorage", "sessionStorage", "document.cookie")
	}
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	for name, set := range p.sets {
		src := set.Lookup("layout.html").Tree.Root.String()
		if !strings.Contains(src, `<script src="/static/offline.js" defer></script>`) {
			t.Errorf("page %s doesn't load offline.js", name)
		}
		// htmx would otherwise keep snapshots of the pages it pushed into
		// the history, the owner's record among them, in sessionStorage.
		if !strings.Contains(src, `"historyCacheSize":0,"refreshOnHistoryMiss":true`) {
			t.Errorf("page %s lets htmx keep pages in its history cache", name)
		}
	}
	off := h.serve(t, request(http.MethodGet, offlinePath, ""))
	wantBody(t, off, `<div id="offline-copy" hidden>`, "data-offline-accounts", "data-offline-shares", "Offline copy.")
}
