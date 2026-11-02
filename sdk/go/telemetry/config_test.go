package telemetry

import (
	"errors"
	"strings"
	"testing"
)

// env returns a getenv over vars.
func env(vars map[string]string) func(string) string {
	return func(name string) string { return vars[name] }
}

func TestParseConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vars map[string]string
		want Config
	}{
		{
			name: "off by default",
			want: Config{ServiceName: "openfief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "a base endpoint over http/protobuf gets the traces path",
			vars: map[string]string{EnvEndpoint: "http://collector.observability:4318"},
			want: Config{Endpoint: "http://collector.observability:4318/v1/traces", ServiceName: "openfief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "a traces endpoint is used as it is and wins",
			vars: map[string]string{EnvEndpoint: "http://ignored.example.org:4318", EnvTracesEndpoint: "https://otlp.example.org/custom/traces"},
			want: Config{Endpoint: "https://otlp.example.org/custom/traces", ServiceName: "openfief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "grpc keeps the endpoint without a path",
			vars: map[string]string{EnvEndpoint: "http://collector.observability:4317", EnvProtocol: "grpc"},
			want: Config{Endpoint: "http://collector.observability:4317", ServiceName: "openfief-core", Protocol: ProtocolGRPC, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "the traces protocol wins over the general one",
			vars: map[string]string{EnvEndpoint: "http://collector:4317", EnvProtocol: "http/protobuf", EnvTracesProtocol: "grpc"},
			want: Config{Endpoint: "http://collector:4317", ServiceName: "openfief-core", Protocol: ProtocolGRPC, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "the service name and a ratio sampler",
			vars: map[string]string{
				EnvEndpoint: "http://collector:4318", EnvServiceName: "fief-core", EnvSampler: "parentbased_traceidratio", EnvSamplerArg: "0.25",
				EnvResourceAttributes: "deployment.environment.name=home",
			},
			want: Config{Endpoint: "http://collector:4318/v1/traces", ServiceName: "fief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_traceidratio", SamplerArg: 0.25},
		},
		{
			name: "OTEL_SDK_DISABLED turns an endpoint off",
			vars: map[string]string{EnvEndpoint: "http://collector:4318", EnvSDKDisabled: "true"},
			want: Config{ServiceName: "openfief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
		{
			name: "the none exporter turns an endpoint off",
			vars: map[string]string{EnvEndpoint: "http://collector:4318", EnvTracesExporter: "none"},
			want: Config{ServiceName: "openfief-core", Protocol: ProtocolHTTP, Sampler: "parentbased_always_on", SamplerArg: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseConfig(env(tt.vars), "openfief-core")
			if err != nil {
				t.Fatalf("ParseConfig: %v", err)
			}
			if got.Endpoint != tt.want.Endpoint || got.Protocol != tt.want.Protocol || got.ServiceName != tt.want.ServiceName ||
				got.Sampler != tt.want.Sampler || got.SamplerArg != tt.want.SamplerArg {
				t.Errorf("ParseConfig = %+v, want %+v", got, tt.want)
			}
			if got.Enabled() != (tt.want.Endpoint != "") {
				t.Errorf("Enabled() = %v with endpoint %q", got.Enabled(), got.Endpoint)
			}
		})
	}
}

func TestParseConfigHeaders(t *testing.T) {
	t.Parallel()
	got, err := ParseConfig(env(map[string]string{
		EnvEndpoint: "https://otlp.example.org", EnvHeaders: "x-api-key=synthetic%3Dkey, x-tenant=home",
	}), "openfief-core")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if len(got.headers) != 2 || got.headers["x-api-key"] != "synthetic=key" || got.headers["x-tenant"] != "home" {
		t.Errorf("headers = %v, want x-api-key and x-tenant decoded", got.headers)
	}
}

func TestParseConfigRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vars map[string]string
		want error
	}{
		{"an endpoint without a scheme", map[string]string{EnvEndpoint: "collector:4318"}, ErrBadEndpoint},
		{"an endpoint of another scheme", map[string]string{EnvEndpoint: "ftp://collector"}, ErrBadEndpoint},
		{"http/json, which Go can't send", map[string]string{EnvEndpoint: "http://c:4318", EnvProtocol: "http/json"}, ErrBadProtocol},
		{"headers that are not pairs", map[string]string{EnvEndpoint: "http://c:4318", EnvHeaders: "secret-without-key"}, ErrBadHeaders},
		{"an unknown sampler", map[string]string{EnvSampler: "sometimes"}, ErrBadSampler},
		{"a ratio above 1", map[string]string{EnvSampler: "traceidratio", EnvSamplerArg: "1.5"}, ErrBadSampler},
		{"a ratio that is not a number", map[string]string{EnvSampler: "traceidratio", EnvSamplerArg: "half"}, ErrBadSampler},
		{"a console exporter", map[string]string{EnvTracesExporter: "console"}, ErrBadExporter},
		{"a disabled flag that is not a bool", map[string]string{EnvSDKDisabled: "yes"}, ErrBadBool},
		{"resource attributes that are not pairs", map[string]string{EnvResourceAttributes: "home"}, ErrBadResource},
		{"a bad value while export is off", map[string]string{EnvTracesExporter: "none", EnvProtocol: "carrier-pigeon"}, ErrBadProtocol},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseConfig(env(tt.vars), "openfief-core")
			if !errors.Is(err, tt.want) {
				t.Errorf("ParseConfig error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseConfigNeverQuotesHeaders(t *testing.T) {
	t.Parallel()
	_, err := ParseConfig(env(map[string]string{EnvEndpoint: "http://c:4318", EnvHeaders: "x-api-key=synthetic-secret,broken"}), "openfief-core")
	if err == nil {
		t.Fatal("ParseConfig accepted a header without a value")
	}
	if got := err.Error(); containsAny(got, "synthetic-secret", "broken") {
		t.Errorf("error %q quotes the headers", got)
	}
}

func TestConfigSampler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		sampler string
		arg     float64
		want    string
	}{
		{"always_on", 1, "AlwaysOnSampler"},
		{"always_off", 1, "AlwaysOffSampler"},
		{"traceidratio", 0.25, "TraceIDRatioBased{0.25}"},
		{"parentbased_always_on", 1, "ParentBased{root:AlwaysOnSampler,"},
		{"parentbased_always_off", 1, "ParentBased{root:AlwaysOffSampler,"},
		{"parentbased_traceidratio", 0.5, "ParentBased{root:TraceIDRatioBased{0.5},"},
	}
	for _, tt := range tests {
		t.Run(tt.sampler, func(t *testing.T) {
			t.Parallel()
			s, ok := samplers(tt.arg)[tt.sampler]
			if !ok {
				t.Fatalf("no sampler %s", tt.sampler)
			}
			if got := s.Description(); !strings.HasPrefix(got, tt.want) {
				t.Errorf("sampler %s = %q, want it to start %q", tt.sampler, got, tt.want)
			}
		})
	}
}
