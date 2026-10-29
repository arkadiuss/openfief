---
title: Over the core's own HTTP server, browser paths reach the view plugin and the core keeps its own paths
status: active
sourceFile: cmd/core/frontdoor_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:46.157Z'
id: TEST-core-front-door
type: test
---
Under task test:pg, starts the core in a fake cluster whose Plugin ui is a ready view with browser path / and whose Service points at an httptest UI, with the TokenRequest API issuing a synthetic core token: GET /login through the core reaches the UI with that token, the browser's Host and http, while /v1, the AuthService and /healthz are answered by the core and never reach the UI; outside a cluster /login is the core's 404.