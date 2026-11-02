package telemetry

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// The standard OpenTelemetry environment variables Config reads. A
// traces-specific variable wins over its general one.
const (
	// EnvEndpoint is the OTLP collector's base URL. Without it, or
	// EnvTracesEndpoint, nothing is exported.
	EnvEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	// EnvTracesEndpoint is the full URL spans are sent to.
	EnvTracesEndpoint = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	// EnvProtocol is the OTLP protocol: grpc or http/protobuf, the default.
	EnvProtocol = "OTEL_EXPORTER_OTLP_PROTOCOL"
	// EnvTracesProtocol is EnvProtocol for traces.
	EnvTracesProtocol = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"
	// EnvHeaders holds headers sent with every export, such as a
	// collector's API key, as comma-separated key=value pairs.
	EnvHeaders = "OTEL_EXPORTER_OTLP_HEADERS"
	// EnvTracesHeaders is EnvHeaders for traces.
	EnvTracesHeaders = "OTEL_EXPORTER_OTLP_TRACES_HEADERS"
	// EnvSampler names the sampler; parentbased_always_on by default.
	EnvSampler = "OTEL_TRACES_SAMPLER"
	// EnvSamplerArg is the ratio samplers' probability, 0 to 1.
	EnvSamplerArg = "OTEL_TRACES_SAMPLER_ARG"
	// EnvServiceName names the service in every span's resource.
	EnvServiceName = "OTEL_SERVICE_NAME"
	// EnvResourceAttributes adds resource attributes, as comma-separated
	// key=value pairs.
	EnvResourceAttributes = "OTEL_RESOURCE_ATTRIBUTES"
	// EnvSDKDisabled set to true turns export off whatever else is set.
	EnvSDKDisabled = "OTEL_SDK_DISABLED"
	// EnvTracesExporter names the exporter: otlp, the default, or none.
	EnvTracesExporter = "OTEL_TRACES_EXPORTER"
)

// Protocol is an OTLP transport.
type Protocol string

// The OTLP transports Config accepts.
const (
	// ProtocolHTTP is OTLP over HTTP with protobuf bodies, the
	// specification's default.
	ProtocolHTTP Protocol = "http/protobuf"
	// ProtocolGRPC is OTLP over gRPC.
	ProtocolGRPC Protocol = "grpc"
)

// The samplers EnvSampler may name.
const (
	samplerAlwaysOn        = "always_on"
	samplerAlwaysOff       = "always_off"
	samplerRatio           = "traceidratio"
	samplerParentAlwaysOn  = "parentbased_always_on"
	samplerParentAlwaysOff = "parentbased_always_off"
	samplerParentRatio     = "parentbased_traceidratio"
)

// The exporters EnvTracesExporter may name.
const (
	exporterOTLP = "otlp"
	exporterNone = "none"
)

const (
	// defaultSamplerRatio is the ratio samplers' probability without
	// EnvSamplerArg.
	defaultSamplerRatio = 1.0
	// floatBits is the precision EnvSamplerArg is parsed with.
	floatBits = 64
	// httpTracesPath is appended to a base endpoint for http/protobuf.
	httpTracesPath = "/v1/traces"
)

// Configuration errors. None of them quotes a header's or an endpoint's
// value: either may hold a credential.
var (
	// ErrBadEndpoint reports an endpoint that is not an http or https URL
	// with a host.
	ErrBadEndpoint = errors.New("OTLP endpoint must be an http:// or https:// URL with a host")
	// ErrBadProtocol reports a protocol other than grpc and http/protobuf.
	ErrBadProtocol = errors.New("OTLP protocol must be grpc or http/protobuf")
	// ErrBadHeaders reports headers that are not key=value pairs.
	ErrBadHeaders = errors.New("OTLP headers must be comma-separated key=value pairs")
	// ErrBadSampler reports an unknown sampler or a ratio outside 0 to 1.
	ErrBadSampler = errors.New("bad trace sampler")
	// ErrBadExporter reports a traces exporter other than otlp and none.
	ErrBadExporter = errors.New("traces exporter must be otlp or none")
	// ErrBadBool reports a boolean variable that is neither true nor false.
	ErrBadBool = errors.New("must be true or false")
	// ErrBadResource reports resource attributes that are not key=value
	// pairs.
	ErrBadResource = errors.New("resource attributes must be comma-separated key=value pairs")
	// errNotKeyValue reports a list item without a key and a value.
	errNotKeyValue = errors.New("not key=value")
)

// Config is where and how a process exports its spans. The zero Config
// exports nothing: tracing is off until an OTLP endpoint is configured.
type Config struct {
	// Endpoint is the URL spans are sent to: for http/protobuf the full
	// URL, ending in /v1/traces for a base URL; for grpc the collector's
	// URL. Empty turns export off.
	Endpoint string
	// Protocol is the OTLP transport.
	Protocol Protocol
	// Sampler names the sampler, one of the specification's names, and
	// SamplerArg is the ratio samplers' probability.
	Sampler    string
	SamplerArg float64
	// ServiceName names the service in every span's resource.
	ServiceName string
	// headers are sent with every export; never logged.
	headers map[string]string
}

// Enabled reports whether spans are exported: an endpoint is configured.
func (c *Config) Enabled() bool { return c.Endpoint != "" }

