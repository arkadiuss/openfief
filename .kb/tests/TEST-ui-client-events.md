---
title: 'The pages'' client events: vocabulary, caps, limit, log, metrics and the off switch'
status: active
sourceFile: internal/views/ui/clientevents_test.go
verification_scope: integration
origin:
  kind: agent
  ref: D-227, 2026-11-18
  recorded_at: '2026-11-18T21:55:28.582Z'
id: TEST-ui-client-events
type: test
---
Table tests of each field's validation (kinds, times, codes, paths, statuses, scripts, lines and messages), of the User-Agent's family and version, and of the per-address limit and its bound; through the UI's handler: events logged without the address and counted, refusals for a body that is no report or is too large, the cap on events, the limit per address, the same-origin rule, and the off switch, with the page loading no reporting script; and the plugin's config default.