package main

import (
	"log/slog"
	"net/http"
	"time"

	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	"github.com/arkadiuss/openfief/internal/plugins"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
	viewskube "github.com/arkadiuss/openfief/internal/views/kube"
)

// withFrontDoor returns api, the browser port's read API, behind the core's
// front door (D-81): the paths the core serves itself (plugins.CoreServesPath:
// the read API, the paths of its Connect services and the health endpoints)
// go to api, which answers 404 for the services only plugins call (D-115),
// and every other path to browser, which proxies it to the one view plugin
// that holds a browser path (D-114).
func withFrontDoor(api, browser http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if plugins.CoreServesPath(r.URL.Path) {
			api.ServeHTTP(w, r)
			return
		}
		browser.ServeHTTP(w, r)
	})
}

// newFrontDoor returns, in cluster cl, the view plugins that serve
// browsers, to start, and the proxy to them, which presents the core's own
// token for each instance (D-59) as serviceAccount.
func newFrontDoor(cl *cluster, serviceAccount string, now func() time.Time, log *slog.Logger) (*viewskube.Browser, http.Handler) {
	browser := viewskube.NewBrowser(cl.dynamic, cl.client, cl.namespace, discoveryResync)
	tokens := identitykube.NewCoreTokens(cl.client, cl.namespace, serviceAccount, now)
	return browser, browserproxy.New(&viewsapp.FrontDoor{Views: browser, Tokens: tokens}, log)
}
