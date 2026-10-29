package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// viewPlugin is a valid view plugin that serves browsers at path.
func viewPlugin(path string) plugins.Plugin {
	return plugins.Plugin{
		Name:  "sample",
		Image: imageA,
		Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityView},
			BrowserPath:  path,
		},
	}
}

// viewGrant approves the view capability and the browser paths.
func viewGrant(paths ...string) plugins.Grant {
	return plugins.Grant{Plugin: "sample", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
		BrowserPaths: paths,
	}}
}

func TestCoreServesPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{path: "/healthz", want: true},
		{path: "/readyz", want: true},
		{path: "/v1", want: true},
		{path: "/v1/accounts", want: true},
		{path: "/openfief.views.v1.ReadService/ListAccounts", want: true},
		{path: "/openfief.identity.v1.AuthService/WhoAmI", want: true},
		{path: "/"},
		{path: "/login"},
		{path: "/v10"},
		{path: "/healthz2"},
		{path: "/static/openfief.css"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			if got := plugins.CoreServesPath(tt.path); got != tt.want {
				t.Errorf("CoreServesPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestManifestBrowserPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		caps []plugins.Capability
		path string
		want []fieldError
	}{
		{name: "the root", caps: []plugins.Capability{plugins.CapabilityView}, path: "/"},
		{name: "a prefix", caps: []plugins.Capability{plugins.CapabilityView}, path: "/views/ghostfolio"},
		{name: "no browser path", caps: []plugins.Capability{plugins.CapabilityConnector}},
		{
			name: "a connector", caps: []plugins.Capability{plugins.CapabilityConnector}, path: "/",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrBrowserPathNeedsView}},
		},
		{
			name: "the read API", caps: []plugins.Capability{plugins.CapabilityView}, path: "/v1",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
		{
			name: "a core service", caps: []plugins.Capability{plugins.CapabilityView}, path: "/openfief.identity.v1.AuthService",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
		{
			name: "a trailing slash", caps: []plugins.Capability{plugins.CapabilityView}, path: "/ui/",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
		{
			name: "relative", caps: []plugins.Capability{plugins.CapabilityView}, path: "ui",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
		{
			name: "upper case", caps: []plugins.Capability{plugins.CapabilityView}, path: "/UI",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
		{
			name: "dot segments", caps: []plugins.Capability{plugins.CapabilityView}, path: "/ui/../v1",
			want: []fieldError{{Field: "browserPath", Err: plugins.ErrInvalidBrowserPath}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := plugins.Manifest{Capabilities: tt.caps, BrowserPath: tt.path}
			got := fieldErrors(m.Validate())
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBrowserPathApproval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		plugin      plugins.Plugin
		grant       plugins.Grant
		wantPending []string
		wantPath    string
	}{
		{name: "approved", plugin: viewPlugin("/"), grant: viewGrant("/"), wantPath: "/"},
		{name: "not approved", plugin: viewPlugin("/"), grant: viewGrant(), wantPending: []string{"/"}},
		{name: "another path approved", plugin: viewPlugin("/"), grant: viewGrant("/ui"), wantPending: []string{"/"}},
		{name: "none requested", plugin: viewPlugin(""), grant: viewGrant("/")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pending, err := plugins.NeedsApproval(&tt.grant, &tt.plugin)
			if err != nil {
				t.Fatalf("NeedsApproval: %v", err)
			}
			if diff := cmp.Diff(tt.wantPending, pending.BrowserPaths); diff != "" {
				t.Errorf("pending browser paths (-want +got):\n%s", diff)
			}
			e, err := plugins.Effective(&tt.plugin, &tt.grant)
			if err != nil {
				t.Fatalf("Effective: %v", err)
			}
			if got := e.BrowserPath(); got != tt.wantPath {
				t.Errorf("BrowserPath = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestBrowserPathNeedsEffectiveView(t *testing.T) {
	t.Parallel()
	p := viewPlugin("/")
	p.Manifest.Capabilities = append(p.Manifest.Capabilities, plugins.CapabilityConnector)
	g := viewGrant("/")
	g.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityConnector}
	e, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	if got := e.BrowserPath(); got != "" {
		t.Errorf("BrowserPath without the view capability granted = %q, want none", got)
	}
	other, err := plugins.Effective(&p, &plugins.Grant{Plugin: "sample", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector, plugins.CapabilityView}, BrowserPaths: []string{"/"},
	}})
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	if e.Equal(other) {
		t.Error("effective permissions that differ in the browser path are equal")
	}
}

func TestGrantBrowserPaths(t *testing.T) {
	t.Parallel()
	g := viewGrant("/", "/v1", "/")
	got := fieldErrors(g.Validate())
	want := []fieldError{
		{Field: "permissions.browserPaths[1]", Err: plugins.ErrInvalidBrowserPath},
		{Field: "permissions.browserPaths[2]", Err: plugins.ErrDuplicate},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Validate (-want +got):\n%s", diff)
	}
}

func TestManifestLabelBrowserPath(t *testing.T) {
	t.Parallel()
	m, err := plugins.ParseManifestLabel(`{"capabilities":["view"],"browserPath":"/"}`)
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	want := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}, BrowserPath: "/"}
	if !m.Equal(&want) {
		t.Errorf("manifest = %+v, want %+v", m, want)
	}
	other := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}}
	if m.Equal(&other) {
		t.Error("manifests that differ in the browser path are equal")
	}
	if _, err := plugins.ParseManifestLabel(`{"capabilities":["connector"],"browserPath":"/"}`); !errors.Is(err, plugins.ErrBrowserPathNeedsView) {
		t.Errorf("a connector's browser path = %v, want ErrBrowserPathNeedsView", err)
	}
}
