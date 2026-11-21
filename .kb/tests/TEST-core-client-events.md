---
title: On the real core with its database, the view plugin reads the client events switch with no session and only the owner's session changes it
status: active
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: cmd/core/clientevents_test.go
verification_scope: integration
id: TEST-core-client-events
type: test
---
Starts the real core over a PostgreSQL database and calls its SettingsService with the view plugin's token: a new core answers enabled; a call with no token or another token is refused; SetClientEvents with an unknown session is unauthenticated and changes nothing; the owner's session turns it off and on and each answer follows. The single audit entry of the change is checked by TestEveryOwnerChangeWritesExactlyOneEntry, which drives SetClientEvents as one of its owner changes.