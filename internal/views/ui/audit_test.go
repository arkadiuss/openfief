package ui

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// auditRows returns the text of the "Who" and "What" cells of each row of the
// first audit table in body, as "who | what".
func auditRows(body string) []string {
	cell := regexp.MustCompile(`(?s)<tr>\s*<td class="mono">[^<]*</td><td>([^<]*)</td><td>([^<]*)</td>`)
	matches := cell.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1]+" | "+m[2])
	}
	return out
}

// TestAccessHistory checks the History section of the owner's Access page
// (D-210): every grant and revocation and every link, newest first, with the
// who, what and outcome and the records by name, a record that is gone by its
// ID, none of the sign-ins, and a link to the whole audit log; and that the
// page asks the core for those actions only.
func TestAccessHistory(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`id="history-title"`, "Invitation cancelled", "Invitation for Dana",
		"Access granted", "Alice → Sleeve Core", "Script token backup-script → Account acc-deleted-1 (removed)",
		"Access ended", "View plugin ui-local → Whole household", "Viewer created", "Enrollment link issued",
		"The core (command line)", `href="/settings/audit"`, "See the whole audit log",
	)
	want := []string{
		"Owner | Invitation cancelled", "Enrollment link 99999999 | Enrollment link used", "Owner | Access ended", "Owner | Access granted",
		"Enrollment link 99999999 | Access granted", "Enrollment link 99999999 | Viewer created", "Enrollment link 99999999 | Enrollment link used",
		"Owner | Enrollment link issued", "The core (command line) | Enrollment link issued",
	}
	if diff := cmp.Diff(want, auditRows(rec.Body.String())[:len(want)]); diff != "" {
		t.Errorf("history rows, newest first (-want +got):\n%s", diff)
	}
	if strings.Contains(rec.Body.String(), "Sign-in") && strings.Count(rec.Body.String(), "<td>Sign-in</td>") > 0 {
		t.Error("the history lists a sign-in")
	}
	var req *identityv1.ListAuditEntriesRequest
	h.auth.locked(func() { req = h.auth.auditReqs[len(h.auth.auditReqs)-1] })
	if diff := cmp.Diff(historyActions(), req.GetActions()); diff != "" || req.GetPageSize() != historyPageSize || req.GetSessionToken() != ownerSession {
		t.Errorf("ListAuditEntries request = %v, want the history actions in a page of %d (%s)", req, historyPageSize, diff)
	}
}

