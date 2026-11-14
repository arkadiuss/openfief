package postgres_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
)

const viewerID = "0123456789abcdef"

func viewerPerson(t *testing.T, id, name string) identity.Person {
	t.Helper()
	p, err := identity.NewViewer(id, name, bytes.Repeat([]byte{5}, identity.OwnerHandleBytes), issuedAt())
	if err != nil {
		t.Fatalf("NewViewer: %v", err)
	}
	return p
}

// storeOlderData stores what a core from before people stored: an owner, a
// passkey and a session.
func storeOlderData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	for _, q := range []string{
		`INSERT INTO identity.owner (id, handle, created_at) VALUES (1, decode(repeat('07', 32), 'hex'), '2026-10-01T09:00:00Z')`,
		`INSERT INTO identity.passkeys (credential_id, rp_id, name, public_key, attestation_type, attestation_format, aaguid,
			sign_count, transports, user_verified, backup_eligible, backup_state, created_at)
			VALUES ('\x0909', 'localhost', 'old laptop', '\x010203', 'none', 'none', '\x', 0, '{internal}', true, false, FALSE, '2026-10-01T09:00:00Z')`,
		`INSERT INTO identity.sessions (id, scope, authenticator, rp_id, key_sha256, created_at, last_seen_at, expires_at, idle_seconds)
			VALUES ('aaaaaaaaaaaaaaaa', 'owner', 'passkey', 'localhost', decode(repeat('42', 32), 'hex'), '2026-10-01T09:00:00Z',
			'2026-10-01T09:00:00Z', '2026-10-08T09:00:00Z', 43200)`,
	} {
		if _, err := pool.Exec(t.Context(), q); err != nil {
			t.Fatalf("storing what an older core stored: %v", err)
		}
	}
}

// TestMigrationGivesExistingPasskeysAndSessionsToTheOwner checks that what an
// older core stored, a passkey, a session and the owner, belongs to the owner
// once the people migration ran, and keeps working (D-65).
func TestMigrationGivesExistingPasskeysAndSessionsToTheOwner(t *testing.T) {
	t.Parallel()
	pool := emptyPool(t)
	ctx := t.Context()
	if err := postgres.MigrateThrough(ctx, pool, 9); err != nil {
		t.Fatalf("migrating to the version before people: %v", err)
	}
	storeOlderData(t, pool)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	checkOlderDataIsTheOwners(t, pool)
}

// checkOnlyTheOwnerIsAPerson checks that the owner is the one person, created
// when the owner was.
func checkOnlyTheOwnerIsAPerson(t *testing.T, store *postgres.Store) {
	t.Helper()
	people, err := store.People(t.Context())
	if err != nil || len(people) != 1 || people[0].ID != identity.OwnerPersonID || people[0].Kind != identity.PersonOwner ||
		!people[0].CreatedAt.Equal(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("People = %+v, %v, want the owner, created when the owner was", people, err)
	}
}

// checkOlderDataIsTheOwners checks that the passkey, the session and the owner
// storeOlderData stored are the owner's and work.
func checkOlderDataIsTheOwners(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	store := postgres.New(pool)
	keys, err := store.Passkeys(ctx)
	if err != nil || len(keys) != 1 || keys[0].Person() != identity.OwnerPersonID || !keys[0].DisabledAt.IsZero() {
		t.Fatalf("Passkeys = %+v, %v, want the old passkey, the owner's and working", keys, err)
	}
	s, _, err := store.Session(ctx, "aaaaaaaaaaaaaaaa")
	if err != nil || s.OfViewer() || s.Scope != identity.SessionScopeOwner {
		t.Errorf("Session = %+v, %v, want the owner's session", s, err)
	}
	checkOnlyTheOwnerIsAPerson(t, store)
	var column string
	if err := pool.QueryRow(ctx, `SELECT person_id FROM identity.passkeys`).Scan(&column); err != nil || column != "owner" {
		t.Errorf("passkeys.person_id = %q, %v, want owner", column, err)
	}
}

