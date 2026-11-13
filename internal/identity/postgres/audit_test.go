package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// Synthetic IDs of the audit tests.
const (
	auditSession = "0123456789abcdef"
	auditOther   = "fedcba9876543210"
	auditTarget  = "00112233445566ff"
)

// sqlState returns the SQLSTATE of err, or "" when it is no database error.
func sqlState(err error) string {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code
	}
	return ""
}

func signIn(at time.Time, session string) identity.AuditEntry {
	return identity.AuditEntry{
		At: at, Actor: identity.Actor{Kind: identity.ActorOwner, ID: session}, Action: identity.ActionSignIn, Outcome: identity.OutcomeOK,
		Refs: []identity.Ref{{Kind: identity.RefSession, ID: session}},
	}
}

// TestAuditLogIsAppendOnly checks that neither layer of D-211's append-only
// rule lets an entry change: the core's role holds no UPDATE, DELETE or
// TRUNCATE on the table, and a trigger refuses them even to a role that does.
func TestAuditLogIsAppendOnly(t *testing.T) {
	t.Parallel()
	statements := map[string]string{
		"UPDATE":                "UPDATE identity.audit_log SET outcome = 'refused'",
		"UPDATE of no row":      "UPDATE identity.audit_log SET outcome = 'refused' WHERE false",
		"DELETE":                "DELETE FROM identity.audit_log",
		"DELETE of no row":      "DELETE FROM identity.audit_log WHERE false",
		"TRUNCATE":              "TRUNCATE identity.audit_log",
		"UPDATE of its own row": "UPDATE identity.audit_log SET at = now() WHERE id = 1",
	}
	for _, tc := range []struct {
		name     string
		appRole  bool
		wantCode string
	}{
		{name: "as the app role", appRole: true, wantCode: "42501"}, // insufficient_privilege
		{name: "as a superuser", appRole: false, wantCode: "23001"}, // restrict_violation, the trigger
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pool := newPool(t, tc.appRole)
			store := postgres.New(pool)
			ctx := t.Context()
			e := signIn(issuedAt(), auditSession)
			if err := store.AppendAudit(ctx, &e); err != nil {
				t.Fatalf("AppendAudit: %v", err)
			}
			for name, sql := range statements {
				if _, err := pool.Exec(ctx, sql); sqlState(err) != tc.wantCode {
					t.Errorf("%s: error = %v, want SQLSTATE %s", name, err, tc.wantCode)
				}
			}
			page, err := store.AuditEntries(ctx, &identity.AuditFilter{})
			if err != nil || len(page.Entries) != 1 || page.Entries[0].Outcome != identity.OutcomeOK {
				t.Errorf("AuditEntries = %+v, %v; want the one entry untouched", page, err)
			}
		})
	}
}

// TestAuditLogKeepsModelCalls checks that the model router's two entries per
// call (D-205, D-211) go in and come back: the role, the model and the call
// ID by name, the outcome sent and then the result, an error class for a
// failure, and the system actor, with no text of a request. A refusal before a
// model was chosen is a result entry alone.
func TestAuditLogKeepsModelCalls(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t, true))
	ctx := t.Context()
	system := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	role := identity.Ref{Kind: identity.RefModelRole, ID: "ask"}
	model := identity.Ref{Kind: identity.RefModel, ID: "openrouter/openai/gpt-oss-120b"}
	call := identity.Ref{Kind: identity.RefModelCall, ID: "00112233445566aa"}
	entries := []identity.AuditEntry{
		{At: issuedAt(), Actor: system, Action: identity.ActionModelCall, Outcome: identity.OutcomeSent, Refs: []identity.Ref{role, model, call}},
		{
			At: issuedAt(), Actor: system, Action: identity.ActionModelCallResult, Outcome: identity.OutcomeFailed,
			Refs: []identity.Ref{role, model, call, {Kind: identity.RefModelError, ID: "unavailable"}},
		},
		{At: issuedAt(), Actor: system, Action: identity.ActionModelCallResult, Outcome: identity.OutcomeRefused, Refs: []identity.Ref{role, {Kind: identity.RefModelCall, ID: "00112233445566bb"}}},
	}
	for i := range entries {
		if err := store.AppendAudit(ctx, &entries[i]); err != nil {
			t.Fatalf("AppendAudit(%s): %v", entries[i].Action, err)
		}
	}
	for _, action := range []identity.AuditAction{identity.ActionModelCall, identity.ActionModelCallResult} {
		page, err := store.AuditEntries(ctx, &identity.AuditFilter{Action: action})
		if err != nil {
			t.Fatalf("AuditEntries(%s): %v", action, err)
		}
		var want []identity.AuditEntry
		for _, e := range slices.Backward(entries) { // newest first
			if e.Action == action {
				want = append(want, e)
			}
		}
		if diff := cmp.Diff(want, page.Entries); diff != "" {
			t.Errorf("%s entries (-want +got):\n%s", action, diff)
		}
	}
}

