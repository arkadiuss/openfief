---
title: The SDK's endpoint takes only https base URLs, adds an extra CA bundle to the system roots and stays within its base URL
status: active
sourceFile: sdk/go/openfief/endpoint_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T21:54:59.703Z'
id: TEST-sdk-go-endpoint
type: test
---
TestBaseURL tables configured and default base URLs: an https URL with optional port and path is kept without its trailing slash, the default applies when nothing is configured, and plain http, other schemes, a user, a query, a fragment and bad ports fail with ErrBaseURL. TestNewEndpointBundle and TestReadCABundle check that a missing or empty bundle adds nothing and that a bundle with no valid PEM certificate fails with ErrCABundle. TestEndpointTrustsExtraCA starts real TLS servers signed by two CAs made for the run, one standing in for the system roots, and calls them through a fake egress proxy: the bundle's CA is trusted, the system roots still apply with and without a bundle, and a CA in neither is refused as an unknown authority. TestEndpointRefusesRedirectElsewhere and TestEndpointWithin check that requests and redirects outside the base URL fail with ErrOutsideEndpoint, and TestEndpointRoots that the system pool is copied, not changed.