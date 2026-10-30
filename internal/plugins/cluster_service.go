package plugins

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Errors of in-cluster services (D-74).
var (
	// ErrInvalidClusterService reports an in-cluster service that is not
	// "namespace/service:port" with a DNS-label namespace other than
	// openfief, a DNS-label service name and a port from 1 to 65535.
	ErrInvalidClusterService = errors.New("invalid in-cluster service: want namespace/service:port outside the openfief namespace")
	// ErrClusterServiceNotRequested reports a Plugin that names an
	// in-cluster service while its manifest asks for none.
	ErrClusterServiceNotRequested = errors.New("the manifest asks for no in-cluster service")
)

// maxClusterServiceDescriptionLen bounds a manifest's description of the
// in-cluster service it asks for.
const maxClusterServiceDescriptionLen = 200

// ClusterServiceRequest is a manifest's request for one in-cluster service
// that the owner picks at install, such as a Ghostfolio running in the same
// cluster (D-74). Description says what the service is for; the CLI shows it
// when it asks for the service.
type ClusterServiceRequest struct {
	Description string
}

// validate checks a request's description: at most 200 bytes on one line.
// A nil request is valid.
func (r *ClusterServiceRequest) validate() []error {
	if r == nil {
		return nil
	}
	if len(r.Description) > maxClusterServiceDescriptionLen || strings.ContainsAny(r.Description, "\r\n") {
		return []error{fieldErr("clusterService.description", ErrInvalidName, r.Description)}
	}
	return nil
}

// ClusterService is one in-cluster service a plugin reaches directly, not
// through the egress proxy (D-74): a Service in another namespace than
// openfief, and the Service port the plugin dials.
type ClusterService struct {
	Namespace string
	Service   string
	Port      uint16
}

// ParseClusterService reads "namespace/service:port".
func ParseClusterService(s string) (ClusterService, error) {
	ns, rest, okNS := strings.Cut(s, "/")
	svc, port, okPort := strings.Cut(rest, ":")
	p, err := strconv.ParseUint(port, 10, 16)
	if !okNS || !okPort || err != nil {
		return ClusterService{}, fmt.Errorf("%w: %q", ErrInvalidClusterService, s)
	}
	c := ClusterService{Namespace: ns, Service: svc, Port: uint16(p)}
	if err := c.validate(); err != nil {
		return ClusterService{}, err
	}
	return c, nil
}

// String returns "namespace/service:port".
func (c ClusterService) String() string {
	return c.Namespace + "/" + c.Service + ":" + strconv.FormatUint(uint64(c.Port), 10)
}

// validate checks the namespace and Service names, DNS labels starting with
// a letter, the namespace not openfief, and a non-zero port. A plugin may
// not reach the core or another plugin this way: those live in openfief.
func (c ClusterService) validate() error {
	if !validPluginName(c.Namespace) || c.Namespace == Namespace || !validPluginName(c.Service) || c.Port == 0 {
		return fmt.Errorf("%w: %q", ErrInvalidClusterService, c.String())
	}
	return nil
}

// validateClusterServices checks each service and reports duplicates.
func validateClusterServices(field string, cs []ClusterService) []error {
	var errs []error
	for i, c := range cs {
		if err := c.validate(); err != nil {
			errs = append(errs, &FieldError{Field: indexed(field, i), Err: err})
		}
	}
	return append(errs, duplicates(field, cs, ClusterService.String)...)
}

// validateClusterService checks the Plugin's in-cluster service: valid, and
// named only when its manifest asks for one.
func (p *Plugin) validateClusterService() []error {
	if p.ClusterService == nil {
		return nil
	}
	if p.Manifest.ClusterService == nil {
		return []error{fieldErr("clusterService", ErrClusterServiceNotRequested, p.ClusterService.String())}
	}
	if err := p.ClusterService.validate(); err != nil {
		return []error{&FieldError{Field: "clusterService", Err: err}}
	}
	return nil
}