// TestAuditLogRefusesFreeText checks the database's own CHECKs: an insert
// that bypasses the domain still can't put free text, a secret or a long
// payload in an entry.
func TestAuditLogRefusesFreeText(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	const insert = `INSERT INTO identity.audit_log (at, actor_kind, actor_id, action, outcome, refs) VALUES (now(), $1, $2, $3, $4, $5::jsonb)`
	tests := []struct {
		name                            string
		kind, id, action, outcome, refs string
		wantErr                         bool
	}{
		{name: "valid control", kind: "owner", id: auditSession, action: "sign_in", outcome: "ok", refs: `[]`},
		{name: "a model call before it is sent", kind: "owner", id: auditSession, action: "model_call", outcome: "sent", refs: `[]`},
		{name: "a model call's result", kind: "owner", id: auditSession, action: "model_call_result", outcome: "failed", refs: `[]`},
		{name: "free-text action", kind: "owner", id: auditSession, action: "transfer 100 PLN", outcome: "ok", refs: `[]`, wantErr: true},
		{name: "free-text outcome", kind: "owner", id: auditSession, action: "sign_in", outcome: "ok, mostly", refs: `[]`, wantErr: true},
		{name: "owner named by a name", kind: "owner", id: "Jan Kowalski", action: "sign_in", outcome: "ok", refs: `[]`, wantErr: true},
		{name: "unknown actor kind", kind: "guest", id: auditSession, action: "sign_in", outcome: "ok", refs: `[]`, wantErr: true},
		{name: "anonymous with an ID", kind: "anonymous", id: auditSession, action: "sign_in", outcome: "refused", refs: `[]`, wantErr: true},
		{name: "refs not an array", kind: "owner", id: auditSession, action: "sign_in", outcome: "ok", refs: `{"note": "my question"}`, wantErr: true},
		{
			name: "too many refs", kind: "owner", id: auditSession, action: "sign_in", outcome: "ok", wantErr: true,
			refs: "[" + strings.Repeat(`{"kind":"session","id":"0123456789abcdef"},`, 8) + `{}]`,
		},
		{name: "huge refs", kind: "owner", id: auditSession, action: "sign_in", outcome: "ok", refs: `["` + strings.Repeat("x", 13000) + `"]`, wantErr: true},
	}
	for _, tc := range tests {
		_, err := pool.Exec(t.Context(), insert, tc.kind, tc.id, tc.action, tc.outcome, tc.refs)
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: error = %v, want an error: %v", tc.name, err, tc.wantErr)
		}
	}
}

