---
title: The PostgreSQL auth store round-trips passkeys, ceremonies, codes and sessions and holds no secret
status: active
sourceFile: internal/identity/postgres/auth_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:40.471Z'
id: TEST-identity-auth-postgres
type: test
---
Under task test:pg on a disposable database: the owner row is created once; passkeys round-trip and record their use; the last passkey can't be deleted; ceremonies are taken once and expired ones are cleaned up; a session round-trips, touches never move back and revocation keeps its first time; the table refuses an owner-scoped demo session; eight concurrent uses of one enrollment code let exactly one through; and an end-to-end enrollment, login and script token over the real store leaves no session or code key in any column.