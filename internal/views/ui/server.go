package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// contentSecurityPolicy lets the page load scripts, styles, images and
// fetches from the UI itself only, nothing inline and nothing from another
// host; no page may frame it, and forms post only to it.
const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; " +
	"connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"

// hstsValue is the Strict-Transport-Security header on HTTPS origins.
const hstsValue = "max-age=31536000"

// ErrMissingDependency reports a Config without one of its clients.
var ErrMissingDependency = errors.New("ui: config needs AuthService, ModelPortfolioService, BondLotService and ReadService clients")

// Config is what the UI needs.
type Config struct {
	// Auth calls the core's AuthService with the plugin's own token in
	// every request's Authorization header (D-27).
	Auth identityv1connect.AuthServiceClient
	// Model calls the core's ModelPortfolioService (D-52) like Auth: with
	// the plugin's own token, the owner's session in each request.
	Model portfoliov1connect.ModelPortfolioServiceClient
	// BondLots calls the core's BondLotService (D-9, D-49) like Auth: with
	// the plugin's own token, the owner's session in each request.
	BondLots ingestv1connect.BondLotServiceClient
	// Read calls the core's public read API (D-6). It must not set
	// Authorization itself: every call carries the session token.
	Read viewsv1connect.ReadServiceClient
	// Today returns today's date; nil is the UTC date of the clock.
	Today func() kernel.Date
	// Log gets one line per failed call to the core, with no token or
	// amount in it; nil discards.
	Log *slog.Logger
}

// server holds the UI's handlers.
type server struct {
	auth     identityv1connect.AuthServiceClient
	model    portfoliov1connect.ModelPortfolioServiceClient
	bondLots ingestv1connect.BondLotServiceClient
	read     viewsv1connect.ReadServiceClient
	pages    *pages
	today    func() kernel.Date
	log      *slog.Logger
}

// New returns the UI's HTTP handler: its pages, their static files, the
// security headers on every response and the same-origin check on every
// state-changing request.
func New(cfg *Config) (http.Handler, error) {
	if cfg.Auth == nil || cfg.Model == nil || cfg.BondLots == nil || cfg.Read == nil {
		return nil, ErrMissingDependency
	}
	p, err := parsePages()
	if err != nil {
		return nil, err
	}
	s := &server{auth: cfg.Auth, model: cfg.Model, bondLots: cfg.BondLots, read: cfg.Read, pages: p, today: cfg.Today, log: cfg.Log}
	if s.today == nil {
		s.today = utcToday
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	static, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, fmt.Errorf("ui: static files: %w", err)
	}
	return securityHeaders(sameOrigin(s.routes(static))), nil
}

// routes maps every path to its handler.
func (s *server) routes(static fs.FS) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", staticFiles(static))
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /login", s.loginPage)
	mux.HandleFunc("POST /login/passkey/begin", s.loginBegin)
	mux.HandleFunc("POST /login/passkey/finish", s.loginFinish)
	mux.HandleFunc("POST /login/demo", s.demoLogin)
	mux.HandleFunc("GET /enroll", s.enrollPage)
	mux.HandleFunc("POST /enroll/begin", s.registerBegin)
	mux.HandleFunc("POST /enroll/finish", s.registerFinish)
	mux.HandleFunc("POST /logout", s.logout)
	mux.HandleFunc("GET /{$}", s.withSession(s.overview))
	mux.HandleFunc("GET /allocation", s.withSession(s.allocationFragment))
	mux.HandleFunc("GET /accounts/{id}", s.withSession(s.accountPage))
	mux.HandleFunc("GET /accounts/{id}/transactions", s.withSession(s.transactionsFragment))
	mux.HandleFunc("POST /accounts/{id}/bond-lots", s.withOwner(s.addBondLot))
	mux.HandleFunc("GET /assertions", s.withSession(s.assertionsPage))
	mux.HandleFunc("GET /price-checks", s.withSession(s.priceChecksPage))
	mux.HandleFunc("GET /plugins", s.withSession(s.pluginsPage))
	mux.HandleFunc("GET /settings", s.withOwner(s.settingsPage))
	mux.HandleFunc("POST /settings/passkeys/{id}/remove", s.withOwner(s.removePasskey))
	mux.HandleFunc("POST /settings/tokens", s.withOwner(s.createToken))
	mux.HandleFunc("POST /settings/tokens/{id}/revoke", s.withOwner(s.revokeToken))
	mux.HandleFunc("POST /settings/enrollment-links", s.withOwner(s.createEnrollmentLink))
	mux.HandleFunc("POST /settings/model", s.withOwner(s.saveModel))
	return mux
}

// staticFiles serves the embedded scripts and styles. They change only with
// the image, so browsers may keep them for an hour.
func staticFiles(static fs.FS) http.Handler {
	files := http.StripPrefix("/static/", http.FileServerFS(static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		files.ServeHTTP(w, r)
	})
}

// healthz answers 200 without a session or a call to the core, for probes.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("ok\n")); err != nil {
		return // the prober went away
	}
}

// securityHeaders sets the security headers of every response: the strict
// CSP, no sniffing, no framing, no Referer, an isolated browsing context,
// passkeys only for the UI itself, HSTS on HTTPS, and no caching of pages,
// which show the owner's record.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Permissions-Policy", "publickey-credentials-get=(self), publickey-credentials-create=(self), "+
			"camera=(), microphone=(), geolocation=()")
		if strings.HasPrefix(requestOrigin(r), schemeHTTPS+"://") {
			h.Set("Strict-Transport-Security", hstsValue)
		}
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// sameOrigin refuses, with 403, every state-changing request a browser did
// not send from the UI's own origin (CSRF). Sec-Fetch-Site decides when the
// browser sent it; otherwise Origin must equal the UI's origin. A request
// with neither is refused too: every browser that can log in sends one.
// SameSite=Strict on the session cookie is the second line of defense.
func sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !safeMethod(r.Method) && !fromOwnOrigin(r) {
			http.Error(w, "cross-origin request refused", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// safeMethod reports whether a method never changes state.
func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// fromOwnOrigin reports whether a browser sent r from the UI's own origin.
func fromOwnOrigin(r *http.Request) bool {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" {
		return site == "same-origin"
	}
	origin := r.Header.Get("Origin")
	return origin != "" && origin == requestOrigin(r)
}
