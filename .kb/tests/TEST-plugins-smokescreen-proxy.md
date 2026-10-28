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
Runs the proxy as production does: smokescreen.New wraps Smokescreen's BuildProxy handler in OpenFief's own http.Server on a local listener, and Proxy.Serve runs until a test context ends. A fake TokenReviewer, a swappable allow-list, a fake resolver and a dialer that sends every connection to a local httptest TLS server keep it off the network. Raw CONNECT requests check that a listed host with the token as the Basic password (any or no user name, scheme in any case) tunnels through, while an unlisted host, a host resolving to a private address or to public and private addresses, no token, an unknown token, the token as the user name, Bearer, malformed Basic credentials, an authenticated plugin without an allow-list, an upstream-proxy header and plain HTTP proxying are refused with their status codes. Go's own HTTP client with the token in the proxy URL gets through. With no allow-list source at all nothing is allowed or dialed; a swapped list applies to the next CONNECT; the healthcheck answers; stopping Serve returns nil, closes open tunnels and stops accepting; New needs a listener and the egress use case and refuses HTTPS_PROXY. Only the public address of the allowed host is ever dialed, which guards against a refused request whose tunnel still dialed.