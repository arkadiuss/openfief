package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    operator.Settings
		wantErr error
	}{
		{
			name: "defaults name the openfief Services",
			want: operator.Settings{
				CoreService: "openfief-core", CorePort: "connect",
				EgressProxyService: "egress-proxy", EgressProxyPort: "proxy", PluginPort: 8080,
			},
		},
		{
			name: "every setting",
			args: []string{
				"--core-service=core", "--core-port=rpc", "--egress-proxy-service=proxy",
				"--egress-proxy-port=connect", "--plugin-port=9090",
			},
			want: operator.Settings{
				CoreService: "core", CorePort: "rpc", EgressProxyService: "proxy", EgressProxyPort: "connect", PluginPort: 9090,
			},
		},
		{name: "plugin port 0", args: []string{"--plugin-port=0"}, wantErr: errPluginPort},
		{name: "plugin port above 65535", args: []string{"--plugin-port=70000"}, wantErr: errPluginPort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o, err := parseFlags(tt.args)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("parseFlags error mismatch (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, o.settings); diff != "" {
				t.Errorf("parseFlags settings mismatch (-want +got):\n%s", diff)
			}
			if !o.leaderElect || o.metricsAddr != "0" {
				t.Errorf("leader election %v, metrics %q; want leader election on and metrics off", o.leaderElect, o.metricsAddr)
			}
		})
	}
}
