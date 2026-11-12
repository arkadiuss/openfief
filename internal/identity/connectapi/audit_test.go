package connectapi_test

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/identity"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// auditStart is the time of the first planted audit entry.
func auditStart() time.Time { return time.Date(2026, 10, 28, 11, 30, 0, 0, time.UTC) }

// auditFixture is the AuthService with a planted owner session and one audit
// entry per audit action.
type auditFixture struct {
	client  apiClient
	token   string
	entries int
}

func newAuditFixture(t *testing.T) auditFixture {
	t.Helper()
	c, store := newServed(t, viewCaller)
	token := staleOwnerSession(t, store)
	actions := identity.AuditActions()
	for i, a := range actions {
		e := identity.AuditEntry{
			At: auditStart().Add(time.Duration(i) * time.Minute), Actor: identity.Actor{Kind: identity.ActorAnonymous}, Action: a, Outcome: identity.OutcomeRefused,
		}
		if err := store.AppendAudit(t.Context(), &e); err != nil {
			t.Fatalf("AppendAudit: %v", err)
		}
	}
	return auditFixture{client: c, token: token, entries: len(actions)}
}

// list calls ListAuditEntries in the owner session and fails the test on an
// error.
func (a auditFixture) list(t *testing.T, req *identityv1.ListAuditEntriesRequest) *identityv1.ListAuditEntriesResponse {
	t.Helper()
	req.Origin, req.SessionToken = local, a.token
	resp, err := a.client.ListAuditEntries(t.Context(), connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListAuditEntries: %v", err)
	}
	return resp.Msg
}

// TestListAuditEntriesOwnerOnly checks that only an owner session reads the
// audit log, over a plugin's authenticated call: a demo session, no session,
// a script token's secret and a caller that is no view plugin are all
// refused (D-211, D-137).
func TestListAuditEntriesOwnerOnly(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	c := newClient(t, viewCaller)
	demo := demoSession(t, c)
	secret := "ofs_0123456789abcdef_" + strings.Repeat("A", 43)
	tests := []struct {
		name  string
		token string
		want  connect.Code
	}{
		{name: "demo (read) session", token: demo, want: connect.CodePermissionDenied},
		{name: "no session", token: "", want: connect.CodeUnauthenticated},
		{name: "script token secret", token: secret, want: connect.CodeUnauthenticated},
		{name: "unknown session", token: identity.FormatSessionToken("0123456789abcdef", make([]byte, identity.KeyBytes)), want: connect.CodeUnauthenticated},
	}
	for _, tc := range tests {
		_, err := c.ListAuditEntries(ctx, connect.NewRequest(&identityv1.ListAuditEntriesRequest{Origin: local, SessionToken: tc.token}))
		wantCode(t, "ListAuditEntries with a "+tc.name, err, tc.want)
	}
	for name, caller := range map[string]string{"no caller token": "", "a caller that is no view plugin": "connector-token"} {
		_, err := newClient(t, caller).ListAuditEntries(ctx, connect.NewRequest(&identityv1.ListAuditEntriesRequest{Origin: local, SessionToken: demo}))
		wantCode(t, "ListAuditEntries from "+name, err, connect.CodePermissionDenied)
	}
	_, err := c.ListAuditEntries(ctx, connect.NewRequest(&identityv1.ListAuditEntriesRequest{Origin: "https://evil.example.org", SessionToken: demo}))
	wantCode(t, "ListAuditEntries on an origin off the list", err, connect.CodePermissionDenied)
}

// TestListAuditEntriesNewestFirst checks every audit action has a contract
// value, so the domain has none the contract lacks, and the order.
func TestListAuditEntriesNewestFirst(t *testing.T) {
	t.Parallel()
	a := newAuditFixture(t)
	all := a.list(t, &identityv1.ListAuditEntriesRequest{})
	if len(all.GetEntries()) != a.entries || all.GetNextPageToken() != "" {
		t.Fatalf("got %d entries and token %q, want %d and none", len(all.GetEntries()), all.GetNextPageToken(), a.entries)
	}
	for i, e := range all.GetEntries() {
		if e.GetAction() == identityv1.AuditAction_AUDIT_ACTION_UNSPECIFIED || e.GetOutcome() != identityv1.AuditOutcome_AUDIT_OUTCOME_REFUSED ||
			e.GetActor().GetKind() != identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ANONYMOUS || e.GetAt() == nil {
			t.Errorf("entry %d = %v: a contract value is unspecified, so the domain has a value the contract lacks", i, e)
		}
		if i > 0 && e.GetSeq() >= all.GetEntries()[i-1].GetSeq() {
			t.Errorf("entry %d is not older than entry %d: want newest first", i, i-1)
		}
	}
}

