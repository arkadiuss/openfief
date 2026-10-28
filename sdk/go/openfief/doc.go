// Package openfief is the Go SDK for OpenFief plugins (D-11), Apache-2.0
// like the contracts (D-5). The generated contract code lives in the gen/
// packages of this module; `task proto:gen` writes it from proto/.
//
// A connector serves the connector contract, openfief.connector.v1, and the
// core calls it (D-49): NewConnectorServer serves an implementation to the
// core only, checking the core's token with a CoreVerifier (D-59), and
// ReceiveFile reads an ImportFile stream. NewEgressClient reaches granted
// external hosts through the egress proxy, with the plugin's egress token as
// the proxy URL's password on every request (D-63). CoreHTTPClient calls the
// core, such as its read API, with the plugin's core token (D-27). Projected
// tokens rotate, so every token is read from its file on every request.
package openfief
