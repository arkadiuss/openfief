package plugins_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestRefuseSecondBrowserView(t *testing.T) {
	t.Parallel()
	ui := plugins.BrowserView{Instance: "ui", Path: "/"}
	tests := []struct {
		name     string
		instance string
		path     string
		held     []plugins.BrowserView
		wantErr  error
	}{
		{name: "the first view", instance: "ui", path: "/"},
		{name: "a plugin with no browser path", instance: "ibkr", held: []plugins.BrowserView{ui}},
		{name: "a second view", instance: "gallery", path: "/gallery", held: []plugins.BrowserView{ui}, wantErr: plugins.ErrBrowserPathHeld},
		{name: "a second view on the same path", instance: "ui-2", path: "/", held: []plugins.BrowserView{ui}, wantErr: plugins.ErrBrowserPathHeld},
		{name: "the holder upgraded", instance: "ui", path: "/", held: []plugins.BrowserView{ui}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := plugins.RefuseSecondBrowserView(tt.instance, tt.path, tt.held)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMayHoldBrowserPath(t *testing.T) {
	t.Parallel()
	// Synthetic creation times of view instances.
	earlier := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	later := time.Date(2026, 10, 2, 8, 0, 0, 0, time.UTC)
	ui := plugins.BrowserView{Instance: "ui", Path: "/", Created: earlier}
	gallery := plugins.BrowserView{Instance: "gallery", Path: "/gallery", Created: later}
	twin := plugins.BrowserView{Instance: "twin", Path: "/", Created: earlier}
	tests := []struct {
		name    string
		self    plugins.BrowserView
		holding bool
		held    []plugins.BrowserView
		wantErr string
	}{
		{name: "no other holder", self: gallery, held: []plugins.BrowserView{gallery}},
		{name: "a newcomer yields to a holder", self: ui, held: []plugins.BrowserView{gallery}, wantErr: "gallery holds /gallery"},
		{name: "the holder keeps it", self: ui, holding: true},
		{name: "of two holders the older keeps it", self: ui, holding: true, held: []plugins.BrowserView{gallery}},
		{name: "of two holders the newer yields", self: gallery, holding: true, held: []plugins.BrowserView{ui}, wantErr: "ui holds /"},
		{name: "created together, the first name keeps it", self: ui, holding: true, held: []plugins.BrowserView{twin}, wantErr: "twin holds /"},
		{name: "the refusal names the senior holder", self: plugins.BrowserView{Instance: "new"}, held: []plugins.BrowserView{gallery, ui}, wantErr: "ui holds /"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := plugins.MayHoldBrowserPath(tt.self, tt.holding, tt.held)
			got := ""
			if err != nil {
				got = err.Error()
			}
			if want := tt.wantErr; (want == "") != (err == nil) || (want != "" && !strings.HasSuffix(got, want)) {
				t.Errorf("MayHoldBrowserPath = %v, want an error ending %q", err, want)
			}
			if err != nil && !errors.Is(err, plugins.ErrBrowserPathHeld) {
				t.Errorf("error %v does not wrap ErrBrowserPathHeld", err)
			}
		})
	}
}
