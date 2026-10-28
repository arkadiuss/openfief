---
title: The embedded Smokescreen proxy tunnels only allowed plugins to allowed, public destinations
status: active
sourceFile: internal/plugins/smokescreen/proxy_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:49.149Z'
id: TEST-plugins-smokescreen-proxy
type: test
---
Runs the real Smokescreen server in-process on a local listener with a fake TokenReviewer, a fixed allow-list, a fake resolver and a dialer that sends every connection to a local httptest TLS server, so no real network is used. Raw CONNECT requests check that a listed host with a valid bearer token tunnels through to the TLS server, while an unlisted host, a host resolving to a private address, a host resolving to both public and private addresses, a missing token, an unaccepted token, Basic auth, an upstream-proxy header and plain HTTP proxying are all refused with their status codes. It also checks that only the public address of the allowed host was ever dialed, that NewConfig needs a listener and refuses to start when HTTPS_PROXY is set. A false pass would be a refused request whose tunnel still dialed; the dial log guards it.