---
title: The egress transport tunnels through the proxy with the current token as the proxy URL's password
status: active
sourceFile: sdk/go/openfief/egress_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:30.890Z'
id: TEST-sdk-go-egress
type: test
---
Runs a fake HTTP CONNECT proxy in front of a TLS httptest server standing in for an external host. Two GETs through EgressTransport, with the synthetic token file rewritten in between, must produce two CONNECTs to the host whose Proxy-Authorization is Basic with user plugin and the first, then the rotated token as password, and no non-CONNECT request. A plain http URL fails with ErrNotHTTPS, a transport without a proxy with ErrNoEgressProxy and an empty token with ErrEmptyToken, each without reaching the proxy. NewEgressClient needs HTTPS_PROXY and OPENFIEF_EGRESS_TOKEN_FILE and reads the token from that file.