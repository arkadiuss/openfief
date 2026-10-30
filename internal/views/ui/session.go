package ui

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// cookieName is the session cookie's name. The __Host- prefix makes the
// browser insist on Secure, Path=/ and no Domain, so no other host can set
// or read it.
const cookieName = "__Host-openfief_session"

// The schemes an origin may have.
const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// htmx's request and response headers, in Go's canonical form.
const (
	hxRequest  = "Hx-Request"
	hxRedirect = "Hx-Redirect"
)

// session is the owner's session as one request sees it: the token from
// the cookie, its scope from WhoAmI, and the origin the browser uses.
type session struct {
	Token  string
	Scope  identityv1.SessionScope
	Origin string
	// Settings are the read API's settings: the number format amounts are
	// written in (D-113) and the tax wrappers an account can have (D-76).
	Settings *viewsv1.GetSettingsResponse
}

// Owner reports whether the session may do what only the owner may: manage
// passkeys, script tokens and enrollment links. A read-scoped session, such
// as the demo's, may not, and its pages leave those actions out.
func (s session) Owner() bool { return s.Scope == identityv1.SessionScope_SESSION_SCOPE_OWNER }

// ScopeLabel names the session's scope for the page header.
func (s session) ScopeLabel() string { return scopeLabel(s.Scope) }

// requestOrigin returns the origin the browser uses for the UI: the scheme
// Traefik or another proxy reports in X-Forwarded-Proto (http without
// one), then the Host. The core checks it against its allow-list, and a
// passkey and a session work only on the host they were made on, so a
// forged value gains nothing.
func requestOrigin(r *http.Request) string {
	scheme := schemeHTTP
	if r.TLS != nil {
		scheme = schemeHTTPS
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == schemeHTTPS || p == schemeHTTP {
		scheme = p
	}
	return scheme + "://" + r.Host
}

// localHTTP reports whether origin is plain HTTP on the loopback host
// (D-60's first login through kubectl port-forward), which browsers treat
// as a secure context without TLS.
func localHTTP(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != schemeHTTP {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// sessionCookie returns the cookie holding a session token: HttpOnly, so no
// script can read it; Secure, which browsers also accept from
// http://localhost, a secure context; SameSite=Strict, so no other site's
// page sends it; Path=/; expiring with the session.
func sessionCookie(token string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// setSession stores a new session's token in the browser.
func setSession(w http.ResponseWriter, token string, s *identityv1.Session) {
	http.SetCookie(w, sessionCookie(token, s.GetExpiresAt().AsTime(), 0))
}

// clearSession removes the session cookie from the browser.
func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, sessionCookie("", time.Unix(0, 0), -1))
}

// sessionToken returns the session token the browser sent, or "".
func sessionToken(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// isHTMX reports whether htmx sent r, to swap part of a page.
func isHTMX(r *http.Request) bool { return r.Header.Get(hxRequest) == "true" }

// sessionHandler is a handler that runs with the owner's session.
type sessionHandler func(w http.ResponseWriter, r *http.Request, sess session)

// withSession runs h with the browser's session after the core confirmed
// it with WhoAmI, and with the read API's settings, whose number format
// every page and fragment h renders writes amounts in (D-113). Without a session, or with one the core no longer
// accepts, the browser goes to the login page instead.
func (s *server) withSession(h sessionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := requestOrigin(r)
		token := sessionToken(r)
		if token == "" {
			toLogin(w, r)
			return
		}
		who, err := s.auth.WhoAmI(r.Context(), connect.NewRequest(&identityv1.WhoAmIRequest{Origin: origin, SessionToken: token}))
		if err != nil {
			s.coreFailed(w, r, "WhoAmI", err)
			return
		}
		sess := session{Token: token, Scope: who.Msg.GetSession().GetScope(), Origin: origin}
		settings, err := s.read.GetSettings(r.Context(), bearer(&viewsv1.GetSettingsRequest{}, sess))
		if err != nil {
			s.coreFailed(w, r, "GetSettings", err)
			return
		}
		sess.Settings = settings.Msg
		h(w, withFormat(r, formatOf(settings.Msg.GetNumberFormat())), sess)
	}
}

// withOwner is withSession for the owner's actions: a read-scoped session
// gets 403. The core refuses such calls anyway; this keeps the page honest.
func (s *server) withOwner(h sessionHandler) http.HandlerFunc {
	return s.withSession(func(w http.ResponseWriter, r *http.Request, sess session) {
		if !sess.Owner() {
			s.errorPage(w, r, http.StatusForbidden, "Only the owner's passkey session can do this.")
			return
		}
		h(w, r, sess)
	})
}

// toLogin sends the browser to the login page: a redirect, or for an htmx
// request the HX-Redirect header, which makes htmx load the whole page.
func toLogin(w http.ResponseWriter, r *http.Request) {
	if isHTMX(r) {
		w.Header().Set(hxRedirect, "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// refusalStatus returns the status a page gets when the core refused a
// request with code, and false for any other failure, which is an outage.
func refusalStatus(code connect.Code) (int, bool) {
	status, ok := map[connect.Code]int{
		connect.CodePermissionDenied:   http.StatusForbidden,
		connect.CodeNotFound:           http.StatusNotFound,
		connect.CodeInvalidArgument:    http.StatusBadRequest,
		connect.CodeFailedPrecondition: http.StatusBadRequest,
		connect.CodeAlreadyExists:      http.StatusConflict,
		connect.CodeResourceExhausted:  http.StatusTooManyRequests,
		connect.CodeUnimplemented:      http.StatusNotImplemented,
	}[code]
	return status, ok
}

// coreFailed answers a failed call to the core: an unknown, expired or
// revoked session drops the cookie and goes to the login page (the
// AuthService contract); a refusal shows the core's reason with its
// status; anything else is 502, logged by method and code only.
func (s *server) coreFailed(w http.ResponseWriter, r *http.Request, method string, err error) {
	code := connect.CodeOf(err)
	if code == connect.CodeUnauthenticated {
		clearSession(w)
		toLogin(w, r)
		return
	}
	if status, refused := refusalStatus(code); refused {
		s.errorPage(w, r, status, "The core refused this: "+connectMessage(err))
		return
	}
	s.log.WarnContext(r.Context(), "call to the core failed", "method", method, "code", code.String())
	s.errorPage(w, r, http.StatusBadGateway, "The OpenFief core could not answer. Try again in a moment.")
}

// connectMessage is a Connect error's own message, without the code the
// error string starts with; empty for other errors.
func connectMessage(err error) string {
	if ce, ok := errors.AsType[*connect.Error](err); ok {
		return ce.Message()
	}
	return ""
}

// bearer returns req with the session token as its bearer token: the read
// API takes the owner's session like a script token (D-6).
func bearer[T any](msg *T, sess session) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+sess.Token)
	return req
}

// utcToday is the clock's date in UTC.
func utcToday() kernel.Date {
	y, m, d := time.Now().UTC().Date()
	t, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return t
}
