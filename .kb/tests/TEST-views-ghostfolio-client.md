---
title: The Ghostfolio client never sends an unfiltered bulk delete and reports errors without response bodies
status: active
sourceFile: internal/views/ghostfolio/client_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:18.297Z'
id: TEST-views-ghostfolio-client
type: test
verification_scope: integration
---
Against an httptest server, DeleteActivitiesInAccounts with no account and Activities with none fail with ErrNoAccountFilter and send no request at all. TestClientErrors checks that deleting an activity Ghostfolio answers 404 for counts as deleted, that a 400 becomes an HTTPError naming the call and status that wraps ErrHTTP and does not contain the response body, and that Login reads the security token once per login.