package operator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()
	schema := `{"type":"object","properties":{"interval":{"type":"string"}},"required":["interval"],"additionalProperties":false}`
	tests := []struct {
		name   string
		schema string
		config string
		want   error
	}{
		{name: "matching config", schema: schema, config: `{"interval":"24h"}`},
		{name: "no schema accepts anything", config: `{"anything":[1,2]}`},
		{name: "no schema, no config"},
		{name: "wrong type", schema: schema, config: `{"interval":24}`, want: operator.ErrConfigRejected},
		{name: "unknown field", schema: schema, config: `{"interval":"24h","extra":true}`, want: operator.ErrConfigRejected},
		{name: "missing config is an empty object", schema: schema, want: operator.ErrConfigRejected},
		{name: "config that is not JSON", schema: schema, config: `{`, want: operator.ErrConfigRejected},
		{name: "schema that is not JSON", schema: `{`, config: `{}`, want: operator.ErrInvalidConfigSchema},
		{name: "schema that is not a schema", schema: `{"type":12}`, config: `{}`, want: operator.ErrInvalidConfigSchema},
		{
			name: "schema that loads a remote document", schema: `{"$ref":"https://schemas.example.com/config.json"}`,
			config: `{}`, want: operator.ErrInvalidConfigSchema,
		},
		{
			name: "schema that loads a local file", schema: `{"$ref":"file:///etc/hostname"}`,
			config: `{}`, want: operator.ErrInvalidConfigSchema,
		},
		{
			name:   "local reference inside the schema",
			schema: `{"$defs":{"d":{"type":"string"}},"type":"object","properties":{"interval":{"$ref":"#/$defs/d"}}}`,
			config: `{"interval":"24h"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := operator.ValidateConfig([]byte(tt.schema), []byte(tt.config))
			if diff := cmp.Diff(tt.want, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ValidateConfig error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