// TestAccessHistoryPages checks that the history is paged: a page that is full
// links to the older entries with the cursor the core handed out, which the next
// request carries, and the last page has no such link.
func TestAccessHistoryPages(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.auditCap = 4 })
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantBody(t, rec, `href="/access?history_before=`)
	next := regexp.MustCompile(`href="/access\?history_before=(\d+)#history-title"`).FindStringSubmatch(rec.Body.String())
	if next == nil {
		t.Fatal("a full page of the history has no link to older entries")
	}
	if got := len(auditRows(rec.Body.String())); got != 4 {
		t.Errorf("the first page has %d rows, want 4", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/access?history_before="+next[1], ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	var req *identityv1.ListAuditEntriesRequest
	h.auth.locked(func() { req = h.auth.auditReqs[len(h.auth.auditReqs)-1] })
	if req.GetPageToken() != next[1] {
		t.Errorf("page token = %q, want %q", req.GetPageToken(), next[1])
	}
	wantBody(t, rec, "Enrollment link issued")
	wantNoBody(t, rec, "Invitation cancelled")
	// A cursor that is no number is left out rather than sent to the core.
	rec = h.serve(t, withSession(request(http.MethodGet, "/access?history_before=abc", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	h.auth.locked(func() {
		if got := h.auth.auditReqs[len(h.auth.auditReqs)-1].GetPageToken(); got != "" {
			t.Errorf("page token = %q for a cursor that is no number, want none", got)
		}
	})
}

// TestAuditLogPage checks Settings' Audit log (D-211): every entry, newest
// first, the actors by name (a viewer by their name, a link or a session that
// is gone by its ID), the records by name or by ID when gone, the outcomes, and
// the navigation link.
func TestAuditLogPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, auditPath, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"<title>Audit log", `aria-current="page">Audit log</a>`, "Alice", "Nobody signed in", "Sleeve Core", "Passkey cGFzc2tl",
		"Whole household", "Session 0123456", "Link 99999999", "refused", `class="pill pill-ok">ok`, `class="pill pill-warn">refused`,
		"Sign-in",
	)
	rows := auditRows(rec.Body.String())
	if len(rows) != 12 || rows[0] != "Owner | Invitation cancelled" || rows[11] != "The core (command line) | Enrollment link issued" {
		t.Errorf("rows = %q, want the 12 entries newest first", rows)
	}
	wantNoBody(t, rec, "Older entries")
}

// TestAuditLogNamesEveryValueOfTheContract checks that the page has words for
// every action, kind of actor and kind of record the contract has (D-211), the
// owner's changes in every context and the operator's installs included, so a
// value added to the contract without its words fails here.
func TestAuditLogNamesEveryValueOfTheContract(t *testing.T) {
	t.Parallel()
	for n, name := range identityv1.AuditAction_name {
		if n != 0 && actionWords(identityv1.AuditAction(n)) == "Unknown action" {
			t.Errorf("%s has no words on the Audit log page", name)
		}
	}
	names := &auditNames{accounts: map[string]string{"xtb-main": "XTB main"}}
	for n, name := range identityv1.AuditRefKind_name {
		got := names.ref(&identityv1.AuditRef{Kind: identityv1.AuditRefKind(n), Id: "xtb-main"})
		if n != 0 && strings.HasPrefix(got, "Record ") {
			t.Errorf("%s has no words on the Audit log page: %q", name, got)
		}
	}
}

// TestAuditLogNamesEveryKindOfActor checks that the page has words for every
// kind of actor the contract has, in the filter and in the table.
func TestAuditLogNamesEveryKindOfActor(t *testing.T) {
	t.Parallel()
	for n, name := range identityv1.AuditActorKind_name {
		if n != 0 && actorKindWords(identityv1.AuditActorKind(n)) == unknownWord {
			t.Errorf("%s has no words in the actor filter", name)
		}
		actor := &identityv1.AuditActor{Kind: identityv1.AuditActorKind(n), Id: "0123456789abcdef"}
		if n != 0 && (&auditNames{}).actor(actor) == unknownWord {
			t.Errorf("%s is no one the page can name", name)
		}
	}
}

// TestAuditLogReadsChangesAsSentencesOfIDs checks that an owner's change and
// the operator's install read as who, what and the records by ID or name,
// never a value.
func TestAuditLogReadsChangesAsSentencesOfIDs(t *testing.T) {
	t.Parallel()
	names := &auditNames{accounts: map[string]string{"xtb-main": "XTB main"}}
	operator := &identityv1.AuditEntry{
		Actor:   &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OPERATOR, Id: "operator"},
		Action:  identityv1.AuditAction_AUDIT_ACTION_PLUGIN_UPGRADED,
		Outcome: identityv1.AuditOutcome_AUDIT_OUTCOME_OK,
		Refs: []*identityv1.AuditRef{
			{Kind: identityv1.AuditRefKind_AUDIT_REF_KIND_PLUGIN, Id: "xtb-main"},
			{Kind: identityv1.AuditRefKind_AUDIT_REF_KIND_PLUGIN_GENERATION, Id: "4"},
		},
	}
	if row := names.row(operator); row.Who != "The operator" || row.What != "Plugin upgraded" || row.Records != "Plugin xtb-main, Generation 4" {
		t.Errorf("the operator's entry reads %+v", row)
	}
	change := &identityv1.AuditEntry{
		Actor:   &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER, Id: "0123456789abcdef"},
		Action:  identityv1.AuditAction_AUDIT_ACTION_SETTINGS_CHANGED,
		Outcome: identityv1.AuditOutcome_AUDIT_OUTCOME_OK,
		Refs: []*identityv1.AuditRef{
			{Kind: identityv1.AuditRefKind_AUDIT_REF_KIND_SETTING, Id: "chart:start"},
			{Kind: identityv1.AuditRefKind_AUDIT_REF_KIND_ACCOUNT, Id: "xtb-main"},
		},
	}
	if row := names.row(change); row.Who != "Owner" || row.What != "Setting changed" || row.Records != "Setting chart:start, Account XTB main" {
		t.Errorf("the owner's change reads %+v", row)
	}
}

