package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	"github.com/arkadiuss/openfief/internal/plugins"
	pluginsapp "github.com/arkadiuss/openfief/internal/plugins/app"
	pluginskube "github.com/arkadiuss/openfief/internal/plugins/kube"
)

// rpName is the relying party name browsers show in passkey prompts.
const rpName = "OpenFief"

// authConfig is serve's authentication configuration.
type authConfig struct {
	// Origins is the allow-list of origins the owner's browser may use;
	// each one's host is a passkey host (D-60).
	Origins identity.Origins
	// Sessions is how long sessions last.
	Sessions identity.SessionPolicy
	// DemoSessions turns the demo authenticator on; it also needs the demo
	// marker in the database.
	DemoSessions bool
}

// stringsFlag is a repeatable string flag.
type stringsFlag []string

// String returns the values joined by commas.
func (s *stringsFlag) String() string { return strings.Join(*s, ",") }

// Set adds one value.
func (s *stringsFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// newAuth returns the auth use cases over pool, as cfg configures them.
func newAuth(pool *pgxpool.Pool, cfg *authConfig, now func() time.Time) (*identityapp.Auth, error) {
	store := identitypg.New(pool)
	auth, err := identityapp.NewAuth(store, webauthn.New(rpName), identityapp.NewTokens(store, now, nil), demoMarker{pool: pool},
		identityapp.AuthConfig{Origins: cfg.Origins, Sessions: cfg.Sessions, DemoSessions: cfg.DemoSessions}, now, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	return auth, nil
}

// withAuthService returns api with the AuthService mounted next to it:
// AuthService calls go to the auth handler, every other path to api.
// Responses carry session tokens and secrets, so nothing may store them.
func withAuthService(api http.Handler, auth *identityapp.Auth, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := identityconnect.New(auth, callers, log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", api)
	return mux
}

// callers is who may call the AuthService: view plugins whose token the
// reviewer accepts. Without a reviewer, outside a cluster, it refuses every
// caller: no build has a fixed token or another way around authentication
// (D-84).
type callers struct {
	reviewer pluginsapp.TokenReviewer
}

// VerifyCaller implements identityconnect.Callers.
func (c callers) VerifyCaller(ctx context.Context, token string) (string, error) {
	if c.reviewer == nil {
		return "", fmt.Errorf("%w: not in a Kubernetes cluster", identityconnect.ErrCallerRefused)
	}
	return kubeCallers(c).VerifyCaller(ctx, token)
}

// newCallers returns who may call the AuthService: in cluster cl, view
// plugins proven by their projected ServiceAccount token (audience
// openfief-core, D-27) through TokenReview, cached; outside one, no one
// (D-84).
func newCallers(cl *cluster, now func() time.Time, log *slog.Logger) (callers, error) {
	if cl == nil {
		log.Info("not in a Kubernetes cluster: the AuthService refuses every caller")
		return callers{}, nil
	}
	view := &viewReviewer{
		next:    pluginskube.NewTokenReviewer(cl.client, plugins.CoreAudience),
		plugins: cl.dynamic.Resource(ingestkube.PluginResource()).Namespace(cl.namespace),
	}
	cached, err := pluginsapp.NewCachingTokenReviewer(view, pluginsapp.DefaultTokenCacheTTL, now)
	if err != nil {
		return callers{}, fmt.Errorf("auth callers: %w", err)
	}
	return callers{reviewer: cached}, nil
}

// kubeCallers accepts view plugins whose token the reviewer accepts.
type kubeCallers struct {
	reviewer pluginsapp.TokenReviewer
}

// VerifyCaller implements identityconnect.Callers.
func (k kubeCallers) VerifyCaller(ctx context.Context, token string) (string, error) {
	username, err := k.reviewer.ReviewToken(ctx, token)
	if errors.Is(err, pluginsapp.ErrUnauthenticated) {
		return "", fmt.Errorf("%w: %w", identityconnect.ErrCallerRefused, err)
	}
	if err != nil {
		return "", fmt.Errorf("reviewing a caller token: %w", err)
	}
	name, ok := plugins.PluginForServiceAccount(username)
	if !ok {
		return "", fmt.Errorf("%w: not a plugin", identityconnect.ErrCallerRefused)
	}
	return name, nil
}

// viewReviewer reviews a projected token with audience openfief-core and
// accepts it only when it is a plugin's whose Plugin resource is not being
// deleted and has view among its effective capabilities.
type viewReviewer struct {
	next    pluginsapp.TokenReviewer
	plugins dynamic.ResourceInterface
}

// ReviewToken implements pluginsapp.TokenReviewer.
func (v *viewReviewer) ReviewToken(ctx context.Context, token string) (string, error) {
	username, err := v.next.ReviewToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("reviewing: %w", err)
	}
	name, ok := plugins.PluginForServiceAccount(username)
	if !ok {
		return "", fmt.Errorf("%w: not a plugin", pluginsapp.ErrUnauthenticated)
	}
	u, err := v.plugins.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return "", fmt.Errorf("%w: plugin %s is not installed", pluginsapp.ErrUnauthenticated, name)
	}
	if err != nil {
		return "", fmt.Errorf("reading plugin %s: %w", name, err)
	}
	if !isView(u) {
		return "", fmt.Errorf("%w: plugin %s is not a view plugin", pluginsapp.ErrUnauthenticated, name)
	}
	return username, nil
}

// isView reports whether a Plugin resource runs a version that may act as
// a view: not being deleted, with view among its effective capabilities.
func isView(u *unstructured.Unstructured) bool {
	if u.GetDeletionTimestamp() != nil {
		return false
	}
	caps, _, err := unstructured.NestedStringSlice(u.Object, "status", "effective", "capabilities")
	return err == nil && slices.Contains(caps, string(plugins.CapabilityView))
}

// demoMarker reads the demo marker that cmd/demo-seed writes: a row
// demo = true in demo.marker. A database without the table is no demo.
type demoMarker struct {
	pool *pgxpool.Pool
}

// DemoMarked implements identityapp.DemoMarker.
func (d demoMarker) DemoMarked(ctx context.Context) (bool, error) {
	var marked bool
	err := d.pool.QueryRow(ctx, `SELECT to_regclass('demo.marker') IS NOT NULL`).Scan(&marked)
	if err != nil {
		return false, fmt.Errorf("looking for the demo marker: %w", err)
	}
	if !marked {
		return false, nil
	}
	if err := d.pool.QueryRow(ctx, `SELECT EXISTS (SELECT FROM demo.marker WHERE demo)`).Scan(&marked); err != nil {
		return false, fmt.Errorf("reading the demo marker: %w", err)
	}
	return marked, nil
}

// enrollLink runs core enroll-link: it stores a one-time enrollment code
// for origin, lasting ttl, and writes the link and its expiry to out. The
// link is the only place the code is shown.
func enrollLink(ctx context.Context, pool *pgxpool.Pool, origin identity.Origin, ttl time.Duration, out io.Writer, now time.Time) error {
	link, err := identityapp.IssueEnrollmentLink(ctx, identitypg.New(pool), nil, now, origin, ttl)
	if err != nil {
		return fmt.Errorf("enroll-link: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\nopen it within %s (until %s) to create a passkey on %s; it works once\n",
		link.URL, ttl, link.ExpiresAt.Format(time.RFC3339), origin.RPID())
	return writeErr(err)
}

// originForHost returns the origin the enroll-link host means: an origin
// as given when it has a scheme, else http:// for localhost and https://
// for any other host.
func originForHost(host string) (identity.Origin, error) {
	text := host
	if !strings.Contains(host, "://") {
		scheme := "https://"
		if host == "localhost" || strings.HasPrefix(host, "localhost:") {
			scheme = "http://"
		}
		text = scheme + host
	}
	o, err := identity.ParseOrigin(text)
	if err != nil {
		return identity.Origin{}, fmt.Errorf("--host: %w", err)
	}
	return o, nil
}
