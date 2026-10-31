---
title: The core's browser proxy forwards to the view plugin with its own token and only the headers it sets
status: active
sourceFile: internal/views/browserproxy/proxy_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:41.853Z'
id: TEST-views-browser-proxy
type: test
verification_scope: integration
---
Runs browserproxy.Proxy against an httptest view plugin, with fake routes and a fake token issuer: a POST with a body, a cookie, a stolen Authorization and forged X-Forwarded-Host, X-Forwarded-For and Forwarded headers reaches the plugin with the core's token, the browser's Host and scheme and none of the forged headers, and the plugin's answer and Set-Cookie come back; no route is 404, two views holding a browser path or unread routes 503, no address, no token or no answer 502, never cached; BrowserScheme takes TLS, then http or https from the route in front, else http.