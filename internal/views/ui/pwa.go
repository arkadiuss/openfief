package ui

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/design"
)

// The paths of the installable app's parts. The service worker and the
// manifest sit at the root, so the worker's scope covers every page.
const (
	manifestPath = "/manifest.webmanifest"
	workerPath   = "/sw.js"
	offlinePath  = "/static/offline.html"
	// cachePrefix starts the name of the worker's one cache; the version
	// ends it.
	cachePrefix = "openfief-static-"
	// versionLen is how many hex digits of the assets' hash name a version.
	versionLen = 16
)

var (
	// errTokens reports a design token the app's colors need that
	// design/tokens.css doesn't define as a hex code.
	errTokens = errors.New("ui: design token missing or not a hex color")
	// tokenDecl matches one custom property declaration in tokens.css.
	tokenDecl = regexp.MustCompile(`(--[a-z0-9-]+)\s*:\s*([^;]+);`)
	// tokenRef matches a value that is another token.
	tokenRef = regexp.MustCompile(`^var\((--[a-z0-9-]+)\)$`)
	// hexColor matches a resolved color.
	hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

const (
	// signOutPath is the request that signs the owner out; the worker
	// empties its cache on it.
	signOutPath = "/logout"
	// pngType is the media type of the app icons.
	pngType = "image/png"
)

// signInPaths are the requests that sign the owner in; the worker fills its
// cache after them.
func signInPaths() []string {
	return []string{"/login/passkey/finish", "/login/demo", "/enroll/finish"}
}

// cachedType reports whether the service worker caches a file with the
// extension ext: styles, scripts, fonts and images, nothing else.
func cachedType(ext string) bool {
	switch ext {
	case ".css", ".js", ".woff2", ".svg", ".png":
		return true
	}
	return false
}

// appColors are the app's colors, read from the design tokens: the top
// bar's ground in light and dark for the browser's own bar, and the page's
// ground for the splash screen.
type appColors struct {
	Surface     string
	SurfaceDark string
	Background  string
}

// readAppColors reads the app's colors from tokens.css: the light values
// from its first :root block, the dark ones from :root[data-theme="dark"].
func readAppColors(css string) (appColors, error) {
	light := tokenBlock(css, ":root {")
	dark := tokenBlock(css, `:root[data-theme="dark"] {`)
	for name, value := range light {
		if _, ok := dark[name]; !ok {
			dark[name] = value
		}
	}
	var c appColors
	for _, t := range []struct {
		into   *string
		tokens map[string]string
		name   string
	}{
		{&c.Surface, light, "--surface"}, {&c.SurfaceDark, dark, "--surface"}, {&c.Background, light, "--bg"},
	} {
		v := resolveToken(t.tokens, t.name)
		if !hexColor.MatchString(v) {
			return appColors{}, fmt.Errorf("%w: %s", errTokens, t.name)
		}
		*t.into = strings.ToLower(v)
	}
	return c, nil
}

// tokenBlock returns the custom properties declared in the block of css
// that opens with start, or none when there is no such block.
func tokenBlock(css, start string) map[string]string {
	tokens := map[string]string{}
	_, body, ok := strings.Cut(css, start)
	if !ok {
		return tokens
	}
	body, _, _ = strings.Cut(body, "}")
	for _, m := range tokenDecl.FindAllStringSubmatch(body, -1) {
		tokens[m[1]] = strings.TrimSpace(m[2])
	}
	return tokens
}

// resolveToken follows var() references from name to a value, at most as
// many steps as there are tokens, so a cycle ends.
func resolveToken(tokens map[string]string, name string) string {
	v := tokens[name]
	for range len(tokens) {
		m := tokenRef.FindStringSubmatch(v)
		if m == nil {
			return v
		}
		v = tokens[m[1]]
	}
	return v
}

// appIcon is the path of an app icon the design system generates from the
// Seal (design/marks).
func appIcon(name string) string { return "/static/design/marks/" + name }

// designColors reads the app's colors from the design system's tokens.
func designColors() (appColors, error) {
	tokens, err := fs.ReadFile(design.Files, "tokens.css")
	if err != nil {
		return appColors{}, fmt.Errorf("ui: read tokens: %w", err)
	}
	return readAppColors(string(tokens))
}

// themeColor is the color of the browser's own bar in a color scheme,
// "light" or "dark": the top bar's ground, so the two read as one.
func themeColor(scheme string) (string, error) {
	c, err := designColors()
	if err != nil {
		return "", err
	}
	if scheme == "dark" {
		return c.SurfaceDark, nil
	}
	return c.Surface, nil
}

// manifestIcon is one icon of the web app manifest.
func manifestIcon(name, sizes, purpose string) map[string]string {
	return map[string]string{"src": appIcon(name), "sizes": sizes, "type": pngType, "purpose": purpose}
}

// buildManifest returns the web app manifest that makes the UI installable,
// in the app's colors. A map, since the manifest's member names are snake
// case, which the linter's JSON tag rule refuses.
func buildManifest(c appColors) ([]byte, error) {
	m := map[string]any{
		"id": "/", "name": "OpenFief", "short_name": "OpenFief", "lang": "en",
		"description": "Your portfolio record, on your own server.",
		"start_url":   "/", "scope": "/", "display": "standalone",
		"background_color": c.Background, "theme_color": c.Surface,
		"icons": []map[string]string{
			manifestIcon("app-icon-192.png", "192x192", "any"),
			manifestIcon("app-icon-512.png", "512x512", "any"),
			manifestIcon("app-icon-maskable-512.png", "512x512", "maskable"),
		},
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ui: manifest: %w", err)
	}
	return data, nil
}

// workerConfig is what the service worker script starts with.
type workerConfig struct {
	Version string   `json:"version"`
	Cache   string   `json:"cache"`
	Assets  []string `json:"assets"`
	Offline string   `json:"offline"`
	SignIn  []string `json:"signIn"`
	SignOut string   `json:"signOut"`
}

// pwa holds the installable app's generated files.
type pwa struct {
	manifest []byte
	worker   []byte
	offline  []byte
}

// assetList is what the service worker may cache, and its version.
type assetList struct {
	urls    []string
	version string
}

// staticAssets lists the URL of every file under the two static file
// systems the worker may cache, sorted, and hashes their bytes, the
// manifest's and the offline page's into a version, so any change to them
// makes a new worker.
func staticAssets(static fs.FS, manifest, offline []byte) (assetList, error) {
	sum := sha256.New()
	var assets []string
	add := func(prefix string, files fs.FS) error {
		return fs.WalkDir(files, ".", func(name string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() || !cachedType(path.Ext(name)) {
				return walkErr
			}
			data, readErr := fs.ReadFile(files, name)
			if readErr != nil {
				return fmt.Errorf("ui: read %s: %w", name, readErr)
			}
			assets = append(assets, prefix+name)
			sum.Write([]byte(prefix + name))
			sum.Write(data)
			return nil
		})
	}
	if err := add("/static/", static); err != nil {
		return assetList{}, err
	}
	if err := add("/static/design/", design.Files); err != nil {
		return assetList{}, err
	}
	sum.Write(manifest)
	sum.Write(offline)
	assets = append(assets, manifestPath, offlinePath)
	slices.Sort(assets)
	return assetList{urls: assets, version: hex.EncodeToString(sum.Sum(nil))[:versionLen]}, nil
}

// buildPWA generates the manifest, the offline page and the service worker
// from the design tokens, the static files and the worker's script.
func buildPWA(p *pages, static fs.FS) (*pwa, error) {
	colors, err := designColors()
	if err != nil {
		return nil, err
	}
	manifest, err := buildManifest(colors)
	if err != nil {
		return nil, err
	}
	offline, err := p.execute("offline", "layout.html", page{Title: "Offline"}, plainFormat())
	if err != nil {
		return nil, err
	}
	list, err := staticAssets(static, manifest, offline)
	if err != nil {
		return nil, err
	}
	script, err := webFS.ReadFile("web/pwa/sw.js")
	if err != nil {
		return nil, fmt.Errorf("ui: read service worker: %w", err)
	}
	cfg, err := json.Marshal(workerConfig{
		Version: list.version, Cache: cachePrefix + list.version, Assets: list.urls, Offline: offlinePath,
		SignIn: signInPaths(), SignOut: signOutPath,
	})
	if err != nil {
		return nil, fmt.Errorf("ui: service worker config: %w", err)
	}
	var worker bytes.Buffer
	worker.WriteString("const CONFIG = ")
	worker.Write(cfg)
	worker.WriteString(";\n")
	worker.Write(script)
	return &pwa{manifest: manifest, worker: worker.Bytes(), offline: offline}, nil
}

// serveBytes answers with body as ctype, keeping the security headers.
func serveBytes(ctype, cacheControl string, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", ctype)
		if cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
		if _, err := w.Write(body); err != nil {
			return // the browser went away
		}
	}
}

// pwaRoutes serves the manifest, the service worker and the offline page,
// all without a session: none holds anything of the owner's. The manifest
// and the worker keep the pages' no-store, so a new version is seen at
// once; the offline page is cached like the static files.
func (a *pwa) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+manifestPath, serveBytes("application/manifest+json", "", a.manifest))
	mux.HandleFunc("GET "+workerPath, serveBytes("text/javascript; charset=utf-8", "", a.worker))
	mux.HandleFunc("GET "+offlinePath, serveBytes("text/html; charset=utf-8", staticCacheControl, a.offline))
}
