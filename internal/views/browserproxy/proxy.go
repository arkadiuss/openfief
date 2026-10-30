// Package browserproxy is the views context's HTTP adapter for D-81: the
// core is the only entry point for browsers, and it reverse-proxies each
// browser request to the one view plugin that holds a browser path (D-114),
// such as the built-in UI, with the core's own token for that instance
// (D-59). It is the only way a browser reaches a plugin: kubectl
// port-forward and the owner's own route (D-28) reach the core's browser
// port alone.
//
// Toward the plugin it sets what the plugin may trust, and only that: the
// browser's Host, the scheme the browser used in X-Forwarded-Proto, and the
// core's token in Authorization. Whatever a client sent in those headers,
// or in any other X-Forwarded-* or Forwarded header, is dropped.
package browserproxy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// The transport's limits toward view plugins.
const (
	// dialTimeout bounds connecting to a view plugin.
	dialTimeout = 10 * time.Second
	// idleConnsPerPlugin is how many idle connections to one plugin stay open.
	idleConnsPerPlugin = 8
	// idleConnTimeout closes an idle connection to a plugin.
	idleConnTimeout = 90 * time.Second
)

// The schemes a browser may use.
const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Proxy is the handler that sends browsers to view plugins.
type Proxy struct {
	door  *app.FrontDoor
	proxy *httputil.ReverseProxy
	log   *slog.Logger
}

// targetKey carries the chosen target from ServeHTTP to the rewrite.
type targetKey struct{}

// New returns the proxy over door, logging failed requests to log (nil
// discards). Calls to plugins never go through an HTTP proxy.
func New(door *app.FrontDoor, log *slog.Logger) *Proxy {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
		MaxIdleConnsPerHost:   idleConnsPerPlugin,
		IdleConnTimeout:       idleConnTimeout,
		ExpectContinueTimeout: time.Second,
	}
	p := &Proxy{door: door, log: log}
	p.proxy = &httputil.ReverseProxy{
		Rewrite:      rewrite,
		Transport:    transport,
		ErrorHandler: p.unreachable,
	}
	return p
}

// ServeHTTP sends r to the one view plugin that holds a browser path
// (D-114), when its path is under it. A path no plugin serves is 404; more
// than one view holding a browser path, or routes not read yet, 503; a
// plugin without an address, a token or an answer, 502.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := p.door.Target(r.Context(), r.URL.Path)
	if err != nil {
		p.refuse(w, r, err)
		return
	}
	p.proxy.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), targetKey{}, target)))
}

// refuse answers a request the front door can't route.
func (p *Proxy) refuse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, views.ErrNoBrowserRoute):
		plain(w, http.StatusNotFound, "not found")
		return
	case errors.Is(err, views.ErrSeveralBrowserViews), errors.Is(err, app.ErrViewsUnknown):
		w.Header().Set("Retry-After", "5")
		plain(w, http.StatusServiceUnavailable, "no view plugin can serve this path now")
	default:
		plain(w, http.StatusBadGateway, "the view plugin can't be reached")
	}
	p.log.Warn("browser request not routed", "path", r.URL.Path, "error", err.Error())
}

// unreachable answers a request whose plugin did not answer.
func (p *Proxy) unreachable(w http.ResponseWriter, r *http.Request, err error) {
	instance := ""
	if t, ok := r.Context().Value(targetKey{}).(app.BrowserTarget); ok {
		instance = t.Instance
	}
	p.log.Warn("view plugin did not answer", "instance", instance, "path", r.URL.Path, "error", err.Error())
	plain(w, http.StatusBadGateway, "the view plugin can't be reached")
}

// rewrite points the outgoing request at the chosen plugin. Rewrite has
// already dropped the hop-by-hop headers and every X-Forwarded-* header the
// client sent; this also drops Forwarded, keeps the browser's Host, sets
// X-Forwarded-Proto and X-Forwarded-Host to what the core saw, and replaces
// Authorization with the core's token for the plugin.
func rewrite(pr *httputil.ProxyRequest) {
	target, ok := pr.In.Context().Value(targetKey{}).(app.BrowserTarget)
	if !ok {
		return // ServeHTTP always sets it
	}
	base, err := url.Parse(target.BaseURL)
	if err != nil {
		return // the transport then fails on the empty host
	}
	pr.SetURL(base)
	pr.Out.Host = pr.In.Host
	h := pr.Out.Header
	h.Del("Forwarded")
	h.Set("X-Forwarded-Proto", BrowserScheme(pr.In))
	h.Set("X-Forwarded-Host", pr.In.Host)
	h.Set("Authorization", "Bearer "+target.Token)
}

// BrowserScheme returns the scheme the browser used to reach the core:
// https on a TLS connection, otherwise the scheme the owner's Ingress in
// front reports in X-Forwarded-Proto when it is http or https, otherwise
// http, as through kubectl port-forward. A forged value gains nothing: the
// core checks every origin against its allow-list, and passkeys and
// sessions work only on the host they were made on.
func BrowserScheme(r *http.Request) string {
	if r.TLS != nil {
		return schemeHTTPS
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == schemeHTTPS || p == schemeHTTP {
		return p
	}
	return schemeHTTP
}

// plain writes a short plain-text answer that nothing may cache.
func plain(w http.ResponseWriter, status int, body string) {
	h := w.Header()
	h.Set("Content-Type", "text/plain; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, body); err != nil {
		return // the browser went away
	}
}
