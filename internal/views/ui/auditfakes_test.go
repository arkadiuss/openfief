package ui

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// The fake core's audit log: its sessions and the viewer Alice's, a link that
// is gone and the removed account acc-deleted-1.
const (
	ownerAuditSession  = "0123456789abcde0"
	viewerAuditSession = "aaaaaaaabbbbbbbb"
	goneLink           = "99999999aaaaaaaa"
	tokenID            = "fedcba9876543210"
)

// errBadPageToken is the fake core's refusal of a page token that is none.
var errBadPageToken = errors.New("bad page token")

// fakeAuditLog returns the fake core's entries, oldest first, with seq 1 to
// 12; the first six are on 1 March, the next three on 2 March, the rest on 3.
func fakeAuditLog() []*identityv1.AuditEntry {
	owner := &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER, Id: ownerAuditSession}
	link := &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ENROLLMENT_CODE, Id: goneLink}
	viewer := &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_VIEWER, Id: viewerAuditSession, PersonId: alicePerson}
	system := &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_SYSTEM, Id: "core"}
	anon := &identityv1.AuditActor{Kind: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ANONYMOUS}
	ref := func(k identityv1.AuditRefKind, id string) *identityv1.AuditRef {
		return &identityv1.AuditRef{Kind: k, Id: id}
	}
	const (
		ok      = identityv1.AuditOutcome_AUDIT_OUTCOME_OK
		refused = identityv1.AuditOutcome_AUDIT_OUTCOME_REFUSED
		person  = identityv1.AuditRefKind_AUDIT_REF_KIND_PERSON
		code    = identityv1.AuditRefKind_AUDIT_REF_KIND_ENROLLMENT_CODE
	)
	at := func(day, n int) *timestamppb.Timestamp {
		return timestamppb.New(time.Date(2026, 3, day, 9, n, 0, 0, time.UTC))
	}
	entries := []*identityv1.AuditEntry{
		{At: at(1, 1), Actor: system, Action: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_ISSUED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(code, "00aa000000000001")}},
		{At: at(1, 2), Actor: owner, Action: identityv1.AuditAction_AUDIT_ACTION_SIGN_IN, Outcome: ok, Refs: []*identityv1.AuditRef{ref(identityv1.AuditRefKind_AUDIT_REF_KIND_SESSION, ownerAuditSession), ref(person, "owner")}},
		{At: at(1, 3), Actor: owner, Action: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_ISSUED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(code, goneLink)}},
		{At: at(1, 4), Actor: link, Action: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_USED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(code, goneLink), ref(identityv1.AuditRefKind_AUDIT_REF_KIND_PASSKEY, "cGFzc2tleS1pZC0x"), ref(person, alicePerson)}},
		{At: at(1, 5), Actor: link, Action: identityv1.AuditAction_AUDIT_ACTION_PERSON_CREATED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(person, alicePerson), ref(code, goneLink)}},
		{At: at(1, 6), Actor: link, Action: identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(person, alicePerson), ref(identityv1.AuditRefKind_AUDIT_REF_KIND_SLEEVE, "core")}},
		{At: at(2, 1), Actor: viewer, Action: identityv1.AuditAction_AUDIT_ACTION_SIGN_IN, Outcome: ok, Refs: []*identityv1.AuditRef{ref(identityv1.AuditRefKind_AUDIT_REF_KIND_SESSION, viewerAuditSession), ref(person, alicePerson)}},
		{At: at(2, 2), Actor: owner, Action: identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(identityv1.AuditRefKind_AUDIT_REF_KIND_SCRIPT_TOKEN, tokenID), ref(identityv1.AuditRefKind_AUDIT_REF_KIND_ACCOUNT, "acc-deleted-1")}},
		{At: at(2, 3), Actor: owner, Action: identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(identityv1.AuditRefKind_AUDIT_REF_KIND_VIEW_PLUGIN, "ui-local"), ref(identityv1.AuditRefKind_AUDIT_REF_KIND_HOUSEHOLD, "home")}},
		{At: at(3, 1), Actor: anon, Action: identityv1.AuditAction_AUDIT_ACTION_SIGN_IN, Outcome: refused},
		{At: at(3, 2), Actor: link, Action: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_USED, Outcome: refused, Refs: []*identityv1.AuditRef{ref(code, goneLink)}},
		{At: at(3, 3), Actor: owner, Action: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_CANCELLED, Outcome: ok, Refs: []*identityv1.AuditRef{ref(code, pendingInvite)}},
	}
	for i, e := range entries {
		e.Seq = int64(i + 1)
	}
	return entries
}

// auditMatches reports whether e passes the filter of r, except its cursor.
func auditMatches(r *identityv1.ListAuditEntriesRequest, e *identityv1.AuditEntry) bool {
	byActor := r.GetActorKind() == identityv1.AuditActorKind_AUDIT_ACTOR_KIND_UNSPECIFIED || e.GetActor().GetKind() == r.GetActorKind()
	byAction := (r.GetAction() == identityv1.AuditAction_AUDIT_ACTION_UNSPECIFIED || e.GetAction() == r.GetAction()) &&
		(len(r.GetActions()) == 0 || slices.Contains(r.GetActions(), e.GetAction()))
	byOutcome := r.GetOutcome() == identityv1.AuditOutcome_AUDIT_OUTCOME_UNSPECIFIED || e.GetOutcome() == r.GetOutcome()
	return byActor && byAction && byOutcome && auditInRange(r, e.GetAt().AsTime())
}

// auditInRange reports whether at is at or after the request's from and before
// its to; an unset bound is open.
func auditInRange(r *identityv1.ListAuditEntriesRequest, at time.Time) bool {
	return (r.GetFrom() == nil || !at.Before(r.GetFrom().AsTime())) && (r.GetTo() == nil || at.Before(r.GetTo().AsTime()))
}

// auditPage returns the page of the fake log r asks for, as the core pages it:
// newest first, at most size entries, the page token the seq of the last entry
// shown when more follow.
func auditPage(r *identityv1.ListAuditEntriesRequest, size int) (*identityv1.ListAuditEntriesResponse, error) {
	before := int64(1 << 62)
	if tok := r.GetPageToken(); tok != "" {
		n, err := strconv.ParseInt(tok, 10, 64)
		if err != nil || n <= 0 {
			return nil, codeErr(connect.CodeInvalidArgument, errBadPageToken)
		}
		before = n
	}
	out := &identityv1.ListAuditEntriesResponse{}
	var shown []*identityv1.AuditEntry
	for _, e := range slices.Backward(fakeAuditLog()) {
		if e.GetSeq() >= before || !auditMatches(r, e) {
			continue
		}
		if len(shown) == size {
			out.NextPageToken = strconv.FormatInt(shown[size-1].GetSeq(), 10)
			break
		}
		shown = append(shown, e)
	}
	out.Entries = shown
	return out, nil
}

// ListAuditEntries answers the owner's call with the fake log, filtered and
// paged as the core does.
func (f *fakeAuth) ListAuditEntries(_ context.Context, r *connect.Request[identityv1.ListAuditEntriesRequest]) (*connect.Response[identityv1.ListAuditEntriesResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.auditReqs = append(f.auditReqs, r.Msg)
	failed, limit := f.auditErr, f.auditCap
	f.mu.Unlock()
	if failed != nil {
		return nil, failed
	}
	size := cmp.Or(int(r.Msg.GetPageSize()), 50)
	if limit > 0 {
		size = min(size, limit)
	}
	page, err := auditPage(r.Msg, size)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(page), nil
}