func TestAppendAuditRefusesAnInvalidEntry(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	bad := signIn(issuedAt(), auditSession)
	bad.Refs[0].ID = "my token"
	if err := store.AppendAudit(t.Context(), &bad); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Fatalf("AppendAudit of a bad reference = %v, want %v", err, identity.ErrInvalidAuditEntry)
	}
	if page, err := store.AuditEntries(t.Context(), &identity.AuditFilter{}); err != nil || len(page.Entries) != 0 {
		t.Errorf("AuditEntries = %+v, %v; want none", page, err)
	}
}

// seededLog is a store holding five entries, oldest first, an hour apart from
// base.
type seededLog struct {
	store   *postgres.Store
	entries []identity.AuditEntry
	base    time.Time
}

func newSeededLog(t *testing.T) seededLog {
	t.Helper()
	store := postgres.New(newPool(t))
	base := issuedAt()
	entries := []identity.AuditEntry{
		signIn(base, auditSession),
		{
			At: base.Add(1 * time.Hour), Actor: identity.Actor{Kind: identity.ActorOwner, ID: auditSession}, Action: identity.ActionScriptTokenCreated,
			Outcome: identity.OutcomeOK, Refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: auditTarget}},
		},
		{At: base.Add(2 * time.Hour), Actor: identity.Actor{Kind: identity.ActorAnonymous}, Action: identity.ActionSignIn, Outcome: identity.OutcomeRefused},
		signIn(base.Add(3*time.Hour), auditOther),
		{
			At: base.Add(4 * time.Hour), Actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}, Action: identity.ActionEnrollmentLinkIssued,
			Outcome: identity.OutcomeOK, Refs: []identity.Ref{{Kind: identity.RefEnrollmentCode, ID: auditTarget}},
		},
	}
	for i := range entries {
		if err := store.AppendAudit(t.Context(), &entries[i]); err != nil {
			t.Fatalf("AppendAudit %d: %v", i, err)
		}
	}
	return seededLog{store: store, entries: entries, base: base}
}

// index returns the position in l.entries of the entry with sequence seq.
func (l seededLog) index(seq int64) int {
	return slices.IndexFunc(l.entries, func(e identity.AuditEntry) bool { return e.Seq == seq })
}

func TestAuditEntriesFilter(t *testing.T) {
	t.Parallel()
	l := newSeededLog(t)
	base := l.base
	tests := []struct {
		name string
		f    identity.AuditFilter
		want []int // indexes of entries, newest first
	}{
		{name: "everything, newest first", want: []int{4, 3, 2, 1, 0}},
		{name: "by actor kind", f: identity.AuditFilter{ActorKind: identity.ActorOwner}, want: []int{3, 1, 0}},
		{name: "by actor", f: identity.AuditFilter{ActorKind: identity.ActorOwner, ActorID: auditSession}, want: []int{1, 0}},
		{name: "by action", f: identity.AuditFilter{Action: identity.ActionSignIn}, want: []int{3, 2, 0}},
		{name: "by actor and action", f: identity.AuditFilter{ActorKind: identity.ActorOwner, Action: identity.ActionSignIn}, want: []int{3, 0}},
		{name: "from is inclusive", f: identity.AuditFilter{From: base.Add(2 * time.Hour)}, want: []int{4, 3, 2}},
		{name: "to is exclusive", f: identity.AuditFilter{To: base.Add(2 * time.Hour)}, want: []int{1, 0}},
		{name: "a range", f: identity.AuditFilter{From: base.Add(time.Hour), To: base.Add(4 * time.Hour)}, want: []int{3, 2, 1}},
		{name: "nothing matches", f: identity.AuditFilter{ActorKind: identity.ActorEnrollmentCode}, want: []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			page, err := l.store.AuditEntries(t.Context(), &tc.f)
			if err != nil {
				t.Fatalf("AuditEntries: %v", err)
			}
			got := make([]int, 0, len(page.Entries))
			for _, e := range page.Entries {
				got = append(got, l.index(e.Seq))
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
			if page.Next != 0 {
				t.Errorf("Next = %d on a single page", page.Next)
			}
		})
	}
}