// TestMigrationOnAnEmptyInstallationHasTheOwnerRow checks that a database
// with no owner yet still has the owner as a person.
func TestMigrationOnAnEmptyInstallationHasTheOwnerRow(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	p, err := store.Person(t.Context(), identity.OwnerPersonID)
	if err != nil || p.Kind != identity.PersonOwner || p.Name != "Owner" {
		t.Errorf("Person(owner) = %+v, %v", p, err)
	}
	if _, err := store.Person(t.Context(), viewerID); !errors.Is(err, app.ErrUnknownPerson) {
		t.Errorf("Person(unknown) = %v, want %v", err, app.ErrUnknownPerson)
	}
}

// viewerWorld is a store with the owner and one viewer, each with a passkey
// and a session.
type viewerWorld struct {
	store                  *postgres.Store
	ownerKey, viewerKey    identity.Passkey
	ownerSession, viewerSS identity.Session
	viewer                 identity.Person
}

func newViewerWorld(t *testing.T) *viewerWorld {
	t.Helper()
	w := &viewerWorld{store: postgres.New(newPool(t)), ownerKey: passkeyFor(1), viewerKey: passkeyFor(2), viewer: viewerPerson(t, viewerID, "Alice")}
	w.viewerKey.PersonID = viewerID
	ctx := t.Context()
	if err := w.store.SavePerson(ctx, &w.viewer); err != nil {
		t.Fatalf("SavePerson: %v", err)
	}
	for _, k := range []*identity.Passkey{&w.ownerKey, &w.viewerKey} {
		if err := w.store.SavePasskey(ctx, k); err != nil {
			t.Fatalf("SavePasskey: %v", err)
		}
	}
	w.ownerSession = savedSession(t, w.store, "1111111111111111", identity.AuthenticatorPasskey, w.ownerKey.CredentialID)
	vs, err := identity.NewSession("2222222222222222", identity.AuthenticatorPasskey, identity.SessionScopeRead, "localhost", issuedAt(), identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if w.viewerSS, err = vs.ForViewer(viewerID); err != nil {
		t.Fatal(err)
	}
	if err := w.store.SaveSession(ctx, &w.viewerSS, identity.HashKey(sessionKey())); err != nil {
		t.Fatalf("SaveSession of a viewer's: %v", err)
	}
	return w
}

// TestViewerPasskeysDoNotCountForTheOwner checks that a viewer's passkeys
// never count for the owner's last-passkey rule nor are removable as the
// owner's, and that passkeys and sessions are counted per person.
func TestViewerPasskeysDoNotCountForTheOwner(t *testing.T) {
	t.Parallel()
	w := newViewerWorld(t)
	ctx := t.Context()
	if err := w.store.DeletePasskey(ctx, w.ownerKey.CredentialID, issuedAt()); !errors.Is(err, app.ErrLastPasskey) {
		t.Errorf("DeletePasskey of the owner's only passkey = %v, want %v although a viewer has one", err, app.ErrLastPasskey)
	}
	if err := w.store.DeletePasskey(ctx, w.viewerKey.CredentialID, issuedAt()); !errors.Is(err, app.ErrUnknownPasskey) {
		t.Errorf("DeletePasskey of a viewer's passkey = %v, want %v", err, app.ErrUnknownPasskey)
	}
	w.checkCounts(t)
}

// checkCounts checks the passkeys and sessions counted per person, and that
// a session past its idle timeout is not counted.
func (w *viewerWorld) checkCounts(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	keys, err := w.store.PasskeyCounts(ctx)
	if err != nil || keys[identity.OwnerPersonID] != 1 || keys[viewerID] != 1 {
		t.Errorf("PasskeyCounts = %v, %v", keys, err)
	}
	sessions, err := w.store.ActiveSessionCounts(ctx, issuedAt().Add(time.Minute))
	if err != nil || sessions[identity.OwnerPersonID] != 1 || sessions[viewerID] != 1 {
		t.Errorf("ActiveSessionCounts = %v, %v", sessions, err)
	}
	past, err := w.store.ActiveSessionCounts(ctx, issuedAt().Add(13*time.Hour))
	if err != nil || past[viewerID] != 0 {
		t.Errorf("ActiveSessionCounts after the idle timeout = %v, %v, want none", past, err)
	}
}

// TestRevokePersonDisablesPasskeysAndEndsSessions checks that revoking a
// viewer disables their passkeys and ends their sessions in one step, leaving
// the owner's, and keeps the person.
func TestRevokePersonDisablesPasskeysAndEndsSessions(t *testing.T) {
	t.Parallel()
	w := newViewerWorld(t)
	ctx := t.Context()
	disabled, ended, err := w.store.RevokePerson(ctx, viewerID, issuedAt().Add(time.Hour))
	if err != nil || ended != 1 || len(disabled) != 1 || !bytes.Equal(disabled[0].CredentialID, w.viewerKey.CredentialID) {
		t.Fatalf("RevokePerson = %+v, %d, %v, want the viewer's passkey and one session", disabled, ended, err)
	}
	w.checkAfterRevocation(t)
	for name, id := range map[string]string{"again": viewerID, "the owner": identity.OwnerPersonID, "an unknown one": "fedcba9876543210"} {
		if _, _, revokeErr := w.store.RevokePerson(ctx, id, issuedAt()); !errors.Is(revokeErr, app.ErrUnknownPerson) {
			t.Errorf("RevokePerson %s = %v, want %v", name, revokeErr, app.ErrUnknownPerson)
		}
	}
}

// checkSessionsAfterRevocation checks that the viewer's session is revoked
// and the owner's untouched.
func (w *viewerWorld) checkSessionsAfterRevocation(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	if s, _, err := w.store.Session(ctx, w.viewerSS.ID); err != nil || s.RevokedAt.IsZero() {
		t.Errorf("the viewer's session = %+v, %v, want revoked", s, err)
	}
	if s, _, err := w.store.Session(ctx, w.ownerSession.ID); err != nil || !s.RevokedAt.IsZero() {
		t.Errorf("the owner's session = %+v, %v, want untouched", s, err)
	}
}

// checkAfterRevocation checks that only the owner's passkey and session still
// work, and that the viewer is kept, revoked.
func (w *viewerWorld) checkAfterRevocation(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	got, err := w.store.Passkeys(ctx)
	if err != nil || len(got) != 1 || !bytes.Equal(got[0].CredentialID, w.ownerKey.CredentialID) {
		t.Errorf("Passkeys after = %+v, %v, want only the owner's", got, err)
	}
	w.checkSessionsAfterRevocation(t)
	w.checkViewerKept(t)
}

// checkViewerKept checks that the owner and then the revoked viewer, with
// their handle, are still the people.
func (w *viewerWorld) checkViewerKept(t *testing.T) {
	t.Helper()
	people, err := w.store.People(t.Context())
	if err != nil || len(people) != 2 || people[0].ID != identity.OwnerPersonID || people[1].ID != viewerID {
		t.Fatalf("People = %+v, %v, want the owner, then the viewer", people, err)
	}
	if v := people[1]; v.Active() || !v.RevokedAt.Equal(issuedAt().Add(time.Hour)) || !bytes.Equal(v.Handle, w.viewer.Handle) {
		t.Errorf("the viewer = %+v, want revoked an hour after, with their handle kept", v)
	}
}

// TestViewerRowsRefuseWhatAViewerMayNotHave checks the database's own limits:
// a viewer's session reads only, a viewer has a handle, the owner can't be
// revoked, and the audit log takes viewer actors and the new actions.
func TestViewerRowsRefuseWhatAViewerMayNotHave(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	ctx := t.Context()
	v := viewerPerson(t, viewerID, "Alice")
	if err := store.SavePerson(ctx, &v); err != nil {
		t.Fatalf("SavePerson: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO identity.sessions (id, scope, authenticator, rp_id, key_sha256, created_at, last_seen_at, expires_at, idle_seconds, person_id)
		VALUES ('3333333333333333', 'owner', 'passkey', 'localhost', decode(repeat('42', 32), 'hex'), current_timestamp, localtimestamp, current_timestamp + interval '1 day', 60, $1)`, viewerID); err == nil {
		t.Error("the database stored an owner-scoped session of a viewer")
	}
	if _, err := pool.Exec(ctx, `INSERT INTO identity.persons (id, kind, name, created_at) VALUES ('fedcba9876543210', 'viewer', 'No handle', now())`); err == nil {
		t.Error("the database stored a viewer without a user handle")
	}
	if _, err := pool.Exec(ctx, `UPDATE identity.persons SET revoked_at = now() WHERE id = 'owner'`); err == nil {
		t.Error("the database revoked the owner")
	}
	for _, action := range []identity.AuditAction{identity.ActionPersonCreated, identity.ActionPersonRevoked} {
		e, err := identity.NewAuditEntry(issuedAt(), identity.Actor{Kind: identity.ActorViewer, ID: "4444444444444444"}, action, identity.OutcomeOK,
			identity.Ref{Kind: identity.RefPerson, ID: viewerID})
		if err != nil {
			t.Fatalf("NewAuditEntry: %v", err)
		}
		if err := store.AppendAudit(ctx, &e); err != nil {
			t.Errorf("AppendAudit(%s by a viewer): %v", action, err)
		}
	}
}

// TestViewerEnrollmentCodeRoundTrip checks that a link's viewer, name and
// grants come back as stored, and a plain code has none.
func TestViewerEnrollmentCodeRoundTrip(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	code, err := identity.NewEnrollmentCode("5555555555555555", localOrigin(t), issuedAt(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	targets := []identity.GrantTarget{{Kind: identity.GrantTargetSleeve, ID: "core"}, identity.Household()}
	if err = code.WithViewer(&identity.ViewerInvite{Name: "Alice", Targets: targets}); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveEnrollmentCode(ctx, &code, identity.HashKey(sessionKey())); err != nil {
		t.Fatalf("SaveEnrollmentCode: %v", err)
	}
	got, _, err := store.EnrollmentCode(ctx, code.ID)
	if err != nil {
		t.Fatalf("EnrollmentCode: %v", err)
	}
	if diff := cmp.Diff(&identity.ViewerInvite{Name: "Alice", Targets: targets}, got.Viewer); diff != "" {
		t.Errorf("Viewer (-want +got):\n%s", diff)
	}
	plain, err := identity.NewEnrollmentCode("6666666666666666", localOrigin(t), issuedAt(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.SaveEnrollmentCode(ctx, &plain, identity.HashKey(sessionKey())); err != nil {
		t.Fatalf("SaveEnrollmentCode: %v", err)
	}
	if got, _, err = store.EnrollmentCode(ctx, plain.ID); err != nil || got.Viewer != nil {
		t.Errorf("a plain code's Viewer = %+v, %v, want none", got.Viewer, err)
	}
}

// TestViewerCeremonyRoundTrip checks that a registration ceremony's person and
// user handle come back as stored.
func TestViewerCeremonyRoundTrip(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	code, err := identity.NewEnrollmentCode("5555555555555555", localOrigin(t), issuedAt(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.SaveEnrollmentCode(ctx, &code, identity.HashKey(sessionKey())); err != nil {
		t.Fatalf("SaveEnrollmentCode: %v", err)
	}
	c := identity.NewCeremony("77777777777777777777777777777777", identity.CeremonyRegistration, localOrigin(t), []byte("state"), issuedAt())
	c.EnrollmentCode, c.PersonID, c.PersonHandle = code.ID, viewerID, bytes.Repeat([]byte{5}, identity.OwnerHandleBytes)
	if err = store.SaveCeremony(ctx, &c); err != nil {
		t.Fatalf("SaveCeremony: %v", err)
	}
	taken, err := store.TakeCeremony(ctx, c.ID)
	if err != nil || taken.PersonID != viewerID || !bytes.Equal(taken.PersonHandle, c.PersonHandle) {
		t.Errorf("TakeCeremony = %+v, %v, want the viewer and handle", taken, err)
	}
}
