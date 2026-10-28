// Package identity is the identity bounded context: who may use OpenFief.
// In M0 it holds only script tokens (docs/DESIGN.md D-27): scoped, expiring,
// read-only personal tokens that the owner's scripts present to the read
// API. Passkeys (D-26) and the core's check of plugin identities arrive
// later.
//
// The package is pure: the clock and the random bytes of a new secret come
// in as parameters. A secret is never stored or logged; only its SHA-256
// hash is kept, and hashes are compared in constant time (D-29: standard
// primitives only).
package identity
