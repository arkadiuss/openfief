package ui

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

// servedWorker is the service worker as served: its CONFIG and the rest of
// the script.
type servedWorker struct {
	cfg    workerConfig
	script string
}

// workerOf serves /sw.js and decodes the CONFIG it starts with.
func workerOf(t *testing.T, h *harness) servedWorker {
	t.Helper()
	rec := h.serve(t, request(http.MethodGet, workerPath, ""))
	wantStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	head, script, ok := strings.Cut(body, ";\n")
	if !ok || !strings.HasPrefix(head, "const CONFIG = ") {
		t.Fatalf("service worker doesn't start with its CONFIG:\n%s", body)
	}
	w := servedWorker{script: script}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(head, "const CONFIG = ")), &w.cfg); err != nil {
		t.Fatalf("CONFIG: %v", err)
	}
	return w
}

// manifestMembers decodes a manifest's string members named by keys, in
// order.
func manifestMembers(t *testing.T, body []byte, keys ...string) []string {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	got := make([]string, 0, len(keys))
	for _, key := range keys {
		var v string
		if err := json.Unmarshal(raw[key], &v); err != nil {
			t.Errorf("manifest %s: %v", key, err)
		}
		got = append(got, v)
	}
	return got
}

// TestManifest checks the web app manifest: served without a session as
// application/manifest+json under the strict CSP, which admits it, kept
// out of caches like a page, standalone at the root, its colors the design
// tokens', and every icon it names served as a PNG.
func TestManifest(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, manifestPath, ""))
	wantStatus(t, rec, http.StatusOK)
	headers := []string{rec.Header().Get("Content-Type"), rec.Header().Get("Cache-Control"), rec.Header().Get("Content-Security-Policy")}
	if diff := cmp.Diff([]string{"application/manifest+json", "no-store", contentSecurityPolicy}, headers); diff != "" {
		t.Errorf("headers (-want +got):\n%s", diff)
	}
	if !strings.Contains(contentSecurityPolicy, "manifest-src 'self'") {
		t.Errorf("CSP %q doesn't admit the UI's own manifest", contentSecurityPolicy)
	}
	got := manifestMembers(t, rec.Body.Bytes(),
		"id", "name", "short_name", "start_url", "scope", "display", "background_color", "theme_color")
	var m struct {
		Icons []map[string]string `json:"icons"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("manifest icons: %v", err)
	}
	icons := m.Icons
	want := []string{"/", "OpenFief", "OpenFief", "/", "/", "standalone", "#fbfbf8", "#ffffff"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	sizes := make([]string, 0, len(icons))
	for _, icon := range icons {
		sizes = append(sizes, icon["sizes"]+" "+icon["purpose"])
		r := h.serve(t, request(http.MethodGet, icon["src"], ""))
		wantStatus(t, r, http.StatusOK)
		if ct := r.Header().Get("Content-Type"); ct != pngType || icon["type"] != pngType {
			t.Errorf("%s: Content-Type %q, declared %q, want image/png", icon["src"], ct, icon["type"])
		}
	}
	if diff := cmp.Diff([]string{"192x192 any", "512x512 any", "512x512 maskable"}, sizes); diff != "" {
		t.Errorf("icons (-want +got):\n%s", diff)
	}
}

// TestServiceWorker checks the worker's script: served at the root without
// a session as JavaScript, never cached by the browser, with its cache
// named for its version.
func TestServiceWorker(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, workerPath, ""))
	headers := []string{rec.Header().Get("Content-Type"), rec.Header().Get("Cache-Control"), rec.Header().Get("Content-Security-Policy")}
	if diff := cmp.Diff([]string{"text/javascript; charset=utf-8", "no-store", contentSecurityPolicy}, headers); diff != "" {
		t.Errorf("headers (-want +got):\n%s", diff)
	}
	cfg := workerOf(t, h).cfg
	if len(cfg.Version) != versionLen || cfg.Cache != cachePrefix+cfg.Version {
		t.Errorf("version %q, cache %q", cfg.Version, cfg.Cache)
	}
	got := []string{cfg.Offline, cfg.SignOut, strings.Join(cfg.SignIn, " ")}
	want := []string{offlinePath, "/logout", "/login/passkey/finish /login/demo /enroll/finish"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("config (-want +got):\n%s", diff)
	}
}

// TestServiceWorkerCachesOnlyStaticFiles checks the worker's one rule:
// every file it may cache is a static file or the manifest, served without
// a session, sets no cookie and is no page but the static offline page; the
// UI's own scripts, stylesheet, fonts and Kasztan asleep are among them;
// and the script stores nothing but that list.
func TestServiceWorkerCachesOnlyStaticFiles(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	w := workerOf(t, h)
	cfg, script := w.cfg, w.script
	for _, asset := range cfg.Assets {
		if page := underPage(asset); page != "" {
			t.Errorf("worker caches %s under the page or API path %s", asset, page)
		}
		wantStaticAsset(t, h, asset)
	}
	for _, want := range []string{
		"/static/style.css", "/static/layout.js", "/static/app.js", "/static/webauthn.js", "/static/vendor/htmx-2.0.11.min.js",
		"/static/design/tokens.css", "/static/design/fonts/fonts.css", "/static/design/marks/favicon.svg",
		"/static/design/marks/app-icon-192.png", "/static/design/mascot/kasztan-asleep.svg", manifestPath, offlinePath,
	} {
		if !strings.Contains(strings.Join(cfg.Assets, "\n")+"\n", want+"\n") {
			t.Errorf("worker doesn't cache %s", want)
		}
	}
	counts := []int{strings.Count(script, ".put("), strings.Count(script, ".addAll("), strings.Count(script, "CONFIG.assets.map")}
	if diff := cmp.Diff([]int{0, 1, 1}, counts); diff != "" {
		t.Errorf("the script stores something besides CONFIG.assets: .put, .addAll, CONFIG.assets.map counts (-want +got):\n%s", diff)
	}
}

// wantStaticAsset fails t unless asset is a static file or the manifest,
// served without a session, setting no cookie, and no page but the static
// offline page.
func wantStaticAsset(t *testing.T, h *harness, asset string) {
	t.Helper()
	if !strings.HasPrefix(asset, "/static/") && asset != manifestPath {
		t.Errorf("worker caches %s, which is neither a static file nor the manifest", asset)
		return
	}
	rec := h.serve(t, request(http.MethodGet, asset, ""))
	wantStatus(t, rec, http.StatusOK)
	if rec.Header().Get("Set-Cookie") != "" {
		t.Errorf("%s sets a cookie", asset)
	}
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") && asset != offlinePath {
		t.Errorf("worker caches %s, an HTML page", asset)
	}
}

// underPage returns the page or API path asset is or sits under, or "".
func underPage(asset string) string {
	for _, page := range []string{"/", "/login", "/enroll", "/accounts", "/allocation", "/value-chart", "/settings", "/price-checks", "/v1", workerPath} {
		if asset == page || (page != "/" && strings.HasPrefix(asset, page+"/")) {
			return page
		}
	}
	return ""
}

// TestOfflinePage checks the static offline page: no session needed, no
// navigation or session, Kasztan asleep, cached like the static files, and
// the same bytes the worker's version counts.
func TestOfflinePage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, offlinePath, ""))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "kasztan-asleep.svg", "can't be reached", `href="/static/style.css"`)
	wantNoBody(t, rec, `aria-label="Main"`, "access</span>", "Sign out")
	if got := rec.Header().Get("Cache-Control"); got != staticCacheControl {
		t.Errorf("Cache-Control = %q, want %q", got, staticCacheControl)
	}
}

// TestEveryPageIsInstallable checks that every page links the manifest,
// Apple's touch icon and the layout's script, colors the browser's bar
// from the tokens in both schemes, and puts the navigation behind a menu
// button.
func TestEveryPageIsInstallable(t *testing.T) {
	t.Parallel()
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	for name, set := range p.sets {
		src := set.Lookup("layout.html").Tree.Root.String()
		for _, w := range []string{
			`rel="manifest" href="/manifest.webmanifest"`, `rel="apple-touch-icon"`, `<script src="/static/layout.js"></script>`,
			`viewport-fit=cover`, `class="menu-toggle" aria-expanded="false" aria-controls="menu"`,
		} {
			if !strings.Contains(src, w) {
				t.Errorf("page %s lacks %s", name, w)
			}
		}
	}
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, "/login", ""))
	wantBody(t, rec,
		`<meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)">`,
		`<meta name="theme-color" content="#18201c" media="(prefers-color-scheme: dark)">`)
}