// TestAuditEntriesPages checks that pages of two walk the whole log, newest
// first, each entry once.
func TestAuditEntriesPages(t *testing.T) {
	t.Parallel()
	l := newSeededLog(t)
	var seen []int
	f := identity.AuditFilter{Limit: 2}
	for pages := 0; pages <= len(l.entries); pages++ {
		page, err := l.store.AuditEntries(t.Context(), &f)
		if err != nil {
			t.Fatalf("AuditEntries: %v", err)
		}
		for _, e := range page.Entries {
			seen = append(seen, l.index(e.Seq))
		}
		if page.Next == 0 {
			break
		}
		f.Before = page.Next
	}
	if diff := cmp.Diff([]int{4, 3, 2, 1, 0}, seen); diff != "" {
		t.Errorf("paged entries (-want +got):\n%s", diff)
	}
}

// TestAuditEntryRoundTrip checks that an entry reads back as it was appended,
// and that a bad filter is refused.
func TestAuditEntryRoundTrip(t *testing.T) {
	t.Parallel()
	l := newSeededLog(t)
	page, err := l.store.AuditEntries(t.Context(), &identity.AuditFilter{ActorKind: identity.ActorSystem})
	if err != nil || len(page.Entries) != 1 {
		t.Fatalf("AuditEntries = %+v, %v", page, err)
	}
	if diff := cmp.Diff(l.entries[4], page.Entries[0]); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
	if _, err := l.store.AuditEntries(t.Context(), &identity.AuditFilter{Limit: -1}); !errors.Is(err, identity.ErrInvalidAuditFilter) {
		t.Errorf("AuditEntries with a bad filter = %v, want %v", err, identity.ErrInvalidAuditFilter)
	}
}

var errTest = errors.New("test failure")

// txFixture is a database with a script token to create, and what it takes to
// count what a transaction left behind.
type txFixture struct {
	pool  *pgxpool.Pool
	store *postgres.Store
	tok   identity.ScriptToken
}

func newTxFixture(t *testing.T) *txFixture {
	t.Helper()
	pool := newPool(t)
	tok, err := identity.NewScriptToken(identity.ScriptTokenID(auditTarget), "nightly", identity.ScopeRead, issuedAt(), time.Hour)
	if err != nil {
		t.Fatalf("NewScriptToken: %v", err)
	}
	return &txFixture{pool: pool, store: postgres.New(pool), tok: tok}
}

// created saves the token and appends the entry that records it.
func (x *txFixture) created(ctx context.Context) error {
	if err := x.store.SaveScriptToken(ctx, &x.tok, make([]byte, 32)); err != nil {
		return fmt.Errorf("saving the token: %w", err)
	}
	e := identity.AuditEntry{
		At: issuedAt(), Actor: identity.Actor{Kind: identity.ActorOwner, ID: auditSession}, Action: identity.ActionScriptTokenCreated,
		Outcome: identity.OutcomeOK, Refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: auditTarget}},
	}
	if err := x.store.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("appending the entry: %w", err)
	}
	return nil
}

// counts is how many tokens and audit entries the database holds.
type counts struct{ tokens, entries int }

