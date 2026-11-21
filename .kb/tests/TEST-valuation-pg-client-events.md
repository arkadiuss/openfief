---
title: 'The store keeps the client events switch as the absence of a setting: on by default, off stored, on again removes the row'
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: internal/valuation/postgres/clientevents_test.go
verification_scope: integration
id: TEST-valuation-pg-client-events
type: test
---
Against PostgreSQL (task test:pg): with nothing stored ClientEventsOff is false; saving true then false and again leaves the value saved last; after saving false no client_events row remains in valuation.settings, so on is the absence of the setting.