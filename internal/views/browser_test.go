package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/views"
)

func TestRouteBrowserPath(t *testing.T) {
	t.Parallel()
	ui := views.BrowserRoute{Instance: "ui", Prefix: "/"}
	gallery := views.BrowserRoute{Instance: "gallery", Prefix: "/gallery"}
	tests := []struct {
		name    string
		routes  []views.BrowserRoute
		path    string
		want    views.BrowserRoute
		wantErr error
	}{
		{name: "the root serves every path", routes: []views.BrowserRoute{ui}, path: "/accounts/a-1", want: ui},
		{name: "the root itself", routes: []views.BrowserRoute{ui}, path: "/", want: ui},
		{name: "a path below the prefix", routes: []views.BrowserRoute{gallery}, path: "/gallery/x", want: gallery},
		{name: "a prefix itself", routes: []views.BrowserRoute{gallery}, path: "/gallery", want: gallery},
		{name: "a prefix covers whole segments only", routes: []views.BrowserRoute{gallery}, path: "/gallery2", wantErr: views.ErrNoBrowserRoute},
		{name: "a path outside the prefix", routes: []views.BrowserRoute{gallery}, path: "/login", wantErr: views.ErrNoBrowserRoute},
		{name: "no view holds a browser path", path: "/", wantErr: views.ErrNoBrowserRoute},
		{name: "two views on different prefixes", routes: []views.BrowserRoute{ui, gallery}, path: "/gallery/x", wantErr: views.ErrSeveralBrowserViews},
		{name: "two views, a path only one covers", routes: []views.BrowserRoute{gallery, ui}, path: "/login", wantErr: views.ErrSeveralBrowserViews},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := views.RouteBrowserPath(tt.routes, tt.path)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("route (-want +got):\n%s", diff)
			}
		})
	}
}
