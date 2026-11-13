// Package authz embeds OpenFief's authorization model (docs/DESIGN.md D-164),
// the SpiceDB schema, for the core to write to the SpiceDB it is configured
// with: authz/schema.zed stays the one source, which `task authz:validate`
// checks against its assertions, and nothing is copied out of it.
package authz

import _ "embed"

// Schema is authz/schema.zed, the model SpiceDB holds relationships over.
//
//go:embed schema.zed
var Schema string
