package plugins_test

import (
	"errors"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic image references; the digests are not of any real image.
const (
	digestA = "sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	digestB = "sha256:0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"
	imageA  = "registry.example.com/openfief/sample-connector@" + digestA
	imageB  = "registry.example.com/openfief/sample-connector@" + digestB
)

// host builds an EgressHost.
func host(h string, port uint16) plugins.EgressHost { return plugins.EgressHost{Host: h, Port: port} }

// samplePlugin is a valid connector plugin that requests two egress hosts,
// two Secrets and one data tier.
func samplePlugin() plugins.Plugin {
	return plugins.Plugin{
		Name:  "sample",
		Image: imageA,
		Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			Egress:       []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
			Secrets: []plugins.SecretMount{
				{Name: "sample-token", MountPath: "/etc/openfief/token"},
				{Name: "sample-cert", MountPath: "/etc/openfief/cert"},
			},
			DataTiers:    []plugins.DataTier{"T1"},
			ConfigSchema: []byte(`{"type":"object"}`),
		},
		Config: []byte(`{"interval":"24h"}`),
	}
}

// fullGrant approves everything samplePlugin requests.
func fullGrant() plugins.Grant {
	return plugins.Grant{Plugin: "sample", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{host("data.example.org", 8443), host("api.example.com", 443)},
		Secrets:      []string{"sample-cert", "sample-token"},
		DataTiers:    []plugins.DataTier{"T1"},
	}}
}

// fieldError is a FieldError's field and its sentinel, for comparison.
type fieldError struct {
	Field string
	Err   error
}

// fieldErrors flattens a validation error into its field errors, in order.
func fieldErrors(err error) []fieldError {
	if err == nil {
		return nil
	}
	errs := []error{err}
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		errs = joined.Unwrap()
	}
	out := make([]fieldError, 0, len(errs))
	for _, e := range errs {
		if fe, ok := errors.AsType[*plugins.FieldError](e); ok {
			out = append(out, fieldError{Field: fe.Field, Err: fe.Err})
		} else {
			out = append(out, fieldError{Err: e})
		}
	}
	return out
}
