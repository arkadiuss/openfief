package connectapi

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// authService implements the generated AuthServiceHandler over the auth use
// cases. Every error goes through toConnectError.
type authService struct {
	auth *app.Auth
	log  *slog.Logger
}

var _ identityv1connect.AuthServiceHandler = (*authService)(nil)

// GetLoginMethods answers which ways in the login page can offer.
func (s *authService) GetLoginMethods(ctx context.Context, req *connect.Request[identityv1.GetLoginMethodsRequest]) (*connect.Response[identityv1.GetLoginMethodsResponse], error) {
	m, err := s.auth.LoginMethods(ctx, req.Msg.GetOrigin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.GetLoginMethodsResponse{PasskeyRegistered: m.PasskeyRegistered, DemoSessions: m.DemoSessions, DemoLoading: m.DemoLoading}), nil
}

// BeginRegistration starts creating a passkey.
func (s *authService) BeginRegistration(ctx context.Context, req *connect.Request[identityv1.BeginRegistrationRequest]) (*connect.Response[identityv1.BeginRegistrationResponse], error) {
	grant := app.RegistrationGrant{EnrollmentCode: req.Msg.GetEnrollmentCode(), SessionToken: req.Msg.GetSessionToken()}
	st, err := s.auth.BeginRegistration(ctx, req.Msg.GetOrigin(), grant)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.BeginRegistrationResponse{
		CeremonyId: st.CeremonyID, OptionsJson: string(st.Options), ExpiresAt: timestamppb.New(st.ExpiresAt), InvitedName: st.InvitedName,
	}), nil
}

// FinishRegistration stores the new passkey.
func (s *authService) FinishRegistration(ctx context.Context, req *connect.Request[identityv1.FinishRegistrationRequest]) (*connect.Response[identityv1.FinishRegistrationResponse], error) {
	m := req.Msg
	r, err := s.auth.FinishRegistration(ctx, m.GetOrigin(), m.GetCeremonyId(), []byte(m.GetCredentialJson()), m.GetName())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	s.logNotProjected(ctx, r.NotProjected)
	out := &identityv1.FinishRegistrationResponse{Passkey: pbPasskey(&r.Passkey)}
	if r.Session != nil {
		out.SessionToken, out.Session = r.Session.Token, pbSession(&r.Session.Session)
	}
	return connect.NewResponse(out), nil
}

// BeginLogin starts a passkey login.
func (s *authService) BeginLogin(ctx context.Context, req *connect.Request[identityv1.BeginLoginRequest]) (*connect.Response[identityv1.BeginLoginResponse], error) {
	st, err := s.auth.BeginLogin(ctx, req.Msg.GetOrigin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.BeginLoginResponse{
		CeremonyId: st.CeremonyID, OptionsJson: string(st.Options), ExpiresAt: timestamppb.New(st.ExpiresAt),
	}), nil
}

// FinishLogin logs in with the passkey authenticator.
func (s *authService) FinishLogin(ctx context.Context, req *connect.Request[identityv1.FinishLoginRequest]) (*connect.Response[identityv1.FinishLoginResponse], error) {
	m := req.Msg
	ns, err := s.auth.Login(ctx, s.auth.Passkey(), m.GetOrigin(), app.Proof{CeremonyID: m.GetCeremonyId(), Credential: []byte(m.GetCredentialJson())})
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.FinishLoginResponse{SessionToken: ns.Token, Session: pbSession(&ns.Session)}), nil
}

// CreateDemoSession logs in with the demo authenticator.
func (s *authService) CreateDemoSession(ctx context.Context, req *connect.Request[identityv1.CreateDemoSessionRequest]) (*connect.Response[identityv1.CreateDemoSessionResponse], error) {
	ns, err := s.auth.Login(ctx, s.auth.Demo(), req.Msg.GetOrigin(), app.Proof{})
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.CreateDemoSessionResponse{SessionToken: ns.Token, Session: pbSession(&ns.Session)}), nil
}

// Logout revokes the session.
func (s *authService) Logout(ctx context.Context, req *connect.Request[identityv1.LogoutRequest]) (*connect.Response[identityv1.LogoutResponse], error) {
	if err := s.auth.Logout(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.LogoutResponse{}), nil
}

// WhoAmI returns the owner and the session.
func (s *authService) WhoAmI(ctx context.Context, req *connect.Request[identityv1.WhoAmIRequest]) (*connect.Response[identityv1.WhoAmIResponse], error) {
	owner, ss, person, err := s.auth.Identify(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.WhoAmIResponse{
		Owner: &identityv1.Owner{CreatedAt: timestamppb.New(owner.CreatedAt)}, Session: pbSession(&ss), Person: pbPerson(&person),
	}), nil
}

