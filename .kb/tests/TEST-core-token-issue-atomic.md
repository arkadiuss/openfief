---
title: core token issue under SpiceDB waits for it, issues nothing without it, and revokes a token whose grant did not reach it
status: active
sourceFile: cmd/core/tokenissue_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T15:35:06.116Z'
id: TEST-core-token-issue-atomic
type: test
---
Under task test:pg, over a fresh database with a SpiceDB that records its writes and can fail them: with the household grant and a SpiceDB that never answers, token issue returns the wait's error, prints nothing, and leaves no token, no audit entry and no write; with one that answers on the third ask it waits and issues; a token without the grant is issued although SpiceDB does not answer and writes nothing; when SpiceDB answers the ping and then refuses the write, the token is revoked and its grant ended, the audit log reads created, grant created, token revoked, grant revoked, nothing is printed, no grant is in force and the error names the token and wraps ErrTokenRevoked; on the happy path the secret is the first line, the token is active, the grant's relationship is the one write and the log has the two entries. The wait itself is tested for an answer at once, a wait of zero asking once, the last answer in the error and an interrupted command. A false pass would be an issue that stores the token before it has asked SpiceDB, which the empty store after a silent SpiceDB would catch.