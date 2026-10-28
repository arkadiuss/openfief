package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// Synthetic tokens; none is a real credential.
const (
	sampleToken = "synthetic-token-sample"
	coreToken   = "synthetic-token-core"
	badToken    = "synthetic-token-rejected"
)

// errReviewUnavailable stands in for a TokenReview API failure.
var errReviewUnavailable = errors.New("token review unavailable")

// fakeReviewer accepts the tokens in its map and returns their usernames.
type fakeReviewer map[string]string

func (f fakeReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	if token == badToken {
		return "", errReviewUnavailable
	}
	username, ok := f[token]
	if !ok {
		return "", fmt.Errorf("%w: rejected by the fake", app.ErrUnauthenticated)
	}
	return username, nil
}

// fakeLists is an AllowListSource over a fixed set of lists.
type fakeLists map[string]plugins.EgressAllowList

func (f fakeLists) AllowList(plugin string) (plugins.EgressAllowList, bool) {
	l, ok := f[plugin]
	return l, ok
}

func newEgress() *app.Egress {
	reviewer := fakeReviewer{
		sampleToken: "system:serviceaccount:openfief:plugin-sample",
		coreToken:   "system:serviceaccount:openfief:openfief-core",
	}
	lists := fakeLists{
		"sample": {Plugin: "sample", Hosts: []plugins.EgressHost{{Host: "api.example.com", Port: 443}}},
		"xtb":    {Plugin: "xtb"},
	}
	return app.NewEgress(reviewer, lists)
}

func TestEgressIdentify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		want    string
		wantErr error
	}{
		{name: "plugin token", token: sampleToken, want: "sample"},
		{name: "no token", token: "", wantErr: app.ErrNoToken},
		{name: "token not accepted", token: "synthetic-token-unknown", wantErr: app.ErrUnauthenticated},
		{name: "review fails", token: badToken, wantErr: errReviewUnavailable},
		{name: "authenticated, not a plugin", token: coreToken, wantErr: app.ErrNotAPlugin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newEgress().Identify(t.Context(), tt.token)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Identify() error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("Identify() = %q, want %q", got, tt.want)
			}
			if err != nil && tt.token != "" && strings.Contains(err.Error(), tt.token) {
				t.Errorf("Identify() error %q contains the token", err)
			}
		})
	}
}

func TestEgressDecide(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		plugin string
		host   string
		port   uint16
		want   app.EgressDecision
	}{
		{
			name: "listed destination", plugin: "sample", host: "api.example.com", port: 443,
			want: app.EgressDecision{Plugin: "sample", Host: "api.example.com", Port: 443, Allowed: true, Reason: app.ReasonListed},
		},
		{
			name: "listed host, other port", plugin: "sample", host: "api.example.com", port: 80,
			want: app.EgressDecision{Plugin: "sample", Host: "api.example.com", Port: 80, Reason: app.ReasonNotListed},
		},
		{
			name: "unlisted host", plugin: "sample", host: "other.example.com", port: 443,
			want: app.EgressDecision{Plugin: "sample", Host: "other.example.com", Port: 443, Reason: app.ReasonNotListed},
		},
		{
			name: "plugin with an empty list", plugin: "xtb", host: "api.example.com", port: 443,
			want: app.EgressDecision{Plugin: "xtb", Host: "api.example.com", Port: 443, Reason: app.ReasonNotListed},
		},
		{
			name: "plugin without a list", plugin: "unknown", host: "api.example.com", port: 443,
			want: app.EgressDecision{Plugin: "unknown", Host: "api.example.com", Port: 443, Reason: app.ReasonNoAllowList},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newEgress().Decide(tt.plugin, tt.host, tt.port)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Decide() (-want +got):\n%s", diff)
			}
		})
	}
}

// TestEgressDecideWithoutSource checks that with no AllowListSource at all,
// nothing is allowed: never Smokescreen's allow-all (D-63).
func TestEgressDecideWithoutSource(t *testing.T) {
	t.Parallel()
	got := app.NewEgress(fakeReviewer{}, nil).Decide("sample", "api.example.com", 443)
	want := app.EgressDecision{Plugin: "sample", Host: "api.example.com", Port: 443, Reason: app.ReasonNoAllowList}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Decide() (-want +got):\n%s", diff)
	}
}