// ParseConfig reads the standard OpenTelemetry variables through getenv,
// such as os.Getenv. Without an endpoint, with OTEL_SDK_DISABLED=true or
// with OTEL_TRACES_EXPORTER=none it returns a Config that exports nothing.
// The service is named defaultService unless OTEL_SERVICE_NAME names it.
// Every value is checked, whether export is on or not, and one it can't
// use is an error, so a typo never silently turns export off or sends
// spans somewhere else.
func ParseConfig(getenv func(string) string, defaultService string) (Config, error) {
	cfg := Config{ServiceName: defaultService, Protocol: ProtocolHTTP, Sampler: samplerParentAlwaysOn, SamplerArg: defaultSamplerRatio}
	if name := strings.TrimSpace(getenv(EnvServiceName)); name != "" {
		cfg.ServiceName = name
	}
	if _, err := keyValues(getenv(EnvResourceAttributes)); err != nil {
		return Config{}, ErrBadResource
	}
	off, err := exportOff(getenv)
	if err != nil {
		return Config{}, err
	}
	if samplerErr := cfg.parseSampler(getenv(EnvSampler), getenv(EnvSamplerArg)); samplerErr != nil {
		return Config{}, samplerErr
	}
	if protocolErr := cfg.parseProtocol(first(getenv, EnvTracesProtocol, EnvProtocol)); protocolErr != nil {
		return Config{}, protocolErr
	}
	if cfg.headers, err = keyValues(first(getenv, EnvTracesHeaders, EnvHeaders)); err != nil {
		return Config{}, ErrBadHeaders
	}
	endpoint, err := cfg.endpoint(getenv(EnvTracesEndpoint), getenv(EnvEndpoint))
	if err != nil {
		return Config{}, err
	}
	if !off {
		cfg.Endpoint = endpoint
	}
	return cfg, nil
}

// exportOff reports whether OTEL_SDK_DISABLED or OTEL_TRACES_EXPORTER turn
// export off.
func exportOff(getenv func(string) string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(getenv(EnvSDKDisabled))) {
	case "true":
		return true, nil
	case "", "false":
	default:
		return false, fmt.Errorf("%s %w", EnvSDKDisabled, ErrBadBool)
	}
	switch strings.TrimSpace(getenv(EnvTracesExporter)) {
	case "", exporterOTLP:
		return false, nil
	case exporterNone:
		return true, nil
	}
	return false, fmt.Errorf("%w: %s", ErrBadExporter, EnvTracesExporter)
}

// first returns the first of names that getenv holds a value for.
func first(getenv func(string) string, names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

// parseProtocol sets the protocol, http/protobuf when raw is empty.
func (c *Config) parseProtocol(raw string) error {
	switch Protocol(raw) {
	case "", ProtocolHTTP:
		c.Protocol = ProtocolHTTP
	case ProtocolGRPC:
		c.Protocol = ProtocolGRPC
	default:
		return ErrBadProtocol
	}
	return nil
}

// endpoint returns the URL spans go to: traces as it is, or base with
// /v1/traces for http/protobuf, or "" when neither is set.
func (c *Config) endpoint(traces, base string) (string, error) {
	traces, base = strings.TrimSpace(traces), strings.TrimSpace(base)
	raw := traces
	if raw == "" {
		raw = base
	}
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", ErrBadEndpoint
	}
	if traces == "" && c.Protocol == ProtocolHTTP {
		u.Path = strings.TrimSuffix(u.Path, "/") + httpTracesPath
	}
	return u.String(), nil
}

// parseSampler sets the sampler and, for the ratio samplers, its ratio.
func (c *Config) parseSampler(name, arg string) error {
	name = strings.TrimSpace(name)
	switch name {
	case "":
		return nil
	case samplerAlwaysOn, samplerAlwaysOff, samplerParentAlwaysOn, samplerParentAlwaysOff:
		c.Sampler = name
		return nil
	case samplerRatio, samplerParentRatio:
		c.Sampler = name
	default:
		return fmt.Errorf("%w: %s names no sampler", ErrBadSampler, EnvSampler)
	}
	if arg = strings.TrimSpace(arg); arg == "" {
		return nil
	}
	ratio, err := strconv.ParseFloat(arg, floatBits)
	if err != nil || ratio < 0 || ratio > 1 {
		return fmt.Errorf("%w: %s must be a number from 0 to 1", ErrBadSampler, EnvSamplerArg)
	}
	c.SamplerArg = ratio
	return nil
}

// samplers returns every sampler EnvSampler may name, the ratio samplers
// with ratio.
func samplers(ratio float64) map[string]sdktrace.Sampler {
	return map[string]sdktrace.Sampler{
		samplerAlwaysOn:        sdktrace.AlwaysSample(),
		samplerAlwaysOff:       sdktrace.NeverSample(),
		samplerRatio:           sdktrace.TraceIDRatioBased(ratio),
		samplerParentAlwaysOn:  sdktrace.ParentBased(sdktrace.AlwaysSample()),
		samplerParentAlwaysOff: sdktrace.ParentBased(sdktrace.NeverSample()),
		samplerParentRatio:     sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)),
	}
}

// keyValues parses comma-separated key=value pairs whose parts may be
// URL-encoded, as the specification writes headers and resource
// attributes.
func keyValues(raw string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	for item := range strings.SplitSeq(raw, ",") {
		rawKey, rawValue, ok := strings.Cut(item, "=")
		if !ok {
			return nil, errNotKeyValue
		}
		key, kerr := url.PathUnescape(strings.TrimSpace(rawKey))
		value, verr := url.PathUnescape(strings.TrimSpace(rawValue))
		if kerr != nil || verr != nil || key == "" {
			return nil, errNotKeyValue
		}
		out[key] = value
	}
	return out, nil
}
