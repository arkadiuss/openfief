package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestServiceAccountName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		plugin string
		want   string
	}{
		{name: "connector", plugin: "sample", want: "plugin-sample"},
		{name: "hyphenated", plugin: "nbp-fx", want: "plugin-nbp-fx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, plugins.ServiceAccountName(tt.plugin)); diff != "" {
				t.Errorf("ServiceAccountName(%q) (-want +got):\n%s", tt.plugin, diff)
			}
		})
	}
}

func TestPluginForServiceAccount(t *testing.T) {
	t.Parallel()
	type result struct {
		Plugin string
		OK     bool
	}
	tests := []struct {
		name     string
		username string
		want     result
	}{
		{name: "plugin account", username: "system:serviceaccount:openfief:plugin-sample", want: result{"sample", true}},
		{name: "round trip", username: "system:serviceaccount:openfief:" + plugins.ServiceAccountName("nbp-fx"), want: result{"nbp-fx", true}},
		{name: "other namespace", username: "system:serviceaccount:default:plugin-sample"},
		{name: "namespace prefix trick", username: "system:serviceaccount:openfief-x:plugin-sample"},
		{name: "not a plugin account", username: "system:serviceaccount:openfief:operator"},
		{name: "empty plugin name", username: "system:serviceaccount:openfief:plugin-"},
		{name: "invalid plugin name", username: "system:serviceaccount:openfief:plugin-Sample"},
		{name: "extra segment", username: "system:serviceaccount:openfief:plugin-sample:x"},
		{name: "a user", username: "owner@example.com"},
		{name: "empty", username: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			name, ok := plugins.PluginForServiceAccount(tt.username)
			if diff := cmp.Diff(tt.want, result{name, ok}); diff != "" {
				t.Errorf("PluginForServiceAccount(%q) (-want +got):\n%s", tt.username, diff)
			}
		})
	}
}
