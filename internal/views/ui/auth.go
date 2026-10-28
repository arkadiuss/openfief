package ui

import (
	"encoding/json"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// maxJSONBody bounds a WebAuthn request body: a credential's JSON is a few
// kilobytes.
const maxJSONBody = 64 << 10

// loginData is what the login page shows.
type loginData struct {
	// Passkey is whether a passkey exists for this host.
	Passkey bool
	// Demo is whether the core offers read-only demo sessions here.
	Demo bool
	// Unavailable is set when the core could not say which ways in exist.
	Unavailable bool
	// NeedsHTTPS is set on plain HTTP away from localhost, where the
	// session cookie can't be kept.
	NeedsHTTPS bool
}

// loginPage shows the ways in the core offers at this origin: the passkey
// button, and "View demo" only when the core says demo sessions work. It
// still renders when the core can't answer, saying so.
func (s *server) loginPage(w http.ResponseWriter, r *http.Request) {
	origin := requestOrigin(r)
	data := loginData{NeedsHTTPS: strings.HasPrefix(origin, "http://") && !localHTTP(origin)}
	methods, err := s.auth.GetLoginMethods(r.Context(), connect.NewRequest(&identityv1.GetLoginMethodsRequest{Origin: origin}))
	if err != nil {
		s.log.WarnContext(r.Context(), "call to the core failed", "method", "GetLoginMethods", "code", connect.CodeOf(err).String())
		data.Unavailable = true
	} else {
		data.Passkey = methods.Msg.GetPasskeyRegistered()
		data.Demo = methods.Msg.GetDemoSessions()
	}
	s.show(w, r, http.StatusOK, "login", page{Title: "Sign in", Data: data})
}

// ceremonyStart is the JSON a WebAuthn ceremony starts with: its ID and the
// options in the WebAuthn Level 3 JSON form, as the core sent them.
type ceremonyStart struct {
	CeremonyID string `json:"ceremonyId"`
	Options    string `json:"options"`
}

// ceremonyEnd is the JSON the browser finishes a ceremony with: the
// credential's toJSON() as text, and for a new passkey its name.
type ceremonyEnd struct {
	CeremonyID string `json:"ceremonyId"`
	Credential string `json:"credential"`
	Name       string `json:"name"`
}

// registerStart is the JSON a registration starts with: the enrollment
// code from the link's fragment, or nothing for an owner's session.
type registerStart struct {
	Code string `json:"code"`
}

// loginBegin starts a passkey login (BeginLogin).
func (s *server) loginBegin(w http.ResponseWriter, r *http.Request) {
	resp, err := s.auth.BeginLogin(r.Context(), connect.NewRequest(&identityv1.BeginLoginRequest{Origin: requestOrigin(r)}))
	if err != nil {
		s.jsonFailed(w, r, "BeginLogin", err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, ceremonyStart{CeremonyID: resp.Msg.GetCeremonyId(), Options: resp.Msg.GetOptionsJson()})
}

// loginFinish finishes a passkey login (FinishLogin) and keeps the new
// owner-scoped session in the cookie.
func (s *server) loginFinish(w http.ResponseWriter, r *http.Request) {
	var in ceremonyEnd
	if !s.readJSON(w, r, &in) {
		return
	}
	resp, err := s.auth.FinishLogin(r.Context(), connect.NewRequest(&identityv1.FinishLoginRequest{
		Origin: requestOrigin(r), CeremonyId: in.CeremonyID, CredentialJson: in.Credential,
	}))
	if err != nil {
		s.jsonFailed(w, r, "FinishLogin", err)
		return
	}
	setSession(w, resp.Msg.GetSessionToken(), resp.Msg.GetSession())
	s.jsonRedirect(w, r, "/")
}

// demoLogin starts a read-only demo session (CreateDemoSession) and opens
// the overview.
func (s *server) demoLogin(w http.ResponseWriter, r *http.Request) {
	resp, err := s.auth.CreateDemoSession(r.Context(), connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: requestOrigin(r)}))
	if err != nil {
		s.coreFailed(w, r, "CreateDemoSession", err)
		return
	}
	setSession(w, resp.Msg.GetSessionToken(), resp.Msg.GetSession())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// enrollPage serves the page that creates a passkey from an enrollment
// link (D-60). The code stays in the link's fragment, which the browser
// never sends: the page's script reads it and posts it to /enroll/begin.
func (s *server) enrollPage(w http.ResponseWriter, r *http.Request) {
	s.show(w, r, http.StatusOK, "enroll", page{Title: "Create a passkey"})
}

// registerBegin starts creating a passkey on this host (BeginRegistration):
// with the enrollment code when the browser sends one, otherwise with the
// owner's session.
func (s *server) registerBegin(w http.ResponseWriter, r *http.Request) {
	var in registerStart
	if !s.readJSON(w, r, &in) {
		return
	}
	req := &identityv1.BeginRegistrationRequest{Origin: requestOrigin(r)}
	switch {
	case in.Code != "":
		req.Authorization = &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: in.Code}
	case sessionToken(r) != "":
		req.Authorization = &identityv1.BeginRegistrationRequest_SessionToken{SessionToken: sessionToken(r)}
	default:
		s.jsonError(w, r, http.StatusBadRequest, "Open the enrollment link again: it has no code.")
		return
	}
	resp, err := s.auth.BeginRegistration(r.Context(), connect.NewRequest(req))
	if err != nil {
		s.jsonFailed(w, r, "BeginRegistration", err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, ceremonyStart{CeremonyID: resp.Msg.GetCeremonyId(), Options: resp.Msg.GetOptionsJson()})
}

// registerFinish stores the new passkey (FinishRegistration). A
// registration from an enrollment code also logs the owner in, so its
// session goes in the cookie and the browser opens the overview; one from
// a session goes back to the settings.
func (s *server) registerFinish(w http.ResponseWriter, r *http.Request) {
	var in ceremonyEnd
	if !s.readJSON(w, r, &in) {
		return
	}
	resp, err := s.auth.FinishRegistration(r.Context(), connect.NewRequest(&identityv1.FinishRegistrationRequest{
		Origin: requestOrigin(r), CeremonyId: in.CeremonyID, CredentialJson: in.Credential, Name: in.Name,
	}))
	if err != nil {
		s.jsonFailed(w, r, "FinishRegistration", err)
		return
	}
	if token := resp.Msg.GetSessionToken(); token != "" {
		setSession(w, token, resp.Msg.GetSession())
		s.jsonRedirect(w, r, "/")
		return
	}
	s.jsonRedirect(w, r, "/settings")
}

// logout revokes the session at the core (Logout), drops the cookie and
// opens the login page. The cookie goes even when the core can't be
// reached, so the browser is logged out either way.
func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if token := sessionToken(r); token != "" {
		_, err := s.auth.Logout(r.Context(), connect.NewRequest(&identityv1.LogoutRequest{Origin: requestOrigin(r), SessionToken: token}))
		if err != nil && connect.CodeOf(err) != connect.CodeUnauthenticated {
			s.log.WarnContext(r.Context(), "call to the core failed", "method", "Logout", "code", connect.CodeOf(err).String())
		}
	}
	clearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// readJSON decodes a small JSON body into v, answering 400 when it can't.
func (s *server) readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		s.jsonError(w, r, http.StatusBadRequest, "malformed request")
		return false
	}
	return true
}

