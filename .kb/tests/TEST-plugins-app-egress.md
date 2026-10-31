---
title: The egress use case identifies plugins by token and decides by their egress allow-list
status: active
sourceFile: internal/plugins/app/egress_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:45.274Z'
id: TEST-plugins-app-egress
type: test
verification_scope: unit
---
Table tests build app.Egress over a fake TokenReviewer (synthetic tokens mapped to usernames) and a fake AllowListSource. Identify returns the plugin for a plugin ServiceAccount token, and ErrNoToken, ErrUnauthenticated, the reviewer's own error or ErrNotAPlugin for an empty token, a rejected token, a failing review and an authenticated non-plugin identity; no error contains the token. Decide allows a listed host and port and refuses another port, an unlisted host, an empty list and a plugin without a list, with the matching reason; with no AllowListSource at all it refuses everything rather than falling back to Smokescreen's allow-all. A false pass would be a non-plugin ServiceAccount being accepted; the core-identity case guards it.