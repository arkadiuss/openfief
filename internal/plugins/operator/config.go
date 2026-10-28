package operator

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Errors of config validation.
var (
	// ErrInvalidConfigSchema reports a manifest config schema that is not a
	// usable JSON Schema.
	ErrInvalidConfigSchema = errors.New("invalid config schema")
	// ErrConfigRejected reports a config its schema rejects.
	ErrConfigRejected = errors.New("config does not match its schema")
	// errExternalRef refuses every $ref outside the schema document.
	errExternalRef = errors.New("config schemas may not load other documents")
)

const schemaURL = "urn:openfief:plugin-config-schema"

// noLoader refuses to load any document, so a config schema can never make
// the operator read a file or reach the network.
type noLoader struct{}

// Load refuses url.
func (noLoader) Load(url string) (any, error) { return nil, fmt.Errorf("%w: %q", errExternalRef, url) }

// ValidateConfig checks a plugin's config against its manifest's JSON Schema.
// Without a schema any config is accepted; a missing config is checked as an
// empty object. Only the schema document itself and the standard
// meta-schemas are available to $ref.
func ValidateConfig(schema, config []byte) error {
	if len(schema) == 0 {
		return nil
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidConfigSchema, err)
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(noLoader{})
	if addErr := c.AddResource(schemaURL, doc); addErr != nil {
		return fmt.Errorf("%w: %w", ErrInvalidConfigSchema, addErr)
	}
	compiled, err := c.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidConfigSchema, err)
	}
	if len(config) == 0 {
		config = []byte("{}")
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(config))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfigRejected, err)
	}
	if err := compiled.Validate(inst); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigRejected, err)
	}
	return nil
}