// TestAuditLogFilters checks the filters of D-211 reach the core as the owner
// set them, by actor kind, action, outcome and range of days (the last day
// included), alone and together, and show the matching entries; and that a
// value that is none of the choices never reaches the core.
func TestAuditLogFilters(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	tests := []struct {
		name, query string
		want        []string
	}{
		{"actor kind", "actor=viewer", []string{"Alice | Sign-in"}},
		{"action", "action=grant_created", []string{"Owner | Access granted", "Enrollment link 99999999 | Access granted"}},
		{"outcome", "outcome=refused", []string{"Enrollment link 99999999 | Enrollment link used", "Nobody signed in | Sign-in"}},
		{"outcome and actor kind", "outcome=refused&actor=anonymous", []string{"Nobody signed in | Sign-in"}},
		{"one day", "from=2026-03-02&to=2026-03-02", []string{"Owner | Access ended", "Owner | Access granted", "Alice | Sign-in"}},
		{"from a day", "from=2026-03-03", []string{"Owner | Invitation cancelled", "Enrollment link 99999999 | Enrollment link used", "Nobody signed in | Sign-in"}},
		{"to a day", "to=2026-03-01&action=person_created", []string{"Enrollment link 99999999 | Viewer created"}},
		{"nothing matches", "action=model_call", []string{}},
	}
	for _, tt := range tests {
		rec := h.serve(t, withSession(request(http.MethodGet, auditPath+"?"+tt.query, ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		if diff := cmp.Diff(tt.want, auditRows(rec.Body.String())); diff != "" {
			t.Errorf("%s: rows (-want +got):\n%s", tt.name, diff)
		}
		wantBody(t, rec, "Clear the filter")
	}
	rec := h.serve(t, withSession(request(http.MethodGet, auditPath+"?actor=owner&from=2026-03-02&to=2026-03-02&outcome=ok&action=grant_revoked", ""), ownerSession))
	wantBody(t, rec, `<option value="owner" selected>`, `<option value="grant_revoked" selected>`, `<option value="ok" selected>`, `value="2026-03-02"`)
	var req *identityv1.ListAuditEntriesRequest
	h.auth.locked(func() { req = h.auth.auditReqs[len(h.auth.auditReqs)-1] })
	if req.GetActorKind() != identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER || req.GetAction() != identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED ||
		req.GetOutcome() != identityv1.AuditOutcome_AUDIT_OUTCOME_OK || req.GetFrom().AsTime().Day() != 2 || req.GetTo().AsTime().Day() != 3 ||
		req.GetPageSize() != auditPageSize || req.GetSessionToken() != ownerSession {
		t.Errorf("ListAuditEntries request = %v", req)
	}
}

// TestAuditLogMalformedFilters checks that a filter value that is none of the
// choices, a date out of order and a cursor that is no number are a 400 that
// never reaches the core.
func TestAuditLogMalformedFilters(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, q := range []string{
		"actor=admin", "action=delete_everything", "outcome=maybe", "from=yesterday", "to=2026-13-45", "from=2026-03-03&to=2026-03-01", "before=abc", "before=-3",
	} {
		rec := h.serve(t, withSession(request(http.MethodGet, auditPath+"?"+q, ""), ownerSession))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("GET ?%s = %d, want 400", q, rec.Code)
		}
	}
	h.auth.locked(func() {
		if len(h.auth.auditReqs) != 0 {
			t.Errorf("a malformed filter reached the core: %v", h.auth.auditReqs)
		}
	})
}

// TestAuditLogPages checks that the log is paged: a full page links to the older
// entries keeping the filter, the link's cursor reaches the core, and the last
// page has no link.
func TestAuditLogPages(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.auditCap = 5 })
	var seen []string
	path := auditPath + "?outcome=ok"
	link := regexp.MustCompile(`href="([^"]*)" rel="next"`)
	for range 5 {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		seen = append(seen, auditRows(rec.Body.String())...)
		m := link.FindStringSubmatch(rec.Body.String())
		if m == nil {
			break
		}
		next := strings.ReplaceAll(m[1], "&amp;", "&")
		if !strings.Contains(next, "outcome=ok") || !strings.Contains(next, "before=") {
			t.Fatalf("the link to older entries is %q, want the filter and a cursor", next)
		}
		path = next
	}
	if len(seen) != 10 || seen[0] != "Owner | Invitation cancelled" || seen[9] != "The core (command line) | Enrollment link issued" {
		t.Errorf("paging saw %d rows %q, want the ten entries that ended ok, once each, newest first", len(seen), seen)
	}
}

// TestAuditLogUnavailable checks that a core that can't answer is said so on the
// log and in the history, not shown as an empty log, and fails neither page.
func TestAuditLogUnavailable(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.auditErr = codeErr(connect.CodeUnavailable, errPluginsUnavailable) })
	rec := h.serve(t, withSession(request(http.MethodGet, auditPath, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The audit log can't be read now")
	wantNoBody(t, rec, "Nothing has been recorded yet")
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The history can't be read now", "People")
	wantNoBody(t, rec, "Nothing has been granted or revoked yet")
}

// TestAuditAndHistoryAreTheOwners checks that a viewer, a read session and the
// demo's session get 403 on the Audit log and on the Access page's history, with
// no entry shown and no call to the core for the log, that the settings
// navigation they do see has no link to it, and that a signed-out browser goes
// to the login page. The demo's session is read-only, so the demo, like a
// viewer, has no audit log.
func TestAuditAndHistoryAreTheOwners(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, session := range []string{viewerSession, emptyViewerSession, readSession, demoSession} {
		for _, path := range []string{auditPath, "/access", auditPath + "?actor=owner"} {
			rec := h.serve(t, withSession(request(http.MethodGet, path, ""), session))
			wantStatus(t, rec, http.StatusForbidden)
			wantNoBody(t, rec, "Access granted", "Sign-in", "Alice", "Enrollment link")
		}
		for _, path := range []string{"/sleeves", "/plugins"} {
			rec := h.serve(t, withSession(request(http.MethodGet, path, ""), session))
			wantNoBody(t, rec, "/settings/audit", "Audit log")
		}
	}
	rec := h.serve(t, request(http.MethodGet, auditPath, ""))
	if loc := rec.Header().Get("Location"); rec.Code != http.StatusSeeOther || !strings.HasPrefix(loc, "/login") {
		t.Errorf("GET %s without a session = %d to %q, want the login page", auditPath, rec.Code, loc)
	}
	h.auth.locked(func() {
		if len(h.auth.auditReqs) != 0 {
			t.Errorf("a refused page reached the core's audit log: %v", h.auth.auditReqs)
		}
	})
}