// TestReadAppColors checks reading the app's colors from tokens: a token
// that is another token, a dark value falling back to the light one, and a
// missing or non-hex token refused.
func TestReadAppColors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		css     string
		want    appColors
		wantErr error
	}{
		{
			name: "references and dark",
			css: ":root {\n  --white: #FFFFFF;\n  --surface: var(--white);\n  --bg: #fbfbf8;\n}\n" +
				":root[data-theme=\"dark\"] {\n  --surface: #18201c;\n}\n",
			want: appColors{Surface: "#ffffff", SurfaceDark: "#18201c", Background: "#fbfbf8"},
		},
		{
			name: "dark falls back to light",
			css:  ":root {\n  --surface: #ffffff;\n  --bg: #fbfbf8;\n}\n",
			want: appColors{Surface: "#ffffff", SurfaceDark: "#ffffff", Background: "#fbfbf8"},
		},
		{name: "missing", css: ":root {\n  --surface: #ffffff;\n}\n", wantErr: errTokens},
		{name: "not hex", css: ":root {\n  --surface: white;\n  --bg: #fbfbf8;\n}\n", wantErr: errTokens},
		{name: "cycle", css: ":root {\n  --surface: var(--bg);\n  --bg: var(--surface);\n}\n", wantErr: errTokens},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := readAppColors(tc.css)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("colors (-want +got):\n%s", diff)
			}
		})
	}
}

// TestWorkerVersion checks that the version changes with any cached file,
// the manifest or the offline page, and only lists the cached file types.
func TestWorkerVersion(t *testing.T) {
	t.Parallel()
	files := func(css string) fstest.MapFS {
		return fstest.MapFS{"style.css": {Data: []byte(css)}, "notes.txt": {Data: []byte("not cached")}}
	}
	base, err := staticAssets(files("a"), []byte("m"), []byte("o"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(base.urls, " "), "notes.txt") || !strings.Contains(strings.Join(base.urls, " "), "/static/style.css") {
		t.Errorf("assets = %v", base.urls)
	}
	for name, args := range map[string][3]string{"file": {"b", "m", "o"}, "manifest": {"a", "n", "o"}, "offline": {"a", "m", "p"}} {
		v, err := staticAssets(files(args[0]), []byte(args[1]), []byte(args[2]))
		if err != nil {
			t.Fatal(err)
		}
		if v.version == base.version {
			t.Errorf("changing the %s keeps the version %s", name, v.version)
		}
	}
}
