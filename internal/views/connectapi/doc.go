// Package connectapi serves the read API (docs/DESIGN.md D-6, D-50): the
// openfief.views.v1.ReadService defined in proto/openfief/views/v1, as
// Connect (and gRPC, gRPC-Web) for generated clients and, transcoded by
// Vanguard from the same handlers, as REST under /v1. The OpenAPI document
// is generated from the same protobuf definitions.
//
// Every method only reads: data enters only through connectors' ingest
// paths (D-7), so REST answers 405 to anything but GET. Every call needs a
// bearer token, checked by an app.TokenVerifier. The log names requests by
// an internal request ID only: never an amount, a holding, an account or a
// token.
package connectapi
