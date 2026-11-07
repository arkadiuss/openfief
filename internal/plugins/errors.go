package plugins

import (
	"errors"
	"fmt"
)

// Validation errors. Each is returned wrapped, with the offending value, inside
// a *FieldError naming the field; test for them with errors.Is.
var (
	ErrInvalidHostname       = errors.New("invalid hostname")
	ErrIPLiteral             = errors.New("IP address literal instead of a hostname")
	ErrWildcardHost          = errors.New("wildcard hostname")
	ErrInvalidPort           = errors.New("invalid port")
	ErrInvalidImageRef       = errors.New("invalid image reference")
	ErrImageNotPinned        = errors.New("image not pinned by digest")
	ErrNoCapability          = errors.New("no capability declared")
	ErrUnknownCapability     = errors.New("unknown capability")
	ErrInvalidDataTier       = errors.New("invalid data tier")
	ErrInvalidName           = errors.New("invalid name")
	ErrInvalidMountPath      = errors.New("invalid mount path")
	ErrOverlappingMountPaths = errors.New("overlapping mount paths")
	ErrDuplicate             = errors.New("duplicate entry")
	ErrInvalidJSON           = errors.New("invalid JSON")
	ErrInvalidTopology       = errors.New("invalid cluster topology")
	ErrInvalidBrowserPath    = errors.New("invalid browser path: not / or a clean lower-case path the core does not serve itself")
	ErrBrowserPathNeedsView  = errors.New("only a view plugin may serve browsers")
	// ErrMarketDataNeedsConnector: only a connector can be a market-data
	// connector (D-123).
	ErrMarketDataNeedsConnector = errors.New("only a connector may declare itself a market-data connector")
	// ErrDeliveryNeedsConnector: only a connector delivers bond terms,
	// index levels (D-179) or index readings (D-186).
	ErrDeliveryNeedsConnector = errors.New("only a connector may declare that it delivers bond terms, index levels or index readings")
	// ErrInstitutionsNeedConnector: only a connector books for
	// institutions (D-182).
	ErrInstitutionsNeedConnector = errors.New("only a connector may declare institutions")
	// ErrSecretNameSeparator: an instance name or a Secret role holds
	// SecretNameSeparator, so a Secret's name could read two ways (D-89).
	ErrSecretNameSeparator = errors.New(`name holds "--", which only separates an instance from a Secret role`)
	// ErrExtraCANotRequested: a Plugin expects an extra CA bundle while its
	// manifest doesn't ask for the CA bundle Secret role.
	ErrExtraCANotRequested = errors.New("an extra CA bundle is expected but the manifest doesn't ask for its Secret role")
)

// Errors that are not about a single field.
var (
	// ErrGrantMismatch reports a Grant that names a different plugin.
	ErrGrantMismatch = errors.New("grant is for a different plugin")
	// ErrNoCapabilityGranted reports that no requested capability is granted,
	// so the plugin has nothing it may do and must not be deployed.
	ErrNoCapabilityGranted = errors.New("no requested capability is granted")
	// ErrMissing reports a nil Plugin or Grant.
	ErrMissing = errors.New("missing plugin or grant")
)

// FieldError is one validation failure: the field it concerns, such as
// "manifest.egress[1]", and the cause, which wraps one of the Err* values.
type FieldError struct {
	Field string
	Err   error
}

// Error returns the field and the cause.
func (e *FieldError) Error() string { return e.Field + ": " + e.Err.Error() }

// Unwrap returns the cause, so errors.Is finds the sentinel.
func (e *FieldError) Unwrap() error { return e.Err }

// fieldErr wraps sentinel with the offending value for field.
func fieldErr(field string, sentinel error, value string) *FieldError {
	return &FieldError{Field: field, Err: fmt.Errorf("%w: %q", sentinel, value)}
}

// indexed names element i of a list field, such as "egress[2]".
func indexed(field string, i int) string { return fmt.Sprintf("%s[%d]", field, i) }