// writeJSON writes v as a JSON response with status.
func (s *server) writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.DebugContext(r.Context(), "write JSON response", "error", err.Error())
	}
}

// jsonRedirect tells the page's script where to go next.
func (s *server) jsonRedirect(w http.ResponseWriter, r *http.Request, path string) {
	s.writeJSON(w, r, http.StatusOK, map[string]string{"redirect": path})
}

// jsonError answers the page's script with status and a message for the
// owner.
func (s *server) jsonError(w http.ResponseWriter, r *http.Request, status int, message string) {
	s.writeJSON(w, r, status, map[string]string{"error": message})
}

// jsonFailed answers a failed WebAuthn call as JSON with a message for the
// owner: the core's own message for a refusal it explains, a generic one
// for an outage, which is logged by method and code only.
func (s *server) jsonFailed(w http.ResponseWriter, r *http.Request, method string, err error) {
	code := connect.CodeOf(err)
	if code == connect.CodeUnauthenticated {
		s.jsonError(w, r, http.StatusForbidden, "Not allowed: "+connectMessage(err))
		return
	}
	if status, refused := refusalStatus(code); refused {
		s.jsonError(w, r, status, "Refused: "+connectMessage(err))
		return
	}
	s.log.WarnContext(r.Context(), "call to the core failed", "method", method, "code", code.String())
	s.jsonError(w, r, http.StatusBadGateway, "The OpenFief core could not answer. Try again in a moment.")
}
