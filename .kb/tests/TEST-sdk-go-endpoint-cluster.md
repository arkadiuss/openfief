---
title: With an in-cluster service, the SDK's endpoint client calls it directly over verified HTTPS
status: active
sourceFile: sdk/go/openfief/cluster_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T23:25:45.167Z'
id: TEST-sdk-go-endpoint-cluster
type: test
---
TestEndpointClusterClient starts a TLS server on 127.0.0.1 standing in for an in-cluster service, with a certificate for an in-cluster DNS name signed by a CA made for the test, and gives its address to an Endpoint as ClusterService. ClusterClient must reach it with no proxy and the base URL host in the Host header when the bundle holds that CA, fail as an unknown authority without the bundle, fail as a hostname mismatch for a base URL naming another host, refuse a request outside the base URL with ErrOutsideEndpoint, and refuse an address without a port with ErrClusterService.