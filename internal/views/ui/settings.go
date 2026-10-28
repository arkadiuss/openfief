package ui

import (
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

const (
	// defaultTokenDays is a new script token's lifetime when the form
	// leaves it out: the core's default, 30 days.
	defaultTokenDays = 30
	// maxTokenDays is the longest a script token lives (D-27).
	maxTokenDays = 365
	// hoursPerDay turns days into a duration.
	hoursPerDay = 24
)

// settingsData is what the settings page shows: the owner's passkeys and
// script tokens, and once, right after it was made, a new token's secret
// or a new enrollment link.
type settingsData struct {
	Origin      string
	Passkeys    []*identityv1.Passkey
	Tokens      []*identityv1.ScriptToken
	NewToken    *identityv1.ScriptToken
	NewSecret   string
	Enrollment  *identityv1.CreateEnrollmentLinkResponse
	Error       string
	DefaultDays int
}

// settingsPage shows the owner's settings.
func (s *server) settingsPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderSettings(w, r, sess, http.StatusOK, &settingsData{})
}

// renderSettings loads the passkeys and script tokens and shows them with
// whatever the last action left in data.
func (s *server) renderSettings(w http.ResponseWriter, r *http.Request, sess session, status int, data *settingsData) {
	keys, err := s.auth.ListPasskeys(r.Context(), connect.NewRequest(&identityv1.ListPasskeysRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "ListPasskeys", err)
		return
	}
	tokens, err := s.auth.ListScriptTokens(r.Context(), connect.NewRequest(&identityv1.ListScriptTokensRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "ListScriptTokens", err)
		return
	}
	data.Origin, data.Passkeys, data.Tokens, data.DefaultDays = sess.Origin, keys.Msg.GetPasskeys(), tokens.Msg.GetTokens(), defaultTokenDays
	s.show(w, r, status, "settings", page{Title: "Settings", Session: &sess, Data: data})
}

// actionFailed shows the settings again with the core's reason when it
// refused an action, and otherwise answers like any failed call.
func (s *server) actionFailed(w http.ResponseWriter, r *http.Request, sess session, method string, err error) {
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderSettings(w, r, sess, status, &settingsData{Error: connectMessage(err)})
		return
	}
	s.coreFailed(w, r, method, err)
}

// removePasskey removes a passkey (RemovePasskey); the core refuses to
// remove the last one.
func (s *server) removePasskey(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.auth.RemovePasskey(r.Context(), connect.NewRequest(&identityv1.RemovePasskeyRequest{
		Origin: sess.Origin, SessionToken: sess.Token, PasskeyId: r.PathValue("id"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "RemovePasskey", err)
		return
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// createToken issues a script token (CreateScriptToken) and shows its
// secret this once; the core never shows it again.
func (s *server) createToken(w http.ResponseWriter, r *http.Request, sess session) {
	days, ok := tokenDays(r.PostFormValue("days"))
	if !ok {
		s.renderSettings(w, r, sess, http.StatusBadRequest, &settingsData{Error: "Days must be a whole number from 1 to 365."})
		return
	}
	resp, err := s.auth.CreateScriptToken(r.Context(), connect.NewRequest(&identityv1.CreateScriptTokenRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: r.PostFormValue("name"),
		Ttl: durationpb.New(time.Duration(days) * hoursPerDay * time.Hour),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "CreateScriptToken", err)
		return
	}
	s.renderSettings(w, r, sess, http.StatusOK, &settingsData{NewToken: resp.Msg.GetToken(), NewSecret: resp.Msg.GetSecret()})
}

// tokenDays reads a token lifetime in days: empty is the default, anything
// else a whole number from 1 to maxTokenDays.
func tokenDays(raw string) (int, bool) {
	if raw == "" {
		return defaultTokenDays, true
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 1 || days > maxTokenDays {
		return 0, false
	}
	return days, true
}

// revokeToken revokes a script token (RevokeScriptToken).
func (s *server) revokeToken(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.auth.RevokeScriptToken(r.Context(), connect.NewRequest(&identityv1.RevokeScriptTokenRequest{
		Origin: sess.Origin, SessionToken: sess.Token, TokenId: r.PathValue("id"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "RevokeScriptToken", err)
		return
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// createEnrollmentLink issues a one-time link for creating a passkey on
// another allowed origin, or on this one when the form names none (D-60),
// and shows it this once.
func (s *server) createEnrollmentLink(w http.ResponseWriter, r *http.Request, sess session) {
	resp, err := s.auth.CreateEnrollmentLink(r.Context(), connect.NewRequest(&identityv1.CreateEnrollmentLinkRequest{
		Origin: sess.Origin, SessionToken: sess.Token, TargetOrigin: r.PostFormValue("target_origin"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "CreateEnrollmentLink", err)
		return
	}
	s.renderSettings(w, r, sess, http.StatusOK, &settingsData{Enrollment: resp.Msg})
}