// ListPasskeys returns every passkey.
func (s *authService) ListPasskeys(ctx context.Context, req *connect.Request[identityv1.ListPasskeysRequest]) (*connect.Response[identityv1.ListPasskeysResponse], error) {
	keys, err := s.auth.Passkeys(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &identityv1.ListPasskeysResponse{Passkeys: make([]*identityv1.Passkey, 0, len(keys))}
	for i := range keys {
		out.Passkeys = append(out.Passkeys, pbPasskey(&keys[i]))
	}
	return connect.NewResponse(out), nil
}

// RemovePasskey removes a passkey.
func (s *authService) RemovePasskey(ctx context.Context, req *connect.Request[identityv1.RemovePasskeyRequest]) (*connect.Response[identityv1.RemovePasskeyResponse], error) {
	signedOut, err := s.auth.RemovePasskey(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetPasskeyId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.RemovePasskeyResponse{SignedOut: signedOut}), nil
}

// CreateScriptToken issues a script token.
func (s *authService) CreateScriptToken(ctx context.Context, req *connect.Request[identityv1.CreateScriptTokenRequest]) (*connect.Response[identityv1.CreateScriptTokenResponse], error) {
	m := req.Msg
	targets, err := grantTargets(m.GetGrants())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	spec := app.NewToken{Name: m.GetName(), TTL: m.GetTtl().AsDuration(), MCPAllowed: m.GetMcpAllowed(), Targets: targets}
	tok, secret, err := s.auth.CreateScriptToken(ctx, m.GetOrigin(), m.GetSessionToken(), spec)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := pbScriptToken(&tok, time.Now())
	for _, t := range targets {
		out.Grants = append(out.Grants, &identityv1.Grant{
			HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_SCRIPT_TOKEN, HolderId: string(tok.ID),
			TargetKind: pbTargetKinds()[t.Kind], TargetId: t.ID, CreatedAt: timestamppb.New(tok.CreatedAt),
		})
	}
	return connect.NewResponse(&identityv1.CreateScriptTokenResponse{Token: out, Secret: secret}), nil
}

// ListScriptTokens returns every script token.
func (s *authService) ListScriptTokens(ctx context.Context, req *connect.Request[identityv1.ListScriptTokensRequest]) (*connect.Response[identityv1.ListScriptTokensResponse], error) {
	toks, err := s.auth.ScriptTokens(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	now := time.Now()
	out := &identityv1.ListScriptTokensResponse{Tokens: make([]*identityv1.ScriptToken, 0, len(toks)), GrantsEnforced: s.auth.GrantsEnforced()}
	for i := range toks {
		t := pbScriptToken(&toks[i].Token, now)
		for j := range toks[i].Grants {
			t.Grants = append(t.Grants, pbGrant(&toks[i].Grants[j]))
		}
		out.Tokens = append(out.Tokens, t)
	}
	return connect.NewResponse(out), nil
}

// RevokeScriptToken revokes a script token.
func (s *authService) RevokeScriptToken(ctx context.Context, req *connect.Request[identityv1.RevokeScriptTokenRequest]) (*connect.Response[identityv1.RevokeScriptTokenResponse], error) {
	if err := s.auth.RevokeScriptToken(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetTokenId()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.RevokeScriptTokenResponse{}), nil
}

// CreateEnrollmentLink issues a one-time enrollment link.
func (s *authService) CreateEnrollmentLink(ctx context.Context, req *connect.Request[identityv1.CreateEnrollmentLinkRequest]) (*connect.Response[identityv1.CreateEnrollmentLinkResponse], error) {
	m := req.Msg
	link, err := s.auth.CreateEnrollmentLink(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetTargetOrigin(), m.GetTtl().AsDuration())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.CreateEnrollmentLinkResponse{Url: link.URL, ExpiresAt: timestamppb.New(link.ExpiresAt)}), nil
}

// BeginPasskeyCheck starts a passkey check in the session (D-183).
func (s *authService) BeginPasskeyCheck(ctx context.Context, req *connect.Request[identityv1.BeginPasskeyCheckRequest]) (*connect.Response[identityv1.BeginPasskeyCheckResponse], error) {
	st, err := s.auth.BeginPasskeyCheck(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.BeginPasskeyCheckResponse{
		CeremonyId: st.CeremonyID, OptionsJson: string(st.Options), ExpiresAt: timestamppb.New(st.ExpiresAt),
	}), nil
}

// FinishPasskeyCheck records a passkey check on the session (D-183).
func (s *authService) FinishPasskeyCheck(ctx context.Context, req *connect.Request[identityv1.FinishPasskeyCheckRequest]) (*connect.Response[identityv1.FinishPasskeyCheckResponse], error) {
	m := req.Msg
	ss, err := s.auth.FinishPasskeyCheck(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetCeremonyId(), []byte(m.GetCredentialJson()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.FinishPasskeyCheckResponse{Session: pbSession(&ss)}), nil
}

// SignOutEverywhere revokes every session (D-183).
func (s *authService) SignOutEverywhere(ctx context.Context, req *connect.Request[identityv1.SignOutEverywhereRequest]) (*connect.Response[identityv1.SignOutEverywhereResponse], error) {
	n, err := s.auth.SignOutEverywhere(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.SignOutEverywhereResponse{RevokedSessions: int64(n)}), nil
}

// ListAuditEntries returns a page of the audit log (D-211) for an owner
// session. A page token that is not one a previous page handed out, or a
// filter out of range, is an invalid argument.
func (s *authService) ListAuditEntries(ctx context.Context, req *connect.Request[identityv1.ListAuditEntriesRequest]) (*connect.Response[identityv1.ListAuditEntriesResponse], error) {
	m := req.Msg
	f, err := auditFilter(m)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	page, err := s.auth.AuditEntries(ctx, m.GetOrigin(), m.GetSessionToken(), &f)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &identityv1.ListAuditEntriesResponse{Entries: make([]*identityv1.AuditEntry, 0, len(page.Entries))}
	for i := range page.Entries {
		out.Entries = append(out.Entries, pbAuditEntry(&page.Entries[i], page.Persons))
	}
	if page.Next != 0 {
		out.NextPageToken = strconv.FormatInt(page.Next, 10)
	}
	return connect.NewResponse(out), nil
}

// auditFilter converts a request to the filter it asks for; an unknown enum
// value or a bad page token wraps app.ErrInvalidRequest.
func auditFilter(m *identityv1.ListAuditEntriesRequest) (identity.AuditFilter, error) {
	f := identity.AuditFilter{ActorID: m.GetActorId(), Limit: int(m.GetPageSize())}
	var ok bool
	if m.GetActorKind() != identityv1.AuditActorKind_AUDIT_ACTOR_KIND_UNSPECIFIED {
		if f.ActorKind, ok = domainActorKinds()[m.GetActorKind()]; !ok {
			return identity.AuditFilter{}, fmt.Errorf("%w: unknown actor kind", app.ErrInvalidRequest)
		}
	}
	if m.GetAction() != identityv1.AuditAction_AUDIT_ACTION_UNSPECIFIED {
		if f.Action, ok = domainActions()[m.GetAction()]; !ok {
			return identity.AuditFilter{}, fmt.Errorf("%w: unknown action", app.ErrInvalidRequest)
		}
	}
	if err := auditOutcomeAndActions(m, &f); err != nil {
		return identity.AuditFilter{}, err
	}
	if m.GetFrom() != nil {
		f.From = m.GetFrom().AsTime()
	}
	if m.GetTo() != nil {
		f.To = m.GetTo().AsTime()
	}
	var err error
	if f.Before, err = auditCursor(m.GetPageToken()); err != nil {
		return identity.AuditFilter{}, err
	}
	return f, nil
}

// auditCursor reads the cursor a page token carries: zero for none, otherwise
// a positive number; anything else wraps app.ErrInvalidRequest.
func auditCursor(tok string) (int64, error) {
	if tok == "" {
		return 0, nil
	}
	before, err := strconv.ParseInt(tok, 10, 64)
	if err != nil || before <= 0 {
		return 0, fmt.Errorf("%w: bad page token", app.ErrInvalidRequest)
	}
	return before, nil
}

// auditOutcomeAndActions sets f's outcome and its set of actions from m; an
// unknown or unspecified member of the set wraps app.ErrInvalidRequest.
func auditOutcomeAndActions(m *identityv1.ListAuditEntriesRequest, f *identity.AuditFilter) error {
	if m.GetOutcome() != identityv1.AuditOutcome_AUDIT_OUTCOME_UNSPECIFIED {
		outcome, ok := invert(pbOutcomes())[m.GetOutcome()]
		if !ok {
			return fmt.Errorf("%w: unknown outcome", app.ErrInvalidRequest)
		}
		f.Outcome = outcome
	}
	for _, a := range m.GetActions() {
		action, ok := domainActions()[a]
		if !ok {
			return fmt.Errorf("%w: unknown action", app.ErrInvalidRequest)
		}
		f.Actions = append(f.Actions, action)
	}
	return nil
}

// The audit enums, from the domain to the contract, as functions: the package
// holds no global state. A value the domain has and a table lacks fails the
// contract test of this package.

// pbActorKinds is the contract value of each domain value.
func pbActorKinds() map[identity.ActorKind]identityv1.AuditActorKind {
	return map[identity.ActorKind]identityv1.AuditActorKind{
		identity.ActorOwner:          identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER,
		identity.ActorViewer:         identityv1.AuditActorKind_AUDIT_ACTOR_KIND_VIEWER,
		identity.ActorEnrollmentCode: identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ENROLLMENT_CODE,
		identity.ActorSystem:         identityv1.AuditActorKind_AUDIT_ACTOR_KIND_SYSTEM,
		identity.ActorAnonymous:      identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ANONYMOUS,
	}
}

// pbActions is the contract value of each domain value.
func pbActions() map[identity.AuditAction]identityv1.AuditAction {
	return map[identity.AuditAction]identityv1.AuditAction{
		identity.ActionSignIn:                  identityv1.AuditAction_AUDIT_ACTION_SIGN_IN,
		identity.ActionSignOut:                 identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT,
		identity.ActionSignOutEverywhere:       identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT_EVERYWHERE,
		identity.ActionPasskeyAdded:            identityv1.AuditAction_AUDIT_ACTION_PASSKEY_ADDED,
		identity.ActionPasskeyRemoved:          identityv1.AuditAction_AUDIT_ACTION_PASSKEY_REMOVED,
		identity.ActionEnrollmentLinkIssued:    identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_ISSUED,
		identity.ActionEnrollmentLinkUsed:      identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_USED,
		identity.ActionEnrollmentLinkCancelled: identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_CANCELLED,
		identity.ActionScriptTokenCreated:      identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_CREATED,
		identity.ActionScriptTokenRevoked:      identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_REVOKED,
		identity.ActionGrantCreated:            identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED,
		identity.ActionGrantRevoked:            identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED,
		identity.ActionModelCall:               identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL,
		identity.ActionModelCallResult:         identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL_RESULT,
		identity.ActionPersonCreated:           identityv1.AuditAction_AUDIT_ACTION_PERSON_CREATED,
		identity.ActionPersonRevoked:           identityv1.AuditAction_AUDIT_ACTION_PERSON_REVOKED,

		identity.ActionMaintainerRun:               identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_RUN,
		identity.ActionMaintainerProposalDropped:   identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_DROPPED,
		identity.ActionMaintainerProposalApplied:   identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_APPLIED,
		identity.ActionMaintainerProposalDismissed: identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_DISMISSED,
	}
}

// pbOutcomes is the contract value of each domain value.
func pbOutcomes() map[identity.AuditOutcome]identityv1.AuditOutcome {
	return map[identity.AuditOutcome]identityv1.AuditOutcome{
		identity.OutcomeOK:      identityv1.AuditOutcome_AUDIT_OUTCOME_OK,
		identity.OutcomeRefused: identityv1.AuditOutcome_AUDIT_OUTCOME_REFUSED,
		identity.OutcomeFailed:  identityv1.AuditOutcome_AUDIT_OUTCOME_FAILED,
		identity.OutcomeSent:    identityv1.AuditOutcome_AUDIT_OUTCOME_SENT,
	}
}

// pbRefKinds is the contract value of each domain value.
func pbRefKinds() map[identity.RefKind]identityv1.AuditRefKind {
	return map[identity.RefKind]identityv1.AuditRefKind{
		identity.RefSession:        identityv1.AuditRefKind_AUDIT_REF_KIND_SESSION,
		identity.RefPasskey:        identityv1.AuditRefKind_AUDIT_REF_KIND_PASSKEY,
		identity.RefEnrollmentCode: identityv1.AuditRefKind_AUDIT_REF_KIND_ENROLLMENT_CODE,
		identity.RefScriptToken:    identityv1.AuditRefKind_AUDIT_REF_KIND_SCRIPT_TOKEN,
		identity.RefPerson:         identityv1.AuditRefKind_AUDIT_REF_KIND_PERSON,
		identity.RefViewPlugin:     identityv1.AuditRefKind_AUDIT_REF_KIND_VIEW_PLUGIN,
		identity.RefAccount:        identityv1.AuditRefKind_AUDIT_REF_KIND_ACCOUNT,
		identity.RefSleeve:         identityv1.AuditRefKind_AUDIT_REF_KIND_SLEEVE,
		identity.RefHousehold:      identityv1.AuditRefKind_AUDIT_REF_KIND_HOUSEHOLD,
		identity.RefModelRole:      identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ROLE,
		identity.RefModel:          identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL,
		identity.RefModelCall:      identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_CALL,
		identity.RefModelError:     identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ERROR,
		identity.RefMaintainerRun:  identityv1.AuditRefKind_AUDIT_REF_KIND_MAINTAINER_RUN,
	}
}

// domainActorKinds is pbActorKinds the other way.
func domainActorKinds() map[identityv1.AuditActorKind]identity.ActorKind {
	return invert(pbActorKinds())
}

// domainActions is pbActions the other way.
func domainActions() map[identityv1.AuditAction]identity.AuditAction { return invert(pbActions()) }

// invert returns m with its values as keys.
func invert[K, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// pbAuditEntry converts an entry. Every field is a closed value, a time or
// an ID.
func pbAuditEntry(e *identity.AuditEntry, persons map[identity.SessionID]string) *identityv1.AuditEntry {
	actor := &identityv1.AuditActor{Kind: pbActorKinds()[e.Actor.Kind], Id: e.Actor.ID}
	if e.Actor.Kind == identity.ActorViewer {
		actor.PersonId = persons[identity.SessionID(e.Actor.ID)]
	}
	out := &identityv1.AuditEntry{
		Seq: e.Seq, At: timestamppb.New(e.At), Actor: actor,
		Action: pbActions()[e.Action], Outcome: pbOutcomes()[e.Outcome], Refs: make([]*identityv1.AuditRef, 0, len(e.Refs)),
	}
	for _, r := range e.Refs {
		out.Refs = append(out.Refs, &identityv1.AuditRef{Kind: pbRefKinds()[r.Kind], Id: r.ID})
	}
	return out
}

// pbSession converts a session; it never carries the token.
func pbSession(s *identity.Session) *identityv1.Session {
	scope := identityv1.SessionScope_SESSION_SCOPE_UNSPECIFIED
	switch s.Scope {
	case identity.SessionScopeRead:
		scope = identityv1.SessionScope_SESSION_SCOPE_READ
	case identity.SessionScopeOwner:
		scope = identityv1.SessionScope_SESSION_SCOPE_OWNER
	}
	auth := identityv1.Authenticator_AUTHENTICATOR_UNSPECIFIED
	switch s.Authenticator {
	case identity.AuthenticatorPasskey:
		auth = identityv1.Authenticator_AUTHENTICATOR_PASSKEY
	case identity.AuthenticatorDemo:
		auth = identityv1.Authenticator_AUTHENTICATOR_DEMO
	}
	out := &identityv1.Session{
		Id: string(s.ID), Scope: scope, Authenticator: auth, RpId: s.RPID,
		CreatedAt: timestamppb.New(s.CreatedAt), LastSeenAt: timestamppb.New(s.LastSeenAt),
		IdleExpiresAt: timestamppb.New(s.IdleExpiresAt()), ExpiresAt: timestamppb.New(s.ExpiresAt), PersonId: s.Person(),
	}
	if !s.PasskeyCheckedAt.IsZero() {
		out.PasskeyCheckedAt = timestamppb.New(s.PasskeyCheckedAt)
	}
	return out
}

// pbPasskey converts a passkey; it never carries the public key.
func pbPasskey(p *identity.Passkey) *identityv1.Passkey {
	out := &identityv1.Passkey{
		Id: p.ID(), Name: p.Name, RpId: p.RPID, CreatedAt: timestamppb.New(p.CreatedAt),
		Transports: p.Transports, BackedUp: p.BackupState,
	}
	if !p.LastUsedAt.IsZero() {
		out.LastUsedAt = timestamppb.New(p.LastUsedAt)
	}
	return out
}

// pbScriptToken converts a script token with its status at now.
func pbScriptToken(t *identity.ScriptToken, now time.Time) *identityv1.ScriptToken {
	status := identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_ACTIVE
	switch t.Status(now) {
	case "revoked":
		status = identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_REVOKED
	case "expired":
		status = identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_EXPIRED
	}
	out := &identityv1.ScriptToken{
		Id: string(t.ID), Name: t.Name, Scope: t.Scope.String(),
		CreatedAt: timestamppb.New(t.CreatedAt), ExpiresAt: timestamppb.New(t.ExpiresAt), Status: status,
		McpAllowed: t.MCPAllowed,
	}
	if !t.RevokedAt.IsZero() {
		out.RevokedAt = timestamppb.New(t.RevokedAt)
	}
	return out
}

// fail maps err for the response.
func (s *authService) fail(ctx context.Context, err error) error {
	return toConnectError(ctx, s.log, err)
}
