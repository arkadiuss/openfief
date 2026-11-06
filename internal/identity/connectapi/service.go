package connectapi

import (
	"context"
	"log/slog"
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
		CeremonyId: st.CeremonyID, OptionsJson: string(st.Options), ExpiresAt: timestamppb.New(st.ExpiresAt),
	}), nil
}

// FinishRegistration stores the new passkey.
func (s *authService) FinishRegistration(ctx context.Context, req *connect.Request[identityv1.FinishRegistrationRequest]) (*connect.Response[identityv1.FinishRegistrationResponse], error) {
	m := req.Msg
	r, err := s.auth.FinishRegistration(ctx, m.GetOrigin(), m.GetCeremonyId(), []byte(m.GetCredentialJson()), m.GetName())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
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
	owner, ss, err := s.auth.WhoAmI(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.WhoAmIResponse{
		Owner: &identityv1.Owner{CreatedAt: timestamppb.New(owner.CreatedAt)}, Session: pbSession(&ss),
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
	tok, secret, err := s.auth.CreateScriptToken(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetName(), m.GetTtl().AsDuration())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.CreateScriptTokenResponse{Token: pbScriptToken(&tok, time.Now()), Secret: secret}), nil
}

// ListScriptTokens returns every script token.
func (s *authService) ListScriptTokens(ctx context.Context, req *connect.Request[identityv1.ListScriptTokensRequest]) (*connect.Response[identityv1.ListScriptTokensResponse], error) {
	toks, err := s.auth.ScriptTokens(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	now := time.Now()
	out := &identityv1.ListScriptTokensResponse{Tokens: make([]*identityv1.ScriptToken, 0, len(toks))}
	for i := range toks {
		out.Tokens = append(out.Tokens, pbScriptToken(&toks[i], now))
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
		IdleExpiresAt: timestamppb.New(s.IdleExpiresAt()), ExpiresAt: timestamppb.New(s.ExpiresAt),
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