func (x *txFixture) counts(t *testing.T) counts {
	t.Helper()
	var c counts
	err := x.pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM identity.script_tokens), (SELECT count(*) FROM identity.audit_log)`).Scan(&c.tokens, &c.entries)
	if err != nil {
		t.Fatalf("counting: %v", err)
	}
	return c
}

// TestAtomicallyRollsBackStateAndEntry checks that a change of state and its
// audit entry share one transaction: when the function fails after both,
// neither stays.
func TestAtomicallyRollsBackStateAndEntry(t *testing.T) {
	t.Parallel()
	x := newTxFixture(t)
	err := x.store.Atomically(t.Context(), func(ctx context.Context) error {
		if createErr := x.created(ctx); createErr != nil {
			return createErr
		}
		return errTest
	})
	if !errors.Is(err, errTest) {
		t.Fatalf("Atomically = %v, want %v", err, errTest)
	}
	if got := x.counts(t); got != (counts{}) {
		t.Fatalf("after a failed transaction: %+v, want none", got)
	}
}

// TestFailedAuditWriteRollsBackTheChange checks that a change goes when its
// entry is refused.
func TestFailedAuditWriteRollsBackTheChange(t *testing.T) {
	t.Parallel()
	x := newTxFixture(t)
	err := x.store.Atomically(t.Context(), func(ctx context.Context) error {
		if saveErr := x.store.SaveScriptToken(ctx, &x.tok, make([]byte, 32)); saveErr != nil {
			return fmt.Errorf("saving the token: %w", saveErr)
		}
		bad := identity.AuditEntry{At: issuedAt(), Actor: identity.Actor{Kind: identity.ActorOwner, ID: "x"}, Action: identity.ActionSignIn, Outcome: identity.OutcomeOK}
		if appendErr := x.store.AppendAudit(ctx, &bad); appendErr != nil {
			return fmt.Errorf("appending the entry: %w", appendErr)
		}
		return nil
	})
	if !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Fatalf("Atomically with a bad entry = %v, want %v", err, identity.ErrInvalidAuditEntry)
	}
	if got := x.counts(t); got != (counts{}) {
		t.Fatalf("after a failed audit write: %+v, want none", got)
	}
}

// TestNestedFailureRollsBackAlone checks that a nested failure the outer
// function handles undoes only the nested part.
func TestNestedFailureRollsBackAlone(t *testing.T) {
	t.Parallel()
	x := newTxFixture(t)
	err := x.store.Atomically(t.Context(), func(ctx context.Context) error {
		if createErr := x.created(ctx); createErr != nil {
			return createErr
		}
		inner := x.store.Atomically(ctx, func(ctx context.Context) error {
			e := signIn(issuedAt(), auditOther)
			if appendErr := x.store.AppendAudit(ctx, &e); appendErr != nil {
				return fmt.Errorf("appending the entry: %w", appendErr)
			}
			return errTest
		})
		if !errors.Is(inner, errTest) {
			return fmt.Errorf("the nested transaction: %w", inner)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Atomically with a handled nested failure = %v", err)
	}
	if got := x.counts(t); got != (counts{tokens: 1, entries: 1}) {
		t.Fatalf("after the committed transaction: %+v, want 1 token and 1 entry", got)
	}
}

// auditDown is the store with an audit log that can't be written.
type auditDown struct{ *postgres.Store }

func (auditDown) AppendAudit(context.Context, *identity.AuditEntry) error { return errTest }

// authed is a core with one owner, a passkey and a signed-in session, over
// one database, built with a healthy audit log; down is the same core with
// an audit log that fails.
type authed struct {
	pool    *pgxpool.Pool
	auth    *app.Auth
	down    *app.Auth
	dev     *virtual.Authenticator
	session string
}

func newAuthed(t *testing.T) *authed {
	t.Helper()
	pool := newPool(t)
	store := postgres.New(pool)
	now := issuedAt()
	clock := func() time.Time { return now }
	origins, err := identity.NewOrigins([]string{localUI})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	cfg := app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy()}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), app.NewTokens(store, clock, nil), nil, cfg, clock, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	downStore := auditDown{store}
	down, err := app.NewAuth(downStore, webauthn.New("OpenFief"), app.NewTokens(downStore, clock, nil), nil, cfg, clock, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	link, err := app.IssueEnrollmentLink(t.Context(), store, nil, now, origins, localOrigin(t), 0,
		identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID})
	if err != nil {
		t.Fatalf("IssueEnrollmentLink: %v", err)
	}
	_, code, _ := strings.Cut(link.URL, "#code=")
	e := enrollOver(t, auth, code)
	return &authed{pool: pool, auth: auth, down: down, dev: e.dev, session: e.session}
}

func (a *authed) count(t *testing.T, table string) int {
	t.Helper()
	var n int
	if err := a.pool.QueryRow(t.Context(), `SELECT count(*) FROM identity.`+table).Scan(&n); err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}
	return n
}

// must fails the test now when err is not nil.
func must(t *testing.T, what string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

// ownerActions runs the use cases of an owner session against the real store,
// and returns the secrets they handled.
func (a *authed) ownerActions(t *testing.T) []string {
	t.Helper()
	ctx := t.Context()
	loginOver(t, a.auth, a.dev)
	// A second passkey, created and removed in the owner session.
	dev2, err := virtual.New(nil)
	must(t, "virtual.New", err)
	st, err := a.auth.BeginRegistration(ctx, localUI, app.RegistrationGrant{SessionToken: a.session})
	must(t, "BeginRegistration", err)
	cred, err := dev2.Create(st.Options, localUI)
	must(t, "Create", err)
	added, err := a.auth.FinishRegistration(ctx, localUI, st.CeremonyID, cred, "phone")
	must(t, "FinishRegistration", err)
	_, err = a.auth.RemovePasskey(ctx, localUI, a.session, added.Passkey.ID())
	must(t, "RemovePasskey", err)
	tok, secret, err := a.auth.CreateScriptToken(ctx, localUI, a.session, app.NewToken{Name: "nightly"})
	must(t, "CreateScriptToken", err)
	must(t, "RevokeScriptToken", a.auth.RevokeScriptToken(ctx, localUI, a.session, string(tok.ID)))
	_, err = a.auth.SignOutEverywhere(ctx, localUI, a.session)
	must(t, "SignOutEverywhere", err)
	// A refused sign-in is recorded on its own.
	_, err = a.auth.Login(ctx, a.auth.Passkey(), localUI, app.Proof{CeremonyID: "nope", Credential: []byte("{}")})
	if !errors.Is(err, app.ErrCeremony) {
		t.Fatalf("Login of an unknown ceremony = %v, want %v", err, app.ErrCeremony)
	}
	return []string{secret, a.session}
}

// trail returns the audit log as one line per entry, oldest first.
func (a *authed) trail(t *testing.T) []string {
	t.Helper()
	page, err := postgres.New(a.pool).AuditEntries(t.Context(), &identity.AuditFilter{})
	must(t, "AuditEntries", err)
	var lines []string
	for _, e := range slices.Backward(page.Entries) {
		words := []string{string(e.Actor.Kind), string(e.Action), string(e.Outcome)}
		for _, r := range e.Refs {
			words = append(words, string(r.Kind))
		}
		lines = append(lines, strings.Join(words, " "))
	}
	return lines
}

// keyOf returns the random part of a token, after its prefix and ID.
func keyOf(token string) string {
	_, rest, _ := strings.Cut(token, "_")
	_, key, _ := strings.Cut(rest, "_")
	return key
}

// TestAuditTrailOfOwnerActions runs the identity use cases against
// PostgreSQL and checks the entries they leave: who, what, which records,
// and no secret.
func TestAuditTrailOfOwnerActions(t *testing.T) {
	t.Parallel()
	a := newAuthed(t)
	secrets := a.ownerActions(t)
	want := []string{
		"system enrollment_link_issued ok enrollment_code",
		"enrollment_code enrollment_link_used ok enrollment_code passkey",
		"enrollment_code passkey_added ok passkey",
		"owner sign_in ok session", // the enrollment's session
		"owner sign_in ok session", // the login's
		"owner passkey_added ok passkey",
		"owner passkey_removed ok passkey",
		"owner script_token_created ok script_token",
		"owner script_token_revoked ok script_token",
		"owner sign_out_everywhere ok",
		"anonymous sign_in refused",
	}
	if diff := cmp.Diff(want, a.trail(t)); diff != "" {
		t.Errorf("audit trail (-want +got):\n%s", diff)
	}
	var dump string
	must(t, "dumping the log", a.pool.QueryRow(t.Context(), `SELECT string_agg(l::text, ' ') FROM identity.audit_log l`).Scan(&dump))
	for _, secret := range secrets {
		if key := keyOf(secret); len(key) < 20 || strings.Contains(dump, key) || strings.Contains(dump, secret) {
			t.Errorf("the audit log holds a secret's key (or the test lost it)")
		}
	}
}

// TestAuditFailsClosedOnCreate checks D-211's rule that a change of state and
// its audit entry commit together: with an audit log that can't be written,
// creating a token, a link and a session each fails and leaves nothing.
func TestAuditFailsClosedOnCreate(t *testing.T) {
	t.Parallel()
	a := newAuthed(t)
	ctx := t.Context()
	tokens, sessions, links := a.count(t, "script_tokens"), a.count(t, "sessions"), a.count(t, "enrollment_codes")
	_, _, err := a.down.CreateScriptToken(ctx, localUI, a.session, app.NewToken{Name: "nightly"})
	if !errors.Is(err, errTest) {
		t.Errorf("CreateScriptToken = %v, want %v", err, errTest)
	}
	_, err = a.down.CreateEnrollmentLink(ctx, localUI, a.session, "", 0)
	if !errors.Is(err, errTest) {
		t.Errorf("CreateEnrollmentLink = %v, want %v", err, errTest)
	}
	st, err := a.auth.BeginLogin(ctx, localUI)
	must(t, "BeginLogin", err)
	assertion, err := a.dev.Get(st.Options, localUI)
	must(t, "Get", err)
	_, err = a.down.Login(ctx, a.down.Passkey(), localUI, app.Proof{CeremonyID: st.CeremonyID, Credential: assertion})
	if !errors.Is(err, errTest) {
		t.Errorf("Login = %v, want %v", err, errTest)
	}
	for table, want := range map[string]int{"script_tokens": tokens, "sessions": sessions, "enrollment_codes": links} {
		if got := a.count(t, table); got != want {
			t.Errorf("%d rows in %s after the failed changes, want %d", got, table, want)
		}
	}
}

// TestAuditFailsClosedOnRevoke checks the same for ending things: a token is
// not revoked and no session ends when the entry can't be written.
func TestAuditFailsClosedOnRevoke(t *testing.T) {
	t.Parallel()
	a := newAuthed(t)
	ctx := t.Context()
	tok, _, err := a.auth.CreateScriptToken(ctx, localUI, a.session, app.NewToken{Name: "nightly"})
	must(t, "CreateScriptToken", err)
	entries := a.count(t, "audit_log")
	err = a.down.RevokeScriptToken(ctx, localUI, a.session, string(tok.ID))
	if !errors.Is(err, errTest) {
		t.Errorf("RevokeScriptToken = %v, want %v", err, errTest)
	}
	got, _, err := postgres.New(a.pool).ScriptToken(ctx, tok.ID)
	if err != nil || !got.RevokedAt.IsZero() {
		t.Errorf("token after the failed revoke = %+v, %v; want it not revoked", got, err)
	}
	_, err = a.down.SignOutEverywhere(ctx, localUI, a.session)
	if !errors.Is(err, errTest) {
		t.Errorf("SignOutEverywhere = %v, want %v", err, errTest)
	}
	if err = a.down.Logout(ctx, localUI, a.session); !errors.Is(err, errTest) {
		t.Errorf("Logout = %v, want %v", err, errTest)
	}
	if _, err = a.auth.Passkeys(ctx, localUI, a.session); err != nil {
		t.Errorf("the session after the failed sign-outs: %v, want it still valid", err)
	}
	if got := a.count(t, "audit_log"); got != entries {
		t.Errorf("%d entries after failed changes, want %d", got, entries)
	}
}
