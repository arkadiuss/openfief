---
title: 'The core''s use case keeps the owner''s client events switch: on by default, owner only, audited, refused without a store'
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: internal/valuation/app/clientevents_test.go
verification_scope: unit
id: TEST-valuation-client-events-switch
type: test
---
Table-style tests of valuation's Settings with in-memory fakes: ClientEventsEnabled is true by default and with no session; SetClientEventsEnabled changes it for the owner and enters one settings_changed entry naming ui:client_events and nothing else per change; a read session changes nothing and enters nothing; a change the audit log refuses is not made; a failing store's error comes back; settings built without a store say off and an error.