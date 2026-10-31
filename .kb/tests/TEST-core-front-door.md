---
title: Over the core's two servers, browsers reach the view plugin on the browser port and plugin-only services only on the plugin port
status: active
sourceFile: cmd/core/frontdoor_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:46.157Z'
id: TEST-core-front-door
type: test
verification_scope: integration
---
Under task test:pg, starts the core in a fake cluster whose Plugin ui is a ready view with browser path / and whose Service points at an httptest UI, with the TokenRequest API issuing a synthetic core token, and serves both of the core's handlers. On the browser port GET /login reaches the UI with that token, the browser's Host and http, /v1 answers 401 and /healthz 200, and one procedure of the AuthService and of each owner service answers 404. On the plugin port the same procedures are mounted, /v1 answers 401, and /login and /healthz are 404. Outside a cluster the browser port answers /login and the plugin-only procedures with 404. A broken split would show a plugin-only procedure on the browser port or the front door on the plugin port.