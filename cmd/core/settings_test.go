package main

import (
	"context"
	"net/http/httptest"
	"slices"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// settingsAPI is a generated SettingsService client.
type settingsAPI struct {
	valuationv1connect.SettingsServiceClient
}

// settingsClient returns a SettingsService client of srv that presents
// token as the view plugin's.
func settingsClient(srv *httptest.Server, token string) settingsAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return settingsAPI{valuationv1connect.NewSettingsServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// TestSettingsService checks that the core serves the owner's settings to
// view plugins (D-78, D-79, D-113): PLN and the Polish default format
// space-comma until the owner chooses another, the wrappers, no marginal
// rate until set; an unoffered format, a rate of 32 and other callers
// refused.
func TestSettingsService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	ctx := t.Context()
	c := settingsClient(oc.srv, uiCaller)
	get := func() *valuationv1.GetOwnerSettingsResponse {
		t.Helper()
		got, err := c.GetOwnerSettings(ctx, connect.NewRequest(&valuationv1.GetOwnerSettingsRequest{Origin: localUI, SessionToken: oc.session}))
		if err != nil {
			t.Fatalf("GetOwnerSettings: %v", err)
		}
		return got.Msg
	}
	first := get()
	summary := []any{
		first.GetReportingCurrency(), first.GetNumberFormat().GetKey(), first.GetNumberFormatDefault(),
		first.GetMarginalRate(), len(first.GetTaxWrappers()), len(first.GetNumberFormats()),
	}
	if !slices.Equal(summary, []any{"PLN", "space-comma", true, "", 4, 7}) {
		t.Fatalf("the settings at first = %v", first)
	}
	if _, err := c.SetNumberFormat(ctx, connect.NewRequest(&valuationv1.SetNumberFormatRequest{Origin: localUI, SessionToken: oc.session, Key: "comma-dot"})); err != nil {
		t.Fatalf("SetNumberFormat: %v", err)
	}
	if _, err := c.SetMarginalRate(ctx, connect.NewRequest(&valuationv1.SetMarginalRateRequest{Origin: localUI, SessionToken: oc.session, Rate: "0.32"})); err != nil {
		t.Fatalf("SetMarginalRate: %v", err)
	}
	if got := get(); got.GetNumberFormat().GetKey() != "comma-dot" || got.GetNumberFormatDefault() || got.GetMarginalRate() != "0.32" {
		t.Errorf("the settings after the changes = %v", got)
	}
	checkSettingsRefusals(t, oc.srv, oc.session)
}

// checkSettingsRefusals checks that the core refuses an unoffered number
// format, a marginal rate of 32 and callers other than the owner's view.
func checkSettingsRefusals(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	ctx := t.Context()
	refusals := []struct {
		name, caller, session, key string
		want                       connect.Code
	}{
		{name: "a format not offered", caller: uiCaller, session: session, key: "apostrophe-comma", want: connect.CodeInvalidArgument},
		{name: "not a view plugin", caller: "wrong", session: session, key: "space-dot", want: connect.CodePermissionDenied},
		{name: "an unknown session", caller: uiCaller, session: "ofsess_nope_nope", key: "space-dot", want: connect.CodeUnauthenticated},
	}
	for _, r := range refusals {
		_, err := settingsClient(srv, r.caller).SetNumberFormat(ctx, connect.NewRequest(&valuationv1.SetNumberFormatRequest{
			Origin: localUI, SessionToken: r.session, Key: r.key,
		}))
		if connect.CodeOf(err) != r.want {
			t.Errorf("%s: SetNumberFormat = %v, want %s", r.name, err, r.want)
		}
	}
	_, err := settingsClient(srv, uiCaller).SetMarginalRate(ctx, connect.NewRequest(&valuationv1.SetMarginalRateRequest{Origin: localUI, SessionToken: session, Rate: "32"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a marginal rate of 32: %v, want invalid argument", err)
	}
}