// TestListAuditEntriesPages checks that pages of four walk every entry once.
func TestListAuditEntriesPages(t *testing.T) {
	t.Parallel()
	a := newAuditFixture(t)
	seen := map[int64]bool{}
	req := &identityv1.ListAuditEntriesRequest{PageSize: 4}
	for pages := 0; pages <= a.entries; pages++ {
		page := a.list(t, req)
		for _, e := range page.GetEntries() {
			if seen[e.GetSeq()] {
				t.Errorf("entry %d on two pages", e.GetSeq())
			}
			seen[e.GetSeq()] = true
		}
		if page.GetNextPageToken() == "" {
			break
		}
		req = &identityv1.ListAuditEntriesRequest{PageSize: 4, PageToken: page.GetNextPageToken()}
	}
	if len(seen) != a.entries {
		t.Errorf("paging saw %d entries, want %d", len(seen), a.entries)
	}
}

// TestListAuditEntriesFilters checks the filters by action, time and actor.
func TestListAuditEntriesFilters(t *testing.T) {
	t.Parallel()
	a := newAuditFixture(t)
	one := a.list(t, &identityv1.ListAuditEntriesRequest{Action: identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT})
	if len(one.GetEntries()) != 1 || one.GetEntries()[0].GetAction() != identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT {
		t.Errorf("filter by action = %v; want the one sign-out entry", one)
	}
	from := auditStart().Add(2 * time.Minute)
	since := a.list(t, &identityv1.ListAuditEntriesRequest{From: timestamppb.New(from), To: timestamppb.New(from.Add(2 * time.Minute))})
	if len(since.GetEntries()) != 2 {
		t.Errorf("filter by time = %v; want 2 entries", since)
	}
	none := a.list(t, &identityv1.ListAuditEntriesRequest{ActorKind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER})
	if len(none.GetEntries()) != 0 {
		t.Errorf("filter by actor kind = %v; want none", none)
	}
}

func TestListAuditEntriesBadRequests(t *testing.T) {
	t.Parallel()
	a := newAuditFixture(t)
	tests := []struct {
		name string
		req  *identityv1.ListAuditEntriesRequest
	}{
		{name: "page token that is no number", req: &identityv1.ListAuditEntriesRequest{PageToken: "not-a-token"}},
		{name: "negative page token", req: &identityv1.ListAuditEntriesRequest{PageToken: "-4"}},
		{name: "page size too large", req: &identityv1.ListAuditEntriesRequest{PageSize: identityMaxPage + 1}},
		{name: "negative page size", req: &identityv1.ListAuditEntriesRequest{PageSize: -1}},
		{name: "actor id without a kind", req: &identityv1.ListAuditEntriesRequest{ActorId: "0123456789abcdef"}},
		{name: "actor id that is no id", req: &identityv1.ListAuditEntriesRequest{ActorKind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER, ActorId: "Jan"}},
		{name: "unknown action", req: &identityv1.ListAuditEntriesRequest{Action: 99}},
		{name: "unknown actor kind", req: &identityv1.ListAuditEntriesRequest{ActorKind: 99}},
		{name: "empty range", req: &identityv1.ListAuditEntriesRequest{From: timestamppb.New(time.Unix(2000, 0)), To: timestamppb.New(time.Unix(1000, 0))}},
	}
	for _, tc := range tests {
		tc.req.Origin, tc.req.SessionToken = local, a.token
		_, err := a.client.ListAuditEntries(t.Context(), connect.NewRequest(tc.req))
		wantCode(t, "ListAuditEntries with a "+tc.name, err, connect.CodeInvalidArgument)
	}
}

// identityMaxPage is identity.MaxAuditPage, as the contract's int32.
const identityMaxPage = int32(identity.MaxAuditPage)
